import { RouterProvider } from 'react-router-dom'

import { useSessionRestore } from '@/features/auth/hooks/useSessionRestore'

import { AppProviders } from './providers'
import { router } from './router'

function SessionBootstrap() {
  useSessionRestore()
  return null
}

export default function App() {
  return (
    <AppProviders>
      <SessionBootstrap />
      <RouterProvider router={router} />
    </AppProviders>
  )
}
