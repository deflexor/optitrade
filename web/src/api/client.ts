import axios from 'axios'

const baseURL = import.meta.env.VITE_API_BASE ?? '/api/v1'

export const api = axios.create({
  baseURL,
  withCredentials: true,
  timeout: 30_000,
  headers: {
    Accept: 'application/json',
    'Content-Type': 'application/json',
  },
})
