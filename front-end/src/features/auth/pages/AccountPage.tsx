import { UserOutlined } from '@ant-design/icons'
import { Avatar, Card, Descriptions, Space, Tag, Typography } from 'antd'

import { useAuth } from '../hooks/useAuth'
import { formatDateTime } from '@/utils/format'

export default function AccountPage() {
  const { user } = useAuth()

  return (
    <div className="yp-page-narrow">
      <Card styles={{ body: { padding: 24 } }}>
        <Space size={16} align="center" style={{ marginBottom: 24 }}>
          <Avatar size={64} style={{ background: '#2563EB', fontSize: 26 }}>
            {user?.name?.[0] ?? <UserOutlined />}
          </Avatar>
          <div>
            <Typography.Title level={4} style={{ margin: 0 }}>
              {user?.name ?? '—'}
            </Typography.Title>
            <Typography.Text type="secondary">{user?.email ?? '—'}</Typography.Text>
          </div>
        </Space>

        <Descriptions
          column={1}
          items={[
            { key: 'name', label: 'Họ tên', children: user?.name ?? '—' },
            { key: 'email', label: 'Email', children: user?.email ?? '—' },
            {
              key: 'role',
              label: 'Vai trò',
              children: user ? <Tag color="blue">{user.role}</Tag> : '—',
            },
            {
              key: 'createdAt',
              label: 'Ngày tạo',
              children: formatDateTime(user?.createdAt),
            },
          ]}
        />
      </Card>
    </div>
  )
}

