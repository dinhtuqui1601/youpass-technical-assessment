import { beforeEach, describe, expect, it } from 'vitest'

import { useDebugStore } from './debug.store'

describe('debug store', () => {
  beforeEach(() => useDebugStore.getState().clear())

  it('records events and increments counters', () => {
    useDebugStore.getState().record({ type: 'REFRESH_STARTED' })
    useDebugStore.getState().incrementRefresh()
    useDebugStore.getState().incrementRetry()

    const state = useDebugStore.getState()
    expect(state.events).toHaveLength(1)
    expect(state.events[0].type).toBe('REFRESH_STARTED')
    expect(state.events[0].timestamp).toBeTruthy()
    expect(state.refreshCount).toBe(1)
    expect(state.retryCount).toBe(1)
  })

  it('clear resets everything', () => {
    useDebugStore.getState().record({ type: 'TOKEN_EXPIRED', requestId: 'r1', status: 401 })
    useDebugStore.getState().incrementFailed()

    useDebugStore.getState().clear()

    const state = useDebugStore.getState()
    expect(state.events).toHaveLength(0)
    expect(state.failedCount).toBe(0)
    expect(state.refreshCount).toBe(0)
  })
})
