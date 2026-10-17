import { useQuery } from '@tanstack/react-query'

import { listShareLinks } from '../api/share-links.api'

export const shareLinksKeys = {
  all: ['share-links'] as const,
  detail: (code: string) => ['share-links', code] as const,
  stats: (code: string) => ['share-links', code, 'stats'] as const,
}

export function useShareLinks() {
  return useQuery({
    queryKey: shareLinksKeys.all,
    queryFn: () => listShareLinks(),
  })
}
