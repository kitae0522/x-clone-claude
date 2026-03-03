import { createContext } from 'react'
import type { User } from '@/types/api'

export interface AuthContextValue {
  user: User | null | undefined
  isLoading: boolean
  isAuthenticated: boolean
  logout: () => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)
