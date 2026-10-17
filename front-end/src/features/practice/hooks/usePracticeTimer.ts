import { useEffect, useState } from 'react'

// usePracticeTimer derives the remaining time from a persisted start timestamp,
// so the countdown is not reset by re-renders, token refresh or reload.
export function usePracticeTimer(startedAt: number | null, durationSec: number) {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])

  const elapsed = startedAt ? Math.max(0, Math.floor((now - startedAt) / 1000)) : 0
  const remaining = Math.max(0, durationSec - elapsed)

  return { remaining, elapsed }
}
