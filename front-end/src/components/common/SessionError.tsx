import { Button, Result } from 'antd'

import { useSessionRestore } from '@/features/auth/hooks/useSessionRestore'

export function SessionError() {
  const { retry } = useSessionRestore()

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Result
        status="warning"
        title="Không thể kết nối máy chủ"
        subTitle="Không thể khôi phục phiên đăng nhập. Kiểm tra kết nối và thử lại."
        extra={
          <Button type="primary" onClick={retry}>
            Thử lại
          </Button>
        }
      />
    </div>
  )
}
