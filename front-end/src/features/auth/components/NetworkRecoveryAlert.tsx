import { useQueryClient } from '@tanstack/react-query'
import { Alert, Button } from 'antd'

import { useAuth } from '../hooks/useAuth'
import { useAuthStore } from '../stores/auth.store'

export function NetworkRecoveryAlert() {
  const { networkError } = useAuth()
  const queryClient = useQueryClient()

  if (!networkError) {
    return null
  }

  const onRetry = () => {
    useAuthStore.getState().setNetworkError(false)
    void queryClient.refetchQueries({ type: 'active' })
  }

  return (
    <Alert
      type="warning"
      showIcon
      title="Mất kết nối"
      description="Không thể kết nối đến máy chủ. Dữ liệu bạn đang nhập vẫn được giữ nguyên."
      action={
        <Button size="small" onClick={onRetry}>
          Thử lại
        </Button>
      }
      style={{ marginBottom: 16 }}
    />
  )
}
