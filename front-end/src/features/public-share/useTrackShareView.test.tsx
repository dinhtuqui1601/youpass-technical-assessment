import { renderHook, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import { server } from '@/test/msw/server'

import { useTrackShareView } from './hooks/useTrackShareView'

describe('useTrackShareView', () => {
  it('fires a single view event only after content loads', async () => {
    let calls = 0
    server.use(
      http.post('/api/public/shares/abc/views', () => {
        calls += 1
        return HttpResponse.json({ success: true, data: { accepted: true }, error: null, requestId: 'r' })
      }),
    )

    const { rerender } = renderHook(({ ok }: { ok: boolean }) => useTrackShareView('abc', ok), {
      initialProps: { ok: false },
    })

    expect(calls).toBe(0)

    rerender({ ok: true })
    await waitFor(() => expect(calls).toBe(1))

    // Re-renders / StrictMode-style remounts must not re-fire.
    rerender({ ok: false })
    rerender({ ok: true })
    expect(calls).toBe(1)
  })

  it('does not fire without a code', () => {
    let calls = 0
    server.use(
      http.post('/api/public/shares/x/views', () => {
        calls += 1
        return HttpResponse.json({ success: true, data: { accepted: true }, error: null, requestId: 'r' })
      }),
    )
    renderHook(() => useTrackShareView(undefined, true))
    expect(calls).toBe(0)
  })
})
