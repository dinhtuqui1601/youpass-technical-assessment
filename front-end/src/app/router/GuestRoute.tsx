import { Navigate, Outlet } from 'react-router-dom'

import { FullPageLoader } from '@/components/common/FullPageLoader'
import { useAuth } from '@/features/auth/hooks/useAuth'

export function GuestRoute() {
  const { status } = useAuth()

  if (status === 'initializing') {
    return <FullPageLoader />
  }
  if (status === 'authenticated') {
    return <Navigate to="/share-links" replace />
  }
  return <Outlet />
}
