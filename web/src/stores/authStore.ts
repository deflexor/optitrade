import { create } from 'zustand'
import { api } from '../api/client'

type AuthState = {
  username: string | null
  ready: boolean
  loginError: string | null
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  bootstrap: () => Promise<void>
  clearError: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  username: null,
  ready: false,
  loginError: null,

  clearError: () => set({ loginError: null }),

  bootstrap: async () => {
    try {
      const { data } = await api.get<{ username: string }>('/auth/me')
      set({ username: data.username, ready: true })
    } catch {
      set({ username: null, ready: true })
    }
  },

  login: async (username: string, password: string) => {
    set({ loginError: null })
    try {
      await api.post('/auth/login', { username, password })
      await get().bootstrap()
    } catch {
      set({
        loginError: 'Sign-in failed. Check credentials.',
        username: null,
      })
      throw new Error('login failed')
    }
  },

  logout: async () => {
    try {
      await api.post('/auth/logout')
    } finally {
      set({ username: null })
    }
  },
}))
