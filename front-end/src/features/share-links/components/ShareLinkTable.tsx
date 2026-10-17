import { PlusOutlined } from '@ant-design/icons'
import { Alert, Button, Empty, Table, Tag, Typography } from 'antd'
import type { TableColumnsType } from 'antd'

import { formatDateTime } from '@/utils/format'
import { getShortUrl } from '@/utils/url'

import type { ShareLink, ShareLinkStatus } from '../types'
import { ShareLinkActions } from './ShareLinkActions'

const statusMeta: Record<ShareLinkStatus, { color: string; label: string }> = {
  ACTIVE: { color: 'green', label: 'Hoạt động' },
  DISABLED: { color: 'orange', label: 'Đã tắt' },
  DELETED: { color: 'red', label: 'Đã xóa' },
}

function StatusTag({ status }: { status: ShareLinkStatus }) {
  const meta = statusMeta[status]
  return <Tag color={meta.color}>{meta.label}</Tag>
}

interface Props {
  links: ShareLink[]
  loading: boolean
  error: boolean
  onRetry: () => void
  viewsByCode?: Record<string, number | undefined>
  onCreate: () => void
}

export function ShareLinkTable({ links, loading, error, onRetry, viewsByCode, onCreate }: Props) {
  if (error) {
    return (
      <Alert
        type="error"
        showIcon
        title="Không thể tải danh sách link chia sẻ"
        action={
          <Button size="small" onClick={onRetry}>
            Thử lại
          </Button>
        }
      />
    )
  }

  const columns: TableColumnsType<ShareLink> = [
    {
      title: 'Bài làm',
      dataIndex: 'submissionId',
      ellipsis: true,
      width: 160,
      render: (v: string) => <Typography.Text code>{v.slice(0, 12)}…</Typography.Text>,
    },
    {
      title: 'Link rút gọn',
      dataIndex: 'code',
      render: (_, record) => (
        <Typography.Text
          copyable={{ text: getShortUrl(record.code), tooltips: ['Sao chép', 'Đã sao chép'] }}
        >
          <Typography.Link
            href={getShortUrl(record.code)}
            target="_blank"
            rel="noreferrer"
            style={{ fontFamily: 'monospace', fontSize: 13 }}
          >
            /s/{record.code}
          </Typography.Link>
        </Typography.Text>
      ),
    },
    {
      title: 'Trạng thái',
      dataIndex: 'status',
      width: 130,
      render: (s: ShareLinkStatus) => <StatusTag status={s} />,
    },
    {
      title: 'Lượt xem',
      key: 'views',
      width: 100,
      align: 'right',
      render: (_, record) => viewsByCode?.[record.code] ?? '—',
    },
    {
      title: 'Ngày tạo',
      dataIndex: 'createdAt',
      width: 160,
      render: formatDateTime,
    },
    {
      title: 'Hết hạn',
      dataIndex: 'expiresAt',
      width: 160,
      render: formatDateTime,
    },
    {
      title: 'Thao tác',
      key: 'actions',
      width: 190,
      render: (_, record) => <ShareLinkActions link={record} />,
    },
  ]

  const empty = (
    <Empty
      image={Empty.PRESENTED_IMAGE_SIMPLE}
      description="Bạn chưa có link chia sẻ nào"
      style={{ padding: '32px 0' }}
    >
      <Button type="primary" icon={<PlusOutlined />} onClick={onCreate}>
        Tạo link chia sẻ đầu tiên
      </Button>
    </Empty>
  )

  return (
    <Table<ShareLink>
      rowKey="id"
      columns={columns}
      dataSource={links}
      loading={loading}
      pagination={{ pageSize: 10, hideOnSinglePage: true }}
      scroll={{ x: 960 }}
      locale={{ emptyText: empty }}
    />
  )
}

