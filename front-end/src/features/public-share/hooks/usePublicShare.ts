import { useQuery } from '@tanstack/react-query'

import { ApiError } from '@/lib/api-error'

import { getPublicShare } from '../api/public-share.api'

export function usePublicShare(code: string) {
  return useQuery({
    queryKey: ['public-share', code],
    queryFn: () => getPublicShare(code),
    enabled: !!code,
    // Never serve stale protected content: always refetch and never cache
    // once the component unmounts.
    staleTime: 0,
    gcTime: 0,
    retry: (failureCount, error) => {
      if (error instanceof ApiError && error.status > 0 && error.status < 500) {
        return false
      }
      return failureCount < 1
    },
  })
}
