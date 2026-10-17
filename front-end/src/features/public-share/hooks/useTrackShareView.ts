import { useEffect, useRef } from 'react'

import { trackView } from '../api/public-share.api'

function newViewId(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  // Fallback UUID v4 generator.
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

// useTrackShareView records a view once per page-open, only after the shared
// content has loaded successfully. It is best-effort and never blocks content.
export function useTrackShareView(code: string | undefined, isSuccess: boolean) {
  const viewIdRef = useRef<string | null>(null)
  const sentRef = useRef(false)

  useEffect(() => {
    if (!code || !isSuccess || sentRef.current) {
      return
    }
    sentRef.current = true
    if (!viewIdRef.current) {
      viewIdRef.current = newViewId()
    }
    void trackView(code, viewIdRef.current).catch(() => {
      // View tracking must never break the page.
    })
  }, [code, isSuccess])
}
