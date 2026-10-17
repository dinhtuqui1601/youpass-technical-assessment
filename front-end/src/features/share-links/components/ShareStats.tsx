import { Card, Col, Descriptions, Row, Statistic, Tag } from 'antd'

import { formatDateTime } from '@/utils/format'

import type { ShareLink, ShareLinkStatus, ShareStats } from '../types'

const statusColor: Record<ShareLinkStatus, string> = {
  ACTIVE: 'green',
  DISABLED: 'orange',
  DELETED: 'red',
}

interface Props {
  link?: ShareLink
  stats?: ShareStats
  loading: boolean
}

export function ShareStats({ link, stats, loading }: Props) {
  return (
    <Row gutter={[16, 16]}>
      <Col xs={24} sm={8}>
        <Card loading={loading}>
          <Statistic title="Tổng lượt xem" value={stats?.totalViews ?? 0} />
        </Card>
      </Col>
      <Col xs={24} sm={8}>
        <Card loading={loading}>
          <Statistic title="Lượt xem gần nhất" value={formatDateTime(stats?.lastViewedAt)} />
        </Card>
      </Col>
      <Col xs={24} sm={8}>
        <Card loading={loading}>
          <Descriptions
            column={1}
            size="small"
            items={[
              {
                key: 'status',
                label: 'Trạng thái',
                children: link ? <Tag color={statusColor[link.status]}>{link.status}</Tag> : '—',
              },
              {
                key: 'created',
                label: 'Ngày tạo',
                children: formatDateTime(link?.createdAt),
              },
              {
                key: 'expires',
                label: 'Hết hạn',
                children: formatDateTime(link?.expiresAt),
              },
            ]}
          />
        </Card>
      </Col>
    </Row>
  )
}
