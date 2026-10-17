import { http, HttpResponse } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'

import { server } from '@/test/msw/server'
import { useAuthStore } from '@/features/auth/stores/auth.store'

import { apiClient } from './api'
import { tokenManager } from './token-manager'

const user = {
  id: 'u1',
  email: 'demo@youpass.vn',
  name: 'Demo',
  role: 'student',
  createdAt: '2026-10-17T00:00:00Z',
}

function resetAuth() {
  tokenManager.clear()
  useAuthStore.setState({
    status: 'authenticated',
    user,
    initError: false,
    sessionVersion: 0,
  })
}

describe('apiClient interceptors', () => {
  beforeEach(resetAuth)

  it('attaches the access token and retries once after ACCESS_TOKEN_EXPIRED (single-flight)', async () => {
    let refreshCalls = 0

    server.use(
      http.post('/api/auth/refresh', () => {
        refreshCalls += 1
        return HttpResponse.json({
          success: true,
          data: { accessToken: 'new-token', expiresIn: 900, user },
          error: null,
          requestId: 'r-refresh',
        })
      }),
      http.get('/api/protected', ({ request }) => {
        const auth = request.headers.get('Authorization')
        if (auth === 'Bearer new-token') {
          return HttpResponse.json({ success: true, data: { ok: true }, error: null, requestId: 'r-ok' })
        }
        return HttpResponse.json(
          { success: false, data: null, error: { code: 'ACCESS_TOKEN_EXPIRED', message: 'expired' }, requestId: 'r-expired' },
          { status: 401 },
        )
      }),
    )

    tokenManager.set('old-token')

    // 5 concurrent requests -> a single refresh request.
    const results = await Promise.all([
      apiClient.get('/protected'),
      apiClient.get('/protected'),
      apiClient.get('/protected'),
      apiClient.get('/protected'),
      apiClient.get('/protected'),
    ])

    for (const res of results) {
      expect(res.data.data).toEqual({ ok: true })
    }
    expect(refreshCalls).toBe(1)
  })

  it('does not refresh on non-ACCESS_TOKEN_EXPIRED errors', async () => {
    let refreshCalls = 0

    server.use(
      http.post('/api/auth/refresh', () => {
        refreshCalls += 1
        return HttpResponse.json({ success: true, data: { accessToken: 'x', expiresIn: 900, user }, error: null, requestId: 'r' })
      }),
      http.get('/api/protected', () =>
        HttpResponse.json(
          { success: false, data: null, error: { code: 'FORBIDDEN', message: 'forbidden' }, requestId: 'r-forbidden' },
          { status: 403 },
        ),
      ),
    )

    tokenManager.set('old-token')

    await expect(apiClient.get('/protected')).rejects.toMatchObject({ code: 'FORBIDDEN' })
    expect(refreshCalls).toBe(0)
  })

  it('does not retry infinitely when the retried request still expires', async () => {
    let refreshCalls = 0

    server.use(
      http.post('/api/auth/refresh', () => {
        refreshCalls += 1
        return HttpResponse.json({ success: true, data: { accessToken: 'new-token', expiresIn: 900, user }, error: null, requestId: 'r' })
      }),
      http.get('/api/protected', () =>
        HttpResponse.json(
          { success: false, data: null, error: { code: 'ACCESS_TOKEN_EXPIRED', message: 'expired' }, requestId: 'r-expired' },
          { status: 401 },
        ),
      ),
    )

    tokenManager.set('old-token')

    // Retry happens once (the _retry guard prevents a loop).
    await expect(apiClient.get('/protected')).rejects.toMatchObject({ code: 'ACCESS_TOKEN_EXPIRED' })
    expect(refreshCalls).toBe(1)
  })
})
