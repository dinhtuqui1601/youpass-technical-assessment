import { apiClient } from '@/lib/api'
import { unwrap } from '@/lib/api-error'
import type { ApiEnvelope } from '@/types/api'

import type { Submission } from '../types'

export async function listSubmissions(): Promise<Submission[]> {
  const res = await apiClient.get<ApiEnvelope<Submission[]>>('/submissions')
  return unwrap(res.data)
}

export async function createSubmission(payload: {
  title: string
  content: string
}): Promise<Submission> {
  const res = await apiClient.post<ApiEnvelope<Submission>>('/submissions', payload)
  return unwrap(res.data)
}
