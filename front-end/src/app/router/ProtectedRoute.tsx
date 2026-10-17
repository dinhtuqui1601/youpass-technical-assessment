import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { FullPageLoader } from '@/components/common/FullPageLoader'
import { SessionError } from '@/components/common/SessionError'
import { useAuth } from '@/features/auth/hooks/useAuth'

export function ProtectedRoute() {
  const { status, initError } = useAuth()
  const location = useLocation()

  if (status === 'initializing') {
    return initError ? <SessionError /> : <FullPageLoader />
  }
  if (status === 'unauthenticated') {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }
  return <Outlet />
}
