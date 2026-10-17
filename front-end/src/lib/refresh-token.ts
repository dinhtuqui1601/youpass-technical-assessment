import { authClient } from './auth-client'
import { tokenManager } from './token-manager'
import { ApiError, isAuthError, toApiError } from './api-error'
import { useAuthStore } from '@/features/auth/stores/auth.store'
import { useDebugStore } from '@/features/auth-debug/stores/debug.store'
import type { LoginResponse } from '@/features/auth/types'
import type { ApiEnvelope } from '@/types/api'

let refreshPromise: Promise<string> | null = null

const isDev = import.meta.env.DEV

// refreshAccessToken performs a single-flight refresh: concurrent callers share
// one in-flight request instead of each firing their own.
export function refreshAccessToken(): Promise<string> {
  if (!refreshPromise) {
    const version = useAuthStore.getState().sessionVersion

    if (isDev) {
      useDebugStore.getState().incrementRefresh()
      useDebugStore.getState().record({ type: 'REFRESH_STARTED' })
    }

    refreshPromise = authClient
      .post<ApiEnvelope<LoginResponse>>('/auth/refresh')
      .then((res) => {
        // If the user logged out while the refresh was in-flight, discard the
        // stale response instead of resurrecting the session.
        if (useAuthStore.getState().sessionVersion !== version) {
          tokenManager.clear()
          throw new ApiError('SESSION_ENDED', 'Session ended', 401)
        }

        const data = res.data.data
        if (!data?.accessToken) {
          throw new ApiError('INVALID_REFRESH_TOKEN', 'No access token returned', 401)
        }

        tokenManager.set(data.accessToken, data.expiresIn)
        useAuthStore.getState().setAuthenticated(data.user)

        if (isDev) {
          useDebugStore.getState().record({ type: 'REFRESH_SUCCESS' })
        }
        return data.accessToken
      })
      .catch((err) => {
        const apiErr = toApiError(err)
        const authErr = isAuthError(err)

        if (isDev) {
          useDebugStore.getState().record({ type: 'REFRESH_FAILED', message: apiErr.message })
        }

        if (authErr) {
          const store = useAuthStore.getState()
          if (store.status === 'authenticated') {
            // Mid-session expiry: keep the current page mounted (learner work is
            // preserved) and show the Session Expired modal instead.
            if (isDev) {
              useDebugStore.getState().record({ type: 'SESSION_EXPIRED' })
            }
            store.setSessionExpired(true)
          } else {
            store.setUnauthenticated()
          }
        }

        throw apiErr
      })
      .finally(() => {
        refreshPromise = null
      })
  }

  return refreshPromise
}

// isRefreshing reports whether a refresh is currently in-flight.
export function isRefreshing(): boolean {
  return refreshPromise !== null
}
