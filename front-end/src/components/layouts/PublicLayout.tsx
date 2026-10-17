import { Layout, Space, Typography } from 'antd'
import { LinkOutlined } from '@ant-design/icons'
import { Link, Outlet } from 'react-router-dom'
import { theme } from 'antd'

const { Header, Content } = Layout

export function PublicLayout() {
  const { token } = theme.useToken()

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header
        style={{
          display: 'flex',
          alignItems: 'center',
          background: token.colorBgContainer,
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
          paddingInline: 24,
        }}
      >
        <Link to="/" style={{ color: 'inherit', textDecoration: 'none' }}>
          <Space size={10}>
            <div
              style={{
                width: 32,
                height: 32,
                borderRadius: 9,
                background: token.colorPrimary,
                color: '#fff',
                display: 'grid',
                placeItems: 'center',
              }}
            >
              <LinkOutlined style={{ fontSize: 16 }} />
            </div>
            <Typography.Text strong style={{ fontSize: 16 }}>
              YouPass
            </Typography.Text>
          </Space>
        </Link>
      </Header>
      <Content style={{ padding: '32px 24px' }}>
        <div className="yp-page-narrow">
          <Outlet />
        </div>
      </Content>
    </Layout>
  )
}

