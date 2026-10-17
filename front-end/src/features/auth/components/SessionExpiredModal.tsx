import { Button, Modal, Result } from 'antd'

import { useAuth } from '../hooks/useAuth'

export function SessionExpiredModal() {
  const { sessionExpired, logout } = useAuth()

  return (
    <Modal open={sessionExpired} closable={false} mask={{ closable: false }} footer={null} centered>
      <Result
        status="warning"
        title="Phiên đăng nhập đã hết hạn"
        subTitle="Vui lòng đăng nhập lại để tiếp tục sử dụng YouPass."
        extra={
          <Button type="primary" onClick={() => void logout()}>
            Đăng nhập lại
          </Button>
        }
      />
    </Modal>
  )
}
