import { useQuery } from '@tanstack/react-query'

import { getShareLink, getShareStats } from '../api/share-links.api'
import { shareLinksKeys } from './useShareLinks'

export function useShareLinkDetail(code: string) {
  return useQuery({
    queryKey: shareLinksKeys.detail(code),
    queryFn: () => getShareLink(code),
    enabled: !!code,
  })
}

export function useShareLinkStats(code: string) {
  return useQuery({
    queryKey: shareLinksKeys.stats(code),
    queryFn: () => getShareStats(code),
    enabled: !!code,
    // Gentle polling: statistics are eventually consistent.
    refetchInterval: 30_000,
  })
}
