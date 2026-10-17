import { useMutation, useQueryClient } from '@tanstack/react-query'

import {
  createShareLink,
  createSubmission,
  deleteShareLink,
  updateShareLinkStatus,
} from '../api/share-links.api'
import type { ShareLinkStatus } from '../types'
import { shareLinksKeys } from './useShareLinks'
import { submissionsKeys } from '@/features/submissions/hooks/useSubmissions'

export function useCreateShareLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (submissionId: string) => createShareLink(submissionId),
    onSuccess: () => qc.invalidateQueries({ queryKey: shareLinksKeys.all }),
  })
}

export function useUpdateShareLinkStatus(code: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (status: ShareLinkStatus) => updateShareLinkStatus(code, status),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: shareLinksKeys.all })
      void qc.invalidateQueries({ queryKey: shareLinksKeys.detail(code) })
    },
  })
}

export function useDeleteShareLink() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (code: string) => deleteShareLink(code),
    onSuccess: () => qc.invalidateQueries({ queryKey: shareLinksKeys.all }),
  })
}

export function useCreateSubmission() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: { title: string; content: string }) => createSubmission(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: submissionsKeys.all }),
  })
}
