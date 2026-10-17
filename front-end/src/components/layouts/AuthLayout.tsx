import { CheckCircleFilled, LinkOutlined } from '@ant-design/icons'
import { Grid, Space, Typography } from 'antd'
import { Outlet } from 'react-router-dom'

const FEATURES = [
  'Quản lý link chia sẻ bài làm gọn gàng',
  'Đăng nhập liền mạch, không gián đoạn',
  'Thống kê lượt xem theo thời gian thực',
]

export function AuthLayout() {
  const screens = Grid.useBreakpoint()
  const showBrand = screens.md

  return (
    <div style={{ minHeight: '100vh', display: 'flex' }}>
      {showBrand && (
        <div
          style={{
            flex: 1,
            background: 'linear-gradient(160deg, #2563EB 0%, #1D4ED8 100%)',
            color: '#fff',
            padding: 48,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'space-between',
          }}
        >
          <Space size={10}>
            <div
              style={{
                width: 40,
                height: 40,
                borderRadius: 11,
                background: 'rgba(255,255,255,0.16)',
                display: 'grid',
                placeItems: 'center',
              }}
            >
              <LinkOutlined style={{ fontSize: 20 }} />
            </div>
            <Typography.Text strong style={{ color: '#fff', fontSize: 20 }}>
              YouPass
            </Typography.Text>
          </Space>

          <div>
            <Typography.Title level={2} style={{ color: '#fff', marginBottom: 12 }}>
              Học tập. Chia sẻ.
              <br />
              Phát triển.
            </Typography.Title>
            <Typography.Paragraph style={{ color: 'rgba(255,255,255,0.85)', fontSize: 15, maxWidth: 440 }}>
              YouPass giúp học viên nộp bài, tạo link chia sẻ và theo dõi lượt xem một cách
              chuyên nghiệp.
            </Typography.Paragraph>

            <Space orientation="vertical" size={12} style={{ marginTop: 24 }}>
              {FEATURES.map((f) => (
                <Space key={f} size={10}>
                  <CheckCircleFilled style={{ color: '#93C5FD', fontSize: 16 }} />
                  <Typography.Text style={{ color: '#EFF6FF', fontSize: 14 }}>{f}</Typography.Text>
                </Space>
              ))}
            </Space>
          </div>

          <Typography.Text style={{ color: 'rgba(255,255,255,0.6)', fontSize: 12 }}>
            © {new Date().getFullYear()} YouPass
          </Typography.Text>
        </div>
      )}

      <div
        style={{
          flex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '32px 24px',
          background: '#F7F9FC',
        }}
      >
        <Outlet />
      </div>
    </div>
  )
}

