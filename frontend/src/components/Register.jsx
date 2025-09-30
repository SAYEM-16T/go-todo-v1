// Register.js
import React, { useState } from "react";
import { authApi } from "../api";

export default function Register({ onRegister }) {
  const [username, setU] = useState("");
  const [email, setE] = useState("");
  const [password, setP] = useState("");
  const [err, setErr] = useState("");

  const submit = async (e) => {
    e.preventDefault();
    setErr("");
    try {
      const res = await authApi.post("/register", { username, email, password });
      onRegister(res.data.token);
    } catch (err) {
      const msg =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        "Register error";
      setErr(msg);
    }
  };

  return (
    <form onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: "0.75rem" }}>
      <h3 style={{ marginBottom: "1rem", textAlign: "center", color: "#333" }}>Register</h3>
      <input
        placeholder="Username"
        value={username}
        onChange={(e) => setU(e.target.value)}
        style={{ padding: "0.5rem", borderRadius: 6, border: "1px solid #ccc" }}
      />
      <input
        placeholder="Email"
        value={email}
        onChange={(e) => setE(e.target.value)}
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
          background: "#3498db",
          color: "#fff",
          padding: "0.5rem",
          border: "none",
          borderRadius: 6,
          cursor: "pointer",
        }}
      >
        Register
      </button>
      {err && <p style={{ color: "red", fontSize: "0.85rem" }}>{err}</p>}
    </form>
  );
}
