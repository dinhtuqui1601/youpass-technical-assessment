import { Card, Empty, Space, Tag, Timeline, Typography } from 'antd'
import dayjs from 'dayjs'

import { useDebugStore } from '../stores/debug.store'
import type { DebugEventType } from '../stores/debug.store'

const typeColor: Record<DebugEventType, string> = {
  REQUEST_SENT: 'default',
  TOKEN_EXPIRED: 'red',
  REFRESH_STARTED: 'processing',
  REFRESH_SUCCESS: 'green',
  REFRESH_FAILED: 'red',
  REQUEST_RETRIED: 'blue',
  REQUEST_SUCCESS: 'green',
  SESSION_EXPIRED: 'red',
}

export function DebugTimeline() {
  const events = useDebugStore((s) => s.events)

  return (
    <Card title="Debug Timeline" size="small">
      {events.length === 0 ? (
        <Empty description="Chưa có sự kiện. Chạy một test ở trên." />
      ) : (
        <Timeline
          items={events.map((e) => ({
            color: typeColor[e.type],
            content: (
              <Space wrap size={8}>
                <Tag color={typeColor[e.type]}>{e.type}</Tag>
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  {dayjs(e.timestamp).format('HH:mm:ss.SSS')}
                </Typography.Text>
                {e.requestId && (
                  <Typography.Text code style={{ fontSize: 12 }}>
                    {e.requestId}
                  </Typography.Text>
                )}
                {e.status !== undefined && (
                  <Typography.Text style={{ fontSize: 12 }}>status {e.status}</Typography.Text>
                )}
                {e.message && (
                  <Typography.Text style={{ fontSize: 12 }}>{e.message}</Typography.Text>
                )}
              </Space>
            ),
          }))}
        />
      )}
    </Card>
  )
}
