import { create } from 'zustand'

export type DebugEventType =
  | 'REQUEST_SENT'
  | 'TOKEN_EXPIRED'
  | 'REFRESH_STARTED'
  | 'REFRESH_SUCCESS'
  | 'REFRESH_FAILED'
  | 'REQUEST_RETRIED'
  | 'REQUEST_SUCCESS'
  | 'SESSION_EXPIRED'

export interface DebugEvent {
  id: string
  timestamp: string
  type: DebugEventType
  requestId?: string
  status?: number
  message?: string
}

interface DebugState {
  events: DebugEvent[]
  refreshCount: number
  retryCount: number
  successCount: number
  failedCount: number

  record: (event: Omit<DebugEvent, 'id' | 'timestamp'>) => void
  incrementRefresh: () => void
  incrementRetry: () => void
  incrementSuccess: () => void
  incrementFailed: () => void
  clear: () => void
}

function newId(): string {
  return Math.random().toString(36).slice(2, 10) + Date.now().toString(36)
}

export const useDebugStore = create<DebugState>((set) => ({
  events: [],
  refreshCount: 0,
  retryCount: 0,
  successCount: 0,
  failedCount: 0,

  record: (event) =>
    set((s) => ({
      events: [
        ...s.events,
        { ...event, id: newId(), timestamp: new Date().toISOString() },
      ].slice(-200),
    })),

  incrementRefresh: () => set((s) => ({ refreshCount: s.refreshCount + 1 })),
  incrementRetry: () => set((s) => ({ retryCount: s.retryCount + 1 })),
  incrementSuccess: () => set((s) => ({ successCount: s.successCount + 1 })),
  incrementFailed: () => set((s) => ({ failedCount: s.failedCount + 1 })),

  clear: () =>
    set({
      events: [],
      refreshCount: 0,
      retryCount: 0,
      successCount: 0,
      failedCount: 0,
    }),
}))
