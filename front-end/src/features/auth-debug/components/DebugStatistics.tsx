import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  SyncOutlined,
} from '@ant-design/icons'
import { Card, Col, Row, Space, Statistic } from 'antd'

import { useDebugStore } from '../stores/debug.store'

export function DebugStatistics() {
  const refreshCount = useDebugStore((s) => s.refreshCount)
  const retryCount = useDebugStore((s) => s.retryCount)
  const successCount = useDebugStore((s) => s.successCount)
  const failedCount = useDebugStore((s) => s.failedCount)

  const items = [
    { label: 'Refresh Requests', value: refreshCount, icon: <ReloadOutlined />, color: '#2563EB' },
    { label: 'Successful Retries', value: retryCount, icon: <SyncOutlined />, color: '#2563EB' },
    { label: 'Success Requests', value: successCount, icon: <CheckCircleOutlined />, color: '#16A34A' },
    { label: 'Failed Requests', value: failedCount, icon: <CloseCircleOutlined />, color: '#DC2626' },
  ]

  return (
    <Row gutter={[16, 16]}>
      {items.map((it) => (
        <Col xs={12} lg={6} key={it.label}>
          <Card styles={{ body: { padding: 20 } }}>
            <Space size={12} align="start">
              <div
                style={{
                  width: 40,
                  height: 40,
                  borderRadius: 10,
                  background: `${it.color}14`,
                  color: it.color,
                  display: 'grid',
                  placeItems: 'center',
                  fontSize: 18,
                  flexShrink: 0,
                }}
              >
                {it.icon}
              </div>
              <Statistic
                title={it.label}
                value={it.value}
                styles={{ content: { fontSize: 22, fontWeight: 600 } }}
              />
            </Space>
          </Card>
        </Col>
      ))}
    </Row>
  )
}

