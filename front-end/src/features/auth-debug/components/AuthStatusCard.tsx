import { Badge, Card, Descriptions, Tag } from 'antd'
import dayjs from 'dayjs'

import { useAuth } from '@/features/auth/hooks/useAuth'
import { useTokenStatus } from '../hooks/useTokenStatus'

const tokenStateColor: Record<string, string> = {
  none: 'default',
  valid: 'green',
  expired: 'red',
  refreshing: 'processing',
}

export function AuthStatusCard() {
  const { status, user } = useAuth()
  const { state, expiresAt, remainingMs } = useTokenStatus()

  const remainingText =
    remainingMs === null ? '—' : `${Math.floor(remainingMs / 1000)}s`

  return (
    <Card title="Session Overview">
      <Descriptions column={{ xs: 1, md: 2 }} items={[
        {
          key: 'auth',
          label: 'Authentication Status',
          children: (
            <Badge
              status={status === 'authenticated' ? 'success' : status === 'initializing' ? 'processing' : 'error'}
              text={status}
            />
          ),
        },
        {
          key: 'token',
          label: 'Access Token State',
          children: <Tag color={tokenStateColor[state]}>{state}</Tag>,
        },
        {
          key: 'session',
          label: 'Session Status',
          children: user ? 'Active' : 'No session',
        },
        {
          key: 'expiresAt',
          label: 'Token Expiration Time',
          children: expiresAt ? dayjs(expiresAt).format('HH:mm:ss') : '—',
        },
        {
          key: 'remaining',
          label: 'Remaining Time',
          children: remainingText,
        },
        {
          key: 'user',
          label: 'User',
          children: user ? `${user.name} (${user.email})` : '—',
        },
      ]} />
    </Card>
  )
}
