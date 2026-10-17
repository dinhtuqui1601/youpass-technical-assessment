import { BugOutlined, FolderOpenOutlined, RightOutlined } from '@ant-design/icons'
import { Avatar, Button, Card, Col, Row, Space, Tag, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'

import { useAuth } from '@/features/auth/hooks/useAuth'
import { usePracticeStore } from '@/features/practice/stores/practice.store'

export default function StudentDashboardPage() {
  const navigate = useNavigate()
  const { user } = useAuth()

  const practiceExamId = usePracticeStore((s) => s.examId)
  const practiceStartedAt = usePracticeStore((s) => s.startedAt)
  const practiceSubmitted = usePracticeStore((s) => s.submitted)
  const inProgress =
    practiceExamId === 'demo' && practiceStartedAt !== null && !practiceSubmitted

  return (
    <Space orientation="vertical" size={24} style={{ width: '100%' }}>
      <Card styles={{ body: { padding: 24 } }}>
        <Space size={16} wrap>
          <Avatar size={56} style={{ background: '#2563EB', fontSize: 22 }}>
            {user?.name?.[0] ?? 'U'}
          </Avatar>
          <div style={{ flex: 1, minWidth: 220 }}>
            <Typography.Title level={4} style={{ margin: 0 }}>
              Xin chào, {user?.name}!
            </Typography.Title>
            <Typography.Text type="secondary">
              {user?.email} · <Tag color="blue">{user?.role}</Tag>
            </Typography.Text>
            <Typography.Paragraph type="secondary" style={{ marginBottom: 0, marginTop: 4 }}>
              Chọn một bài luyện tập bên dưới để bắt đầu, hoặc xem lại các bài làm đã nộp.
            </Typography.Paragraph>
          </div>
        </Space>
      </Card>

      <div>
        <Typography.Title level={5} style={{ marginBottom: 16 }}>
          Bài luyện tập
        </Typography.Title>
        <Row gutter={[16, 16]}>
          <Col xs={24} md={12} lg={8}>
            <Card hoverable styles={{ body: { padding: 24 } }}>
              <Space orientation="vertical" size={12} style={{ width: '100%' }}>
                <div>
                  <Typography.Text strong style={{ fontSize: 16, display: 'block' }}>
                    Bài kiểm tra mẫu
                  </Typography.Text>
                  <Typography.Text type="secondary">6 câu hỏi · 15 phút</Typography.Text>
                </div>
                <Space size={8}>
                  <Tag color="gold">Demo</Tag>
                  {inProgress && <Tag color="blue">Đang làm</Tag>}
                </Space>
                <Button
                  type="primary"
                  block
                  icon={<RightOutlined />}
                  onClick={() => navigate('/practice/demo')}
                >
                  {inProgress ? 'Tiếp tục làm bài' : 'Bắt đầu làm bài'}
                </Button>
              </Space>
            </Card>
          </Col>

          <Col xs={24} md={12} lg={8}>
            <Card hoverable styles={{ body: { padding: 24 } }}>
              <Space orientation="vertical" size={12} style={{ width: '100%' }}>
                <div>
                  <Typography.Text strong style={{ fontSize: 16, display: 'block' }}>
                    Bài làm của tôi
                  </Typography.Text>
                  <Typography.Text type="secondary">Danh sách bài làm đã nộp lên server</Typography.Text>
                </div>
                <Button block icon={<FolderOpenOutlined />} onClick={() => navigate('/submissions')}>
                  Xem bài làm
                </Button>
              </Space>
            </Card>
          </Col>
        </Row>
      </div>

      {import.meta.env.DEV && (
        <Typography.Text type="secondary" style={{ fontSize: 13 }}>
          <BugOutlined /> Development:{' '}
          <Typography.Link onClick={() => navigate('/auth/debug')}>
            mở Authentication Debug Panel
          </Typography.Link>
        </Typography.Text>
      )}
    </Space>
  )
}

