import { App, Button, Card, Divider, Flex, Space, Tag, Typography } from 'antd'
import { useState } from 'react'

import { apiClient } from '@/lib/api'
import { toApiError } from '@/lib/api-error'
import { refreshAccessToken } from '@/lib/refresh-token'

import { useDebugStore } from '../stores/debug.store'

interface ResultItem {
  label: string
  status: string | number
  note: string
}

export function ConcurrentTestPanel() {
  const { message } = App.useApp()
  const [results, setResults] = useState<ResultItem[]>([])
  const [running, setRunning] = useState(false)

  const record = useDebugStore((s) => s.record)
  const incrementRetry = useDebugStore((s) => s.incrementRetry)
  const incrementSuccess = useDebugStore((s) => s.incrementSuccess)
  const incrementFailed = useDebugStore((s) => s.incrementFailed)

  const runRealConcurrent = async () => {
    setRunning(true)
    setResults([])
    const outcome = await Promise.all(
      Array.from({ length: 5 }, async (_, i) => {
        try {
          const res = await apiClient.get('/share-links')
          return { label: `Request ${i + 1}`, status: res.status, note: 'ok' }
        } catch (e) {
          const err = toApiError(e)
          return { label: `Request ${i + 1}`, status: err.status, note: err.message }
        }
      }),
    )
    setResults(outcome)
    setRunning(false)
  }

  const runSimulatedExpired = async () => {
    setRunning(true)
    setResults([])

    await Promise.all(
      Array.from({ length: 5 }, async (_, i) => {
        record({ type: 'TOKEN_EXPIRED', requestId: `sim-${i + 1}`, status: 401 })
        try {
          // Single-flight: all 5 share one real /auth/refresh request.
          await refreshAccessToken()
          record({ type: 'REQUEST_RETRIED', requestId: `sim-${i + 1}` })
          record({ type: 'REQUEST_SUCCESS', requestId: `sim-${i + 1}`, status: 200 })
          incrementRetry()
          incrementSuccess()
          return 'ok'
        } catch (e) {
          incrementFailed()
          return 'fail'
        }
      }),
    )

    setResults([
      {
        label: 'Simulated Test',
        status: 'ok',
        note: '5 expired requests → 1 real refresh (xem timeline & Refresh Requests)',
      },
    ])
    setRunning(false)
    message.info('Đã chạy 5 request mô phỏng với token hết hạn')
  }

  const simulateRefreshFailure = () => {
    record({ type: 'REFRESH_STARTED' })
    record({ type: 'REFRESH_FAILED', message: 'Simulated refresh failure' })
    record({ type: 'SESSION_EXPIRED' })
    incrementFailed()
  }

  const simulateNetworkFailure = () => {
    record({ type: 'REFRESH_STARTED' })
    record({ type: 'REFRESH_FAILED', message: 'Simulated network timeout' })
    incrementFailed()
  }

  const clearLogs = () => {
    useDebugStore.getState().clear()
    setResults([])
  }

  return (
    <Card title="Test Controls">
      <Space orientation="vertical" size={16} style={{ width: '100%' }}>
        <div>
          <Typography.Text strong style={{ fontSize: 14 }}>
            Real API Tests
          </Typography.Text>
          <div style={{ marginTop: 8 }}>
            <Button type="primary" loading={running} onClick={() => void runRealConcurrent()}>
              Test 5 Concurrent Requests (real API)
            </Button>
          </div>
        </div>

        <div>
          <Typography.Text strong style={{ fontSize: 14 }}>
            Simulated Tests
          </Typography.Text>
          <div style={{ marginTop: 8 }}>
            <Space wrap>
              <Button loading={running} onClick={() => void runSimulatedExpired()}>
                Simulate Expired Token (5 concurrent)
              </Button>
              <Button onClick={simulateRefreshFailure}>Test Refresh Failure (Simulated)</Button>
              <Button onClick={simulateNetworkFailure}>Test Network Failure (Simulated)</Button>
            </Space>
          </div>
        </div>

        <Divider style={{ margin: '4px 0' }} />
        <Button danger onClick={clearLogs}>
          Clear Debug Logs
        </Button>

        {results.length > 0 && (
          <Flex vertical gap={8}>
            {results.map((r) => (
              <Space key={r.label} wrap>
                <Typography.Text>{r.label}</Typography.Text>
                <Tag color={r.status === 200 || r.status === 'ok' ? 'green' : 'red'}>{r.status}</Tag>
                <Typography.Text type="secondary">{r.note}</Typography.Text>
              </Space>
            ))}
          </Flex>
        )}
      </Space>
    </Card>
  )
}

