import { Alert, Space, Typography } from 'antd'

import { AuthStatusCard } from '../components/AuthStatusCard'
import { ConcurrentTestPanel } from '../components/ConcurrentTestPanel'
import { DebugStatistics } from '../components/DebugStatistics'
import { DebugTimeline } from '../components/DebugTimeline'

export default function AuthDebugPage() {
  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      <div>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Authentication Debug Panel
        </Typography.Title>
        <Typography.Text type="secondary">
          Kiểm chứng Seamless Authentication (chỉ hiển thị ở development)
        </Typography.Text>
      </div>

      <Alert
        type="info"
        showIcon
        title="Không hiển thị Access Token / Refresh Token. Các test 'Simulated' là mô phỏng cục bộ, không gọi API thật."
      />

      <AuthStatusCard />
      <DebugStatistics />
      <ConcurrentTestPanel />
      <DebugTimeline />
    </Space>
  )
}
