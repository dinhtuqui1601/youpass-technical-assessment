export type { Submission } from '@/features/submissions/types'

export type ShareLinkStatus = 'ACTIVE' | 'DISABLED' | 'DELETED'

export interface ShareLink {
  id: string
  code: string
  submissionId: string
  status: ShareLinkStatus
  expiresAt?: string | null
  url: string
  createdAt: string
  updatedAt: string
}

export interface ShareStats {
  code: string
  totalViews: number
  lastViewedAt?: string | null
}
