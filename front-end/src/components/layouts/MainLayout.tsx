import {
  AppstoreOutlined,
  BugOutlined,
  FileTextOutlined,
  LinkOutlined,
  LogoutOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { Avatar, Breadcrumb, Dropdown, Layout, Menu, Space, Typography, theme } from 'antd'
import type { MenuProps } from 'antd'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'

import { NetworkRecoveryAlert } from '@/features/auth/components/NetworkRecoveryAlert'
import { SessionExpiredModal } from '@/features/auth/components/SessionExpiredModal'
import { useAuth } from '@/features/auth/hooks/useAuth'

const { Sider, Header, Content } = Layout

const pageTitles: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/submissions': 'Bài làm của tôi',
  '/share-links': 'Quản lý link chia sẻ',
  '/account': 'Tài khoản',
  '/auth/debug': 'Authentication Debug',
}

export function MainLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()
  const { token } = theme.useToken()

  const selectedKey = location.pathname.startsWith('/share-links')
    ? '/share-links'
    : location.pathname.startsWith('/submissions') || location.pathname.startsWith('/practice')
      ? '/submissions'
      : location.pathname === '/account'
        ? '/account'
        : location.pathname === '/auth/debug'
          ? '/auth/debug'
          : '/dashboard'

  const title = location.pathname.startsWith('/share-links/')
    ? 'Chi tiết link chia sẻ'
    : location.pathname.startsWith('/practice')
      ? 'Làm bài'
      : (pageTitles[selectedKey] ?? 'Dashboard')

  const menuItems: MenuProps['items'] = [
    { key: '/dashboard', icon: <AppstoreOutlined />, label: 'Dashboard' },
    { key: '/submissions', icon: <FileTextOutlined />, label: 'Bài làm' },
    { key: '/share-links', icon: <LinkOutlined />, label: 'Link chia sẻ' },
    { key: '/account', icon: <UserOutlined />, label: 'Tài khoản' },
    ...(import.meta.env.DEV
      ? [{ key: '/auth/debug', icon: <BugOutlined />, label: 'Authentication Debug' }]
      : []),
  ]

  const userMenu: MenuProps = {
    items: [
      {
        key: 'info',
        icon: <UserOutlined />,
        label: (
          <Space orientation="vertical" size={0}>
            <Typography.Text strong>{user?.name}</Typography.Text>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {user?.email}
            </Typography.Text>
          </Space>
        ),
        disabled: true,
      },
      { type: 'divider' },
      { key: 'account', label: 'Tài khoản', onClick: () => navigate('/account') },
      {
        key: 'logout',
        icon: <LogoutOutlined />,
        label: 'Đăng xuất',
        onClick: () => {
          void logout()
        },
      },
    ],
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        breakpoint="lg"
        collapsedWidth={0}
        width={240}
        style={{
          background: token.colorBgContainer,
          borderRight: `1px solid ${token.colorBorderSecondary}`,
          display: 'flex',
          flexDirection: 'column',
          position: 'sticky',
          top: 0,
          height: '100vh',
          overflow: 'auto',
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            paddingInline: 20,
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            flexShrink: 0,
          }}
        >
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: 10,
              background: token.colorPrimary,
              color: '#fff',
              display: 'grid',
              placeItems: 'center',
              flexShrink: 0,
            }}
          >
            <LinkOutlined style={{ fontSize: 18 }} />
          </div>
          <div style={{ lineHeight: 1.2 }}>
            <Typography.Text strong style={{ fontSize: 16, display: 'block' }}>
              YouPass
            </Typography.Text>
            <Typography.Text type="secondary" style={{ fontSize: 11 }}>
              Learn &amp; Share
            </Typography.Text>
          </div>
        </div>

        <Typography.Text
          type="secondary"
          style={{
            fontSize: 11,
            letterSpacing: '0.08em',
            textTransform: 'uppercase',
            paddingInline: 24,
            marginTop: 16,
            display: 'block',
          }}
        >
          Điều hướng
        </Typography.Text>

        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ flex: 1, borderInlineEnd: 'none', marginTop: 4 }}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            paddingInline: 24,
            background: token.colorBgContainer,
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            position: 'sticky',
            top: 0,
            zIndex: 10,
          }}
        >
          <Breadcrumb items={[{ title: 'YouPass' }, { title }]} />
          <Dropdown menu={userMenu} trigger={['click']}>
            <Space style={{ cursor: 'pointer' }}>
              <Avatar size={32} style={{ background: token.colorPrimary }}>
                {user?.name?.[0] ?? <UserOutlined />}
              </Avatar>
              <span>{user?.name ?? user?.email}</span>
            </Space>
          </Dropdown>
        </Header>

        <Content style={{ padding: '24px 24px 40px' }}>
          <div className="yp-page">
            <NetworkRecoveryAlert />
            <Outlet />
            <SessionExpiredModal />
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}
