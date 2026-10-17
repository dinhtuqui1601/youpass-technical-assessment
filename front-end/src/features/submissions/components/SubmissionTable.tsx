import { LinkOutlined } from '@ant-design/icons'
import { App, Button, Space, Table, Typography } from 'antd'
import type { TableProps } from 'antd'
import dayjs from 'dayjs'

import { toApiError } from '@/lib/api-error'
import { useCreateShareLink } from '@/features/share-links/hooks/useShareLinkMutations'
import { getShortUrl } from '@/utils/url'

import type { Submission } from '../types'

function formatContent(content: string): string {
  try {
    return JSON.stringify(JSON.parse(content), null, 2)
  } catch {
    return content
  }
}

interface Props {
  submissions: Submission[]
  loading: boolean
}

export function SubmissionTable({ submissions, loading }: Props) {
  const { message } = App.useApp()
  const createLinkMutation = useCreateShareLink()

  const createLink = async (submission: Submission) => {
    try {
      const link = await createLinkMutation.mutateAsync(submission.id)
      message.success({
        content: (
          <Space>
            <span>Đã tạo link:</span>
            <Typography.Text copyable={{ text: getShortUrl(link.code) }} strong>
              {getShortUrl(link.code)}
            </Typography.Text>
          </Space>
        ),
      })
    } catch (e) {
      message.error(toApiError(e).message)
    }
  }

  const columns: TableProps<Submission>['columns'] = [
    {
      title: 'Tiêu đề',
      dataIndex: 'title',
      key: 'title',
      render: (title: string) => <Typography.Text strong>{title}</Typography.Text>,
    },
    {
      title: 'Submission ID',
      dataIndex: 'id',
      key: 'id',
      width: 250,
      render: (id: string) => (
        <Typography.Text copyable={{ text: id }} style={{ fontFamily: 'monospace', fontSize: 12 }}>
          {id}
        </Typography.Text>
      ),
    },
    {
      title: 'Ngày tạo',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 170,
      render: (v: string) => dayjs(v).format('DD/MM/YYYY HH:mm'),
    },
    {
      title: 'Hành động',
      key: 'action',
      width: 130,
      render: (_, record) => (
        <Button
          size="small"
          type="primary"
          ghost
          icon={<LinkOutlined />}
          loading={createLinkMutation.isPending && createLinkMutation.variables === record.id}
          onClick={() => void createLink(record)}
        >
          Tạo link
        </Button>
      ),
    },
  ]

  return (
    <Table<Submission>
      rowKey="id"
      columns={columns}
      dataSource={submissions}
      loading={loading}
      pagination={{ pageSize: 10, hideOnSinglePage: true }}
      expandable={{
        rowExpandable: () => true,
        expandedRowRender: (record) => (
          <pre style={{ whiteSpace: 'pre-wrap', margin: 0, fontSize: 12 }}>{formatContent(record.content)}</pre>
        ),
      }}
    />
  )
}
