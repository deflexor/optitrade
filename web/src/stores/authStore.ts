import { create } from 'zustand'

type AuthState = {
  isAuthed: boolean
}

export const useAuthStore = create<AuthState>(() => ({
  isAuthed: false,
}))
