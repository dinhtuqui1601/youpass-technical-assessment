import { useEffect, useRef } from 'react'

import { refreshAccessToken } from '@/lib/refresh-token'

import { useAuthStore } from '../stores/auth.store'

// restoreSession attempts to restore the session from the HttpOnly refresh
// cookie. On a genuine auth failure the store becomes unauthenticated; on a
// network error the store stays in initializing with initError set (retryable).
export async function restoreSession(): Promise<void> {
  const store = useAuthStore.getState()
  if (store.status !== 'initializing') {
    return
  }
  store.setInitError(false)

  try {
    await refreshAccessToken()
  } catch {
    const latest = useAuthStore.getState()
    if (latest.status === 'initializing') {
      latest.setInitError(true)
    }
  }
}

// useSessionRestore triggers a single session restore on mount. It is safe to
// call from multiple components because refreshAccessToken is single-flight.
export function useSessionRestore() {
  const startedRef = useRef(false)

  useEffect(() => {
    if (startedRef.current) {
      return
    }
    startedRef.current = true
    void restoreSession()
  }, [])

  return { retry: () => void restoreSession() }
}
