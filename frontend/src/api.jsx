// src/api.js
import axios from "axios";

// Use Vite env variables
const API_AUTH_BASE = import.meta.env.VITE_AUTH_URL;

export const authApi = axios.create({
  baseURL: `${API_AUTH_BASE}/api/v1/auth`,
});


export function setAuthToken(token) {
  if (token) {
    authApi.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  } else {
    delete authApi.defaults.headers.common["Authorization"];
  }
}
