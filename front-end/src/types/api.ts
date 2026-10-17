export interface ApiErrorBody {
  code: string
  message: string
}

export interface ApiEnvelope<T> {
  success: boolean
  data: T | null
  error: ApiErrorBody | null
  requestId: string
}
