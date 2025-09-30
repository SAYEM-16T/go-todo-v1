// TodoApp.js
import React, { useState, useEffect } from "react";
import { authApi } from "../api";

export default function TodoApp({ token, onLogout }) {
  const [username, setUsername] = useState(""); 




  useEffect(() => {
    const loadUser = async () => {
      if (!token) return;
      try {
        const res = await authApi.get("/me", {
          headers: { Authorization: `Bearer ${token}` },
        });
        setUsername(res.data.username); // set username after login
      } catch (err) {
        console.error("Load user error:", err);
      }
    };

    loadUser();
  }, [token]);

  

  return (
    <div
      style={{
        padding: "2rem",
        fontFamily: "'Segoe UI', Tahoma, Geneva, Verdana, sans-serif",
        backgroundColor: "#f5f5f5",
        minHeight: "100vh",
      }}
    >
      <div
        style={{
          maxWidth: 600,
          margin: "0 auto",
          backgroundColor: "#fff",
          borderRadius: 12,
          padding: "2rem",
          boxShadow: "0 4px 12px rgba(0,0,0,0.1)",
        }}
      >
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <h2 style={{ margin: 0 }}>Welcome, {username || "User"}</h2>
          <button
            onClick={onLogout}
            style={{
              background: "#e74c3c",
              color: "#fff",
              border: "none",
              borderRadius: 6,
              padding: "0.4rem 0.8rem",
              cursor: "pointer",
            }}
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  );
}
