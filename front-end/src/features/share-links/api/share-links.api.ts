import { apiClient } from '@/lib/api'
import { unwrap } from '@/lib/api-error'
import type { ApiEnvelope } from '@/types/api'

import type { ShareLink, ShareLinkStatus, ShareStats, Submission } from '../types'

export async function listShareLinks(limit = 50, offset = 0): Promise<ShareLink[]> {
  const res = await apiClient.get<ApiEnvelope<ShareLink[]>>('/share-links', {
    params: { limit, offset },
  })
  return unwrap(res.data)
}

export async function getShareLink(code: string): Promise<ShareLink> {
  const res = await apiClient.get<ApiEnvelope<ShareLink>>(`/share-links/${code}`)
  return unwrap(res.data)
}

export async function updateShareLinkStatus(code: string, status: ShareLinkStatus): Promise<ShareLink> {
  const res = await apiClient.patch<ApiEnvelope<ShareLink>>(`/share-links/${code}`, { status })
  return unwrap(res.data)
}

export async function deleteShareLink(code: string): Promise<void> {
  await apiClient.delete(`/share-links/${code}`)
}

export async function getShareStats(code: string): Promise<ShareStats> {
  const res = await apiClient.get<ApiEnvelope<ShareStats>>(`/share-links/${code}/stats`)
  return unwrap(res.data)
}

export async function createSubmission(payload: { title: string; content: string }): Promise<Submission> {
  const res = await apiClient.post<ApiEnvelope<Submission>>('/submissions', payload)
  return unwrap(res.data)
}

export async function createShareLink(submissionId: string): Promise<ShareLink> {
  const res = await apiClient.post<ApiEnvelope<ShareLink>>(`/submissions/${submissionId}/share`)
  return unwrap(res.data)
}
