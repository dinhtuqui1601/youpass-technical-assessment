import { isAxiosError } from 'axios'
import type { AxiosError } from 'axios'

import type { ApiEnvelope } from '@/types/api'

// ApiError is the normalized error surfaced to the UI and TanStack Query.
export class ApiError extends Error {
  readonly code: string
  readonly status: number
  readonly requestId?: string

  constructor(code: string, message: string, status: number, requestId?: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
    this.requestId = requestId
  }
}

// unwrap extracts the data payload from a successful envelope, throwing an
// ApiError when the envelope reports a business error.
export function unwrap<T>(envelope: ApiEnvelope<T>): T {
  if (!envelope.success || envelope.error) {
    throw new ApiError(
      envelope.error?.code ?? 'UNKNOWN_ERROR',
      envelope.error?.message ?? 'Unknown error',
      0,
      envelope.requestId,
    )
  }
  return envelope.data as T
}

// toApiError normalizes any thrown value into an ApiError.
export function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error
  }
  if (isAxiosError(error)) {
    return fromAxiosError(error)
  }
  return new ApiError('UNKNOWN_ERROR', error instanceof Error ? error.message : 'Unknown error', 0)
}

function fromAxiosError(error: AxiosError): ApiError {
  const data = error.response?.data as Partial<ApiEnvelope<unknown>> | undefined
  const code = data?.error?.code ?? (error.code === 'ECONNABORTED' ? 'NETWORK_TIMEOUT' : 'NETWORK_ERROR')
  const message = data?.error?.message ?? error.message
  return new ApiError(code, message, error.response?.status ?? 0, data?.requestId)
}

// getErrorCode extracts the machine-readable error code from a rejected request.
export function getErrorCode(error: unknown): string | null {
  if (isAxiosError(error)) {
    const data = error.response?.data as Partial<ApiEnvelope<unknown>> | undefined
    return data?.error?.code ?? null
  }
  if (error instanceof ApiError) {
    return error.code
  }
  return null
}

// isAuthError reports whether an error is an authentication failure (HTTP 401),
// as opposed to a network or server error.
export function isAuthError(error: unknown): boolean {
  if (error instanceof ApiError) {
    return error.status === 401
  }
  if (isAxiosError(error)) {
    return error.response?.status === 401
  }
  return false
}
