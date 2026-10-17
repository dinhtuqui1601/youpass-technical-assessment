import { describe, expect, it } from 'vitest'

import { ApiError, toApiError, unwrap } from './api-error'

describe('api-error', () => {
  it('unwrap returns data on success', () => {
    const result = unwrap({ success: true, data: { a: 1 }, error: null, requestId: 'r' })
    expect(result).toEqual({ a: 1 })
  })

  it('unwrap throws an ApiError on failure', () => {
    expect(() =>
      unwrap({ success: false, data: null, error: { code: 'X', message: 'boom' }, requestId: 'r' }),
    ).toThrowError('boom')
  })

  it('toApiError normalizes an axios 401 into an ApiError', () => {
    const err = toApiError({
      isAxiosError: true,
      code: 'ERR_BAD_REQUEST',
      message: 'Request failed',
      response: {
        status: 401,
        data: { success: false, data: null, error: { code: 'ACCESS_TOKEN_EXPIRED', message: 'expired' }, requestId: 'r1' },
      },
    })

    expect(err).toBeInstanceOf(ApiError)
    expect(err.code).toBe('ACCESS_TOKEN_EXPIRED')
    expect(err.status).toBe(401)
    expect(err.message).toBe('expired')
  })

  it('toApiError falls back to a network error without a response', () => {
    const err = toApiError({ isAxiosError: true, code: 'ERR_NETWORK', message: 'Network Error' })
    expect(err.code).toBe('NETWORK_ERROR')
    expect(err.status).toBe(0)
  })
})
