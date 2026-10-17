import axios from 'axios'

// Dedicated client for the refresh endpoint. It intentionally does NOT go
// through the main response interceptor, which would otherwise trigger a
// recursive refresh loop.
export const authClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api',
  timeout: 10_000,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
})
