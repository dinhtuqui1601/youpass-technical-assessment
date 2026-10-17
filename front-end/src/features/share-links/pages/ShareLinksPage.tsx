import {
  CheckCircleOutlined,
  EyeOutlined,
  LinkOutlined,
  PauseCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import { Button, Card, Col, Row, Space, Statistic, Typography } from 'antd'
import { useState } from 'react'

import { CreateShareModal } from '../components/CreateShareModal'
import { ShareLinkTable } from '../components/ShareLinkTable'
import { useShareLinks } from '../hooks/useShareLinks'
import { useShareLinksStats } from '../hooks/useShareLinksStats'

const STATS = {
  total: { icon: <LinkOutlined />, color: '#2563EB', bg: '#EFF6FF' },
  active: { icon: <CheckCircleOutlined />, color: '#16A34A', bg: '#F0FDF4' },
  disabled: { icon: <PauseCircleOutlined />, color: '#D97706', bg: '#FFFBEB' },
  views: { icon: <EyeOutlined />, color: '#64748B', bg: '#F1F5F9' },
} as const

export default function ShareLinksPage() {
  const [open, setOpen] = useState(false)
  const { data, isLoading, isError, refetch } = useShareLinks()
  const links = data ?? []
  const { statsByCode, totalViews } = useShareLinksStats(links)

  const activeCount = links.filter((l) => l.status === 'ACTIVE').length
  const disabledCount = links.filter((l) => l.status === 'DISABLED').length

  const viewsByCode: Record<string, number> = {}
  for (const [code, stats] of Object.entries(statsByCode)) {
    if (stats) {
      viewsByCode[code] = stats.totalViews
    }
  }

  const statsItems = [
    { key: 'total', label: 'Tổng số link', value: links.length },
    { key: 'active', label: 'Đang hoạt động', value: activeCount },
    { key: 'disabled', label: 'Đã tắt', value: disabledCount },
    { key: 'views', label: 'Tổng lượt xem', value: totalViews },
  ]

  return (
    <Space orientation="vertical" size={24} style={{ width: '100%' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 12 }}>
        <div>
          <Typography.Title level={4} style={{ margin: 0 }}>
            Quản lý link chia sẻ
          </Typography.Title>
          <Typography.Text type="secondary">
            Tạo và quản lý URL rút gọn để chia sẻ bài làm của bạn
          </Typography.Text>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>
          Tạo link chia sẻ
        </Button>
      </div>

      <Row gutter={[16, 16]}>
        {statsItems.map((s) => {
          const cfg = STATS[s.key as keyof typeof STATS]
          return (
            <Col xs={12} sm={12} lg={6} key={s.key}>
              <Card styles={{ body: { padding: 20 } }}>
                <Space size={14} align="start">
                  <div
                    style={{
                      width: 42,
                      height: 42,
                      borderRadius: 10,
                      background: cfg.bg,
                      color: cfg.color,
                      display: 'grid',
                      placeItems: 'center',
                      fontSize: 18,
                      flexShrink: 0,
                    }}
                  >
                    {cfg.icon}
                  </div>
                  <Statistic
                    title={s.label}
                    value={s.value}
                    styles={{ content: { fontSize: 24, fontWeight: 600 } }}
                  />
                </Space>
              </Card>
            </Col>
          )
        })}
      </Row>

      <ShareLinkTable
        links={links}
        loading={isLoading}
        error={isError}
        onRetry={() => void refetch()}
        viewsByCode={viewsByCode}
        onCreate={() => setOpen(true)}
      />

      <CreateShareModal open={open} onClose={() => setOpen(false)} />
    </Space>
  )
}

