import { PlusOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Space, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'

import { toApiError } from '@/lib/api-error'

import { SubmissionTable } from '../components/SubmissionTable'
import { useSubmissions } from '../hooks/useSubmissions'

export default function SubmissionsPage() {
  const navigate = useNavigate()
  const { data, isPending, isError, error } = useSubmissions()

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      <Card>
        <Space style={{ width: '100%', justifyContent: 'space-between', flexWrap: 'wrap' }} size={16}>
          <div>
            <Typography.Title level={4} style={{ margin: 0 }}>
              Bài làm của tôi
            </Typography.Title>
            <Typography.Text type="secondary">
              Danh sách bài làm đã nộp lên server.
            </Typography.Text>
          </div>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/practice/demo')}>
            Làm bài mới
          </Button>
        </Space>
      </Card>

      {isError ? (
        <Alert
          type="error"
          showIcon
          title="Không thể tải danh sách bài làm"
          description={toApiError(error).message}
        />
      ) : (
        <SubmissionTable submissions={data ?? []} loading={isPending} />
      )}
    </Space>
  )
}
