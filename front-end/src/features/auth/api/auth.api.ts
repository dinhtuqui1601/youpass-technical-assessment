import { apiClient } from '@/lib/api'
import { unwrap } from '@/lib/api-error'
import type { ApiEnvelope } from '@/types/api'

import type { AuthUser, LoginRequest, LoginResponse, RegisterRequest } from '../types'

export async function login(payload: LoginRequest): Promise<LoginResponse> {
  const res = await apiClient.post<ApiEnvelope<LoginResponse>>('/auth/login', payload)
  return unwrap(res.data)
}

export async function register(payload: RegisterRequest): Promise<AuthUser> {
  const res = await apiClient.post<ApiEnvelope<AuthUser>>('/auth/register', payload)
  return unwrap(res.data)
}

export async function getMe(): Promise<AuthUser> {
  const res = await apiClient.get<ApiEnvelope<AuthUser>>('/auth/me')
  return unwrap(res.data)
}
