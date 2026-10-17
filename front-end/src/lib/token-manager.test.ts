import { beforeEach, describe, expect, it } from 'vitest'

import { tokenManager } from './token-manager'

describe('tokenManager', () => {
  beforeEach(() => tokenManager.clear())

  it('is empty by default', () => {
    expect(tokenManager.get()).toBeNull()
  })

  it('stores and clears a token', () => {
    tokenManager.set('access-token', 60)
    expect(tokenManager.get()).toBe('access-token')
    expect(tokenManager.getExpiresAt()).toBeGreaterThan(Date.now())

    tokenManager.clear()
    expect(tokenManager.get()).toBeNull()
    expect(tokenManager.getExpiresAt()).toBeNull()
  })
})
