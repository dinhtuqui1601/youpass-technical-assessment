import { Card, Typography } from 'antd'

import { LoginForm } from '../components/LoginForm'

export default function LoginPage() {
  return (
    <div style={{ width: '100%', maxWidth: 420 }}>
      <Typography.Title level={2} style={{ marginBottom: 4 }}>
        Đăng nhập
      </Typography.Title>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 24 }}>
        Chào mừng trở lại! Vui lòng đăng nhập để tiếp tục.
      </Typography.Paragraph>

      <Card styles={{ body: { padding: 24 } }}>
        <LoginForm />
      </Card>
    </div>
  )
}

