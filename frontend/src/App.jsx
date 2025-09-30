// App.js
import React, { useState, useEffect } from "react";
import Login from "./components/Login";
import Register from "./components/Register";
import TodoApp from "./components/TodoApp";
import { setAuthToken } from "./api";

export default function App() {
  const [token, setToken] = useState(() => {
    const saved = localStorage.getItem("token");
    if (saved) setAuthToken(saved);
    return saved;
  });

  useEffect(() => {
    setAuthToken(token);
    if (token) localStorage.setItem("token", token);
    else localStorage.removeItem("token");
  }, [token]);

  const handleLogin = (t) => {
    setAuthToken(t);
    setToken(t);
  };


  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        justifyContent: "center",
        alignItems: "flex-start",
        backgroundColor: "#f5f5f5",
        padding: "3rem 1rem",
        fontFamily: "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
      }}
    >
      {!token ? (
        <div
          style={{
            display: "flex",
            flexDirection: "row",
            gap: "2rem",
            maxWidth: 800,
            width: "100%",
            flexWrap: "wrap",
          }}
        >


          <div
            style={{
              flex: 1,
              minWidth: 280,
              background: "#fff",
              padding: "2rem",
              borderRadius: 12,
              boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
            }}
          >
            <Login onLogin={handleLogin} />
          </div>

          <div
            style={{
              flex: 1,
              minWidth: 280,
              background: "#fff",
              padding: "2rem",
              borderRadius: 12,
              boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
            }}
          >
            <Register onRegister={handleLogin} />
          </div>
        </div>
      ) : (
        <TodoApp token={token} onLogout={() => handleLogin(null)} />
      )}
    </div>
  );
}
