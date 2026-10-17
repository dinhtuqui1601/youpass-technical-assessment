import axios from 'axios'
import type { AxiosError, InternalAxiosRequestConfig } from 'axios'

import { tokenManager } from './token-manager'
import { refreshAccessToken } from './refresh-token'
import { getErrorCode, toApiError } from './api-error'
import { useAuthStore } from '@/features/auth/stores/auth.store'
import { useDebugStore } from '@/features/auth-debug/stores/debug.store'

// apiClient is the shared HTTP client. It attaches the in-memory access token
// and transparently refreshes it once when an API returns ACCESS_TOKEN_EXPIRED.
export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api',
  timeout: 15_000,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
})

const isDev = import.meta.env.DEV

interface RetriableConfig extends InternalAxiosRequestConfig {
  _retry?: boolean
}

apiClient.interceptors.request.use((config) => {
  const token = tokenManager.get()
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => {
    useAuthStore.getState().setNetworkError(false)

    if (isDev) {
      const cfg = response.config as RetriableConfig | undefined
      if (cfg?._retry) {
        useDebugStore.getState().incrementSuccess()
        useDebugStore.getState().record({
          type: 'REQUEST_SUCCESS',
          requestId: cfg.url ?? undefined,
          status: response.status,
        })
      }
    }
    return response
  },
  async (error: AxiosError) => {
    const config = error.config as RetriableConfig | undefined
    const status = error.response?.status
    const code = getErrorCode(error)

    // Retry once, and only for ACCESS_TOKEN_EXPIRED. FORBIDDEN and other auth
    // errors must not trigger a refresh.
    if (status === 401 && code === 'ACCESS_TOKEN_EXPIRED' && config && !config._retry) {
      config._retry = true

      if (isDev) {
        useDebugStore.getState().record({
          type: 'TOKEN_EXPIRED',
          requestId: config.url ?? undefined,
          status,
        })
      }

      try {
        await refreshAccessToken()

        if (isDev) {
          useDebugStore.getState().incrementRetry()
          useDebugStore.getState().record({
            type: 'REQUEST_RETRIED',
            requestId: config.url ?? undefined,
          })
        }

        const token = tokenManager.get()
        if (token) {
          config.headers.set('Authorization', `Bearer ${token}`)
        }
        return apiClient(config)
      } catch (refreshError) {
        if (isDev) {
          useDebugStore.getState().incrementFailed()
        }
        return Promise.reject(toApiError(refreshError))
      }
    }

    const apiErr = toApiError(error)
    if (apiErr.status === 0) {
      // Network error / timeout: never log the user out.
      useAuthStore.getState().setNetworkError(true)
    }
    if (isDev) {
      useDebugStore.getState().incrementFailed()
    }
    return Promise.reject(apiErr)
  },
)
