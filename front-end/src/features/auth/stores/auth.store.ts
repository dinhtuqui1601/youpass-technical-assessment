import { create } from 'zustand'

import { authClient } from '@/lib/auth-client'
import { queryClient } from '@/lib/queryClient'
import { tokenManager } from '@/lib/token-manager'

import type { AuthUser } from '../types'

export type AuthStatus = 'initializing' | 'authenticated' | 'unauthenticated'

interface AuthState {
  status: AuthStatus
  user: AuthUser | null
  // initError is set when session restore failed due to a network error
  // (the session may still be valid and the user can retry).
  initError: boolean
  // sessionVersion increments on logout so an in-flight refresh response can
  // be discarded instead of resurrecting a closed session.
  sessionVersion: number
  // sessionExpired is set when the refresh token itself is invalid/revoked
  // mid-session. It triggers the Session Expired modal without unmounting the
  // current page (so learner work is preserved).
  sessionExpired: boolean
  // networkError is set when a request fails with a network error (no logout).
  networkError: boolean

  setInitializing: () => void
  setAuthenticated: (user: AuthUser) => void
  setUnauthenticated: () => void
  setInitError: (value: boolean) => void
  setSessionExpired: (value: boolean) => void
  setNetworkError: (value: boolean) => void
  logout: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set, get) => ({
  status: 'initializing',
  user: null,
  initError: false,
  sessionVersion: 0,
  sessionExpired: false,
  networkError: false,

  setInitializing: () => set({ status: 'initializing', initError: false }),
  setAuthenticated: (user) =>
    set({ status: 'authenticated', user, initError: false, sessionExpired: false }),
  setUnauthenticated: () =>
    set({ status: 'unauthenticated', user: null, initError: false, sessionExpired: false }),
  setInitError: (value) => set({ initError: value }),
  setSessionExpired: (value) => set({ sessionExpired: value }),
  setNetworkError: (value) => set({ networkError: value }),

  logout: async () => {
    set({ sessionVersion: get().sessionVersion + 1 })
    tokenManager.clear()
    set({
      status: 'unauthenticated',
      user: null,
      sessionExpired: false,
      networkError: false,
    })
    queryClient.clear()

    try {
      await authClient.post('/auth/logout')
    } catch {
      // Logout is best-effort; local state is already cleared.
    }
  },
}))
