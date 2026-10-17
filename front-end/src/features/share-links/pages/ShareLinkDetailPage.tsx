import { ArrowLeftOutlined, CopyOutlined, DeleteOutlined } from '@ant-design/icons'
import { Alert, App, Button, Card, Popconfirm, Result, Space, Switch, Tag, Typography } from 'antd'
import { useNavigate, useParams } from 'react-router-dom'

import { toApiError } from '@/lib/api-error'
import { copyText } from '@/utils/clipboard'
import { getShortUrl } from '@/utils/url'

import { ShareStats } from '../components/ShareStats'
import { useDeleteShareLink, useUpdateShareLinkStatus } from '../hooks/useShareLinkMutations'
import { useShareLinkDetail, useShareLinkStats } from '../hooks/useShareLinkStats'

export default function ShareLinkDetailPage() {
  const { code = '' } = useParams()
  const navigate = useNavigate()
  const { message } = App.useApp()

  const detailQuery = useShareLinkDetail(code)
  const statsQuery = useShareLinkStats(code)

  const updateMutation = useUpdateShareLinkStatus(code)
  const deleteMutation = useDeleteShareLink()

  if (detailQuery.isLoading) {
    return <Card loading />
  }

  if (detailQuery.isError) {
    return (
      <Result
        status="404"
        title="Không tìm thấy link"
        extra={
          <Button type="primary" onClick={() => navigate('/share-links')}>
            Quay lại danh sách
          </Button>
        }
      />
    )
  }

  const link = detailQuery.data
  if (!link) return null

  const onCopy = async () => {
    const ok = await copyText(getShortUrl(link.code))
    if (ok) message.success('Đã sao chép link')
    else message.error('Không thể sao chép link')
  }

  const onToggle = (checked: boolean) => {
    updateMutation.mutate(checked ? 'ACTIVE' : 'DISABLED', {
      onSuccess: (updated) => {
        message.success(updated.status === 'ACTIVE' ? 'Đã bật link chia sẻ' : 'Đã tắt link chia sẻ')
      },
      onError: (e) => message.error(toApiError(e).message),
    })
  }

  const onDelete = () => {
    deleteMutation.mutate(code, {
      onSuccess: () => {
        message.success('Đã xóa link')
        navigate('/share-links')
      },
      onError: (e) => message.error(toApiError(e).message),
    })
  }

  return (
    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
      <Space style={{ width: '100%', justifyContent: 'space-between', flexWrap: 'wrap' }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/share-links')}>
          Danh sách
        </Button>
        <Space wrap>
          <Button icon={<CopyOutlined />} onClick={onCopy}>
            Sao chép link
          </Button>

          {link.status === 'DELETED' ? (
            <Tag color="red">Đã xóa</Tag>
          ) : (
            <>
              <Space>
                <Typography.Text type="secondary">
                  {link.status === 'ACTIVE' ? 'Đang bật' : 'Đang tắt'}
                </Typography.Text>
                <Switch
                  checked={link.status === 'ACTIVE'}
                  loading={updateMutation.isPending}
                  onChange={onToggle}
                />
              </Space>
              <Popconfirm
                title="Xóa link này?"
                description="Link đã xóa không thể khôi phục."
                okText="Xóa"
                okButtonProps={{ danger: true }}
                cancelText="Hủy"
                onConfirm={onDelete}
              >
                <Button danger icon={<DeleteOutlined />} loading={deleteMutation.isPending}>
                  Xóa
                </Button>
              </Popconfirm>
            </>
          )}
        </Space>
      </Space>

      <Card>
        <Typography.Title level={4} style={{ marginTop: 0 }}>
          Thống kê link chia sẻ
        </Typography.Title>
        <Typography.Paragraph style={{ marginBottom: 0 }}>
          Short URL:{' '}
          <a href={getShortUrl(link.code)} target="_blank" rel="noreferrer">
            {getShortUrl(link.code)}
          </a>
        </Typography.Paragraph>
      </Card>

      {statsQuery.isError ? (
        <Alert
          type="warning"
          showIcon
          title="Không thể tải thống kê"
          action={
            <Button size="small" onClick={() => void statsQuery.refetch()}>
              Thử lại
            </Button>
          }
        />
      ) : null}

      <ShareStats link={link} stats={statsQuery.data} loading={statsQuery.isLoading} />
    </Space>
  )
}
