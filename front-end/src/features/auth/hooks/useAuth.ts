import { useAuthStore } from '../stores/auth.store'

export function useAuth() {
  const status = useAuthStore((s) => s.status)
  const user = useAuthStore((s) => s.user)
  const initError = useAuthStore((s) => s.initError)
  const sessionExpired = useAuthStore((s) => s.sessionExpired)
  const networkError = useAuthStore((s) => s.networkError)
  const logout = useAuthStore((s) => s.logout)

  return {
    status,
    user,
    initError,
    sessionExpired,
    networkError,
    isAuthenticated: status === 'authenticated',
    isInitializing: status === 'initializing',
    logout,
  }
}
