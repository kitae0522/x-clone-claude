import { useMe, useLogout } from '@/hooks/useAuth'
import { AuthContext } from '@/contexts/authContextValue'
import type { AuthContextValue } from '@/contexts/authContextValue'

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { data: user, isLoading } = useMe()
  const logoutMutation = useLogout()

  const value: AuthContextValue = {
    user: user ?? null,
    isLoading,
    isAuthenticated: !!user,
    logout: () => logoutMutation.mutate(),
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
