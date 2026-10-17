import { http, HttpResponse } from 'msw'

export const testUser = {
  id: 'u1',
  email: 'demo@youpass.vn',
  name: 'Demo User',
  role: 'student',
  createdAt: '2026-10-17T00:00:00Z',
}

const ok = (data: unknown, requestId = 'req-1') =>
  HttpResponse.json({ success: true, data, error: null, requestId })

export const handlers = [
  http.post('/api/auth/login', () =>
    ok({ accessToken: 'test-access-token', expiresIn: 900, user: testUser }, 'req-login'),
  ),
  http.post('/api/auth/refresh', () =>
    ok({ accessToken: 'refreshed-token', expiresIn: 900, user: testUser }, 'req-refresh'),
  ),
  http.post('/api/auth/logout', () => ok({ loggedOut: true }, 'req-logout')),
  http.get('/api/auth/me', () => ok(testUser, 'req-me')),
]
