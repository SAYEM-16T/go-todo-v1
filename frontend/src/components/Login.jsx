// Login.js
import React, { useState } from "react";
import { authApi } from "../api";

export default function Login({ onLogin }) {
  const [emailOrUsername, setEmailOrUsername] = useState("");
  const [password, setP] = useState("");
  const [err, setErr] = useState("");

  const submit = async (e) => {
    e.preventDefault();
    setErr("");
    try {
      const res = await authApi.post("/login", { email: emailOrUsername, password });
      onLogin(res.data.token);
    } catch (err) {
      const msg =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "Login error";
      setErr(msg);
    }
  };

  return (
    <form onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
      <h3 style={{ marginBottom: "1rem", textAlign: "center", color: "#333" }}>Login</h3>
      <input
        placeholder="Email or Username"
        value={emailOrUsername}
        onChange={(e) => setEmailOrUsername(e.target.value)}
        style={{ padding: "0.5rem", borderRadius: 6, border: "1px solid #ccc" }}
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setP(e.target.value)}
        style={{ padding: "0.5rem", borderRadius: 6, border: "1px solid #ccc" }}
      />
      <button
        type="submit"
        style={{
          background: "#2ecc71",
          color: "#fff",
          padding: "0.5rem",
          border: "none",
          borderRadius: 6,
          cursor: "pointer",
        }}
      >
        Login
      </button>
      {err && <p style={{ color: "red", fontSize: "0.85rem" }}>{err}</p>}
    </form>
  );
}
