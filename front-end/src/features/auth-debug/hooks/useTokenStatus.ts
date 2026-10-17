import { useEffect, useState } from 'react'

import { isRefreshing } from '@/lib/refresh-token'
import { tokenManager } from '@/lib/token-manager'

export type TokenState = 'none' | 'valid' | 'expired' | 'refreshing'

export function useTokenStatus() {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])

  const hasToken = tokenManager.get() !== null
  const expiresAt = tokenManager.getExpiresAt()
  const expired = hasToken && expiresAt !== null && expiresAt <= now
  const refreshing = isRefreshing()

  let state: TokenState = 'none'
  if (refreshing) {
    state = 'refreshing'
  } else if (!hasToken) {
    state = 'none'
  } else if (expired) {
    state = 'expired'
  } else {
    state = 'valid'
  }

  const remainingMs = expiresAt ? Math.max(0, expiresAt - now) : null

  return { state, expiresAt, remainingMs }
}
