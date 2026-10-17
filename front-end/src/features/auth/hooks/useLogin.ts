import { useMutation } from '@tanstack/react-query'

import { tokenManager } from '@/lib/token-manager'

import { login } from '../api/auth.api'
import { useAuthStore } from '../stores/auth.store'
import type { LoginRequest } from '../types'

export function useLogin() {
  return useMutation({
    mutationFn: (payload: LoginRequest) => login(payload),
    onSuccess: (data) => {
      tokenManager.set(data.accessToken, data.expiresIn)
      useAuthStore.getState().setAuthenticated(data.user)
    },
  })
}
