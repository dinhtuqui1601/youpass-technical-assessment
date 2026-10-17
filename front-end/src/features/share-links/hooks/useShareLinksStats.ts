import { useQueries } from '@tanstack/react-query'

import { getShareStats } from '../api/share-links.api'
import type { ShareLink, ShareStats } from '../types'
import { shareLinksKeys } from './useShareLinks'

// useShareLinksStats fetches per-link statistics for the currently visible links
// in parallel, then derives the total view count. The backend exposes stats only
// per-code (no aggregate endpoint), so this is the accurate way to compute it.
export function useShareLinksStats(links: ShareLink[] | undefined) {
  const codes = links?.map((l) => l.code) ?? []

  const results = useQueries({
    queries: codes.map((code) => ({
      queryKey: shareLinksKeys.stats(code),
      queryFn: () => getShareStats(code),
      enabled: codes.length > 0,
      refetchInterval: 30_000 as const,
    })),
  })

  const statsByCode: Record<string, ShareStats | undefined> = {}
  let totalViews = 0
  let loading = false

  results.forEach((result, index) => {
    const code = codes[index]
    statsByCode[code] = result.data
    if (result.data) {
      totalViews += result.data.totalViews
    }
    if (result.isLoading) {
      loading = true
    }
  })

  return { statsByCode, totalViews, loading }
}
