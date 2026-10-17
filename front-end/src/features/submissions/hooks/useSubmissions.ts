import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { createSubmission, listSubmissions } from '../api/submissions.api'

export const submissionsKeys = {
  all: ['submissions'] as const,
}

export function useSubmissions() {
  return useQuery({
    queryKey: submissionsKeys.all,
    queryFn: () => listSubmissions(),
  })
}

export function useCreateSubmission() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: { title: string; content: string }) => createSubmission(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: submissionsKeys.all }),
  })
}
