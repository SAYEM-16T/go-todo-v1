package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

//
// ===== Models =====
//

type UserDoc struct {
	ID        string `bson:"_id" json:"id"`
	Username  string `bson:"username" json:"username"`
	Email     string `bson:"email" json:"email"`
	Password  string `bson:"password" json:"-"`
	CreatedAt int64  `bson:"createdAt" json:"createdAt"`
	UpdatedAt int64  `bson:"updatedAt" json:"updatedAt"`
}

//
// ===== Request DTOs =====
//

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type MeResponse struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

//
// ===== Utils =====
//

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func nowUnix() int64 { return time.Now().Unix() }

func hashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}

func checkPassword(hash, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

func generateToken(userID string, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
		"iat":    time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

func parseAndValidateToken(tokenStr string, secret []byte) (jwt.MapClaims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// generate a 16-byte hex ID
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

//
// ===== REST Handlers =====
//

type RestHandler struct {
	DB     *mongo.Collection
	JWTKey []byte
}

func (h *RestHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "all fields required"})
		return
	}

	// Check if email already exists
	count, err := h.DB.CountDocuments(r.Context(), bson.M{"email": strings.ToLower(req.Email)})
	if err != nil {
		log.Printf("Mongo count error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if count > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email already registered"})
		return
	}

	hashed, err := hashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "password hashing failed"})
		return
	}

	user := UserDoc{
		ID:        generateID(),
		Username:  req.Username,
		Email:     strings.ToLower(req.Email),
		Password:  hashed,
		CreatedAt: nowUnix(),
		UpdatedAt: nowUnix(),
	}

	_, err = h.DB.InsertOne(r.Context(), user)
	if err != nil {
		log.Printf("Mongo insert error: %v", err) // log the actual error
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create user"})
		return
	}

	token, err := generateToken(user.ID, h.JWTKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token generation failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *RestHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}

	var user UserDoc
	if err := h.DB.FindOne(r.Context(), bson.M{"email": strings.ToLower(req.Email)}).Decode(&user); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if err := checkPassword(user.Password, req.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, err := generateToken(user.ID, h.JWTKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token":    token,
		"username": user.Username,
	})

}

func (h *RestHandler) Me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing token"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := parseAndValidateToken(token, h.JWTKey)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		return
	}

	userID, _ := claims["userId"].(string)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token payload"})
		return
	}

	var user UserDoc
	if err := h.DB.FindOne(r.Context(), bson.M{"_id": userID}).Decode(&user); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, MeResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

func (h *RestHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token required"})
		return
	}

	if _, err := parseAndValidateToken(req.Token, h.JWTKey); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]bool{"valid": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"valid": true})
}

//
// ===== Router & Server =====
//

func newRouter(db *mongo.Collection, jwtKey []byte) http.Handler {
	h := &RestHandler{DB: db, JWTKey: jwtKey}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Group(func(r chi.Router) {
		r.Post("/api/v1/auth/register", h.Register)
		r.Post("/api/v1/auth/login", h.Login)
		r.Get("/api/v1/auth/me", h.Me)
		r.Post("/api/v1/auth/validate", h.ValidateToken)
	})

	return r
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("REST: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	_ = godotenv.Load() // ignore error if no .env

	restPort := getenvDefault("REST_PORT", ":10000")
	mongoURI := getenvDefault("MONGO_URI", "mongodb://localhost:27017")
	dbName := getenvDefault("DB_NAME", "todoapp")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	// Connect MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}
	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatalf("mongo ping error: %v", err)
	}
	db := client.Database(dbName).Collection("users")
	log.Printf("Connected to MongoDB: %s", mongoURI)

	// Start server
	router := newRouter(db, []byte(jwtSecret))
	log.Printf("Auth REST server listening on %s", restPort)
	if err := http.ListenAndServe(restPort, loggingMiddleware(router)); err != nil {
		log.Fatalf("REST serve error: %v", err)
	}
}
