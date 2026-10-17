import { apiClient } from '@/lib/api'
import { unwrap } from '@/lib/api-error'
import type { ApiEnvelope } from '@/types/api'

import type { PublicShare } from '../types'

export async function getPublicShare(code: string): Promise<PublicShare> {
  const res = await apiClient.get<ApiEnvelope<PublicShare>>(`/public/shares/${code}`)
  return unwrap(res.data)
}

export async function trackView(code: string, viewId: string): Promise<void> {
  await apiClient.post(`/public/shares/${code}/views`, { viewId })
}
