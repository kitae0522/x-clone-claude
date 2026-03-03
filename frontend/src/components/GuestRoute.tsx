import { Navigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuthContext'

export default function GuestRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) return <p>Loading...</p>
  if (isAuthenticated) return <Navigate to="/" replace />

  return <>{children}</>
}
