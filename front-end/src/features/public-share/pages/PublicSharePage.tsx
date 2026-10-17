import { Button, Result, Skeleton } from 'antd'
import { useParams } from 'react-router-dom'

import { ApiError } from '@/lib/api-error'

import { SubmissionViewer } from '../components/SubmissionViewer'
import { usePublicShare } from '../hooks/usePublicShare'
import { useTrackShareView } from '../hooks/useTrackShareView'

interface ErrorView {
  status: '404' | '403' | '500' | 'warning'
  title: string
}

function mapError(code: string | null): ErrorView {
  switch (code) {
    case 'SHARE_LINK_DISABLED':
      return { status: 'warning', title: 'Link chia sẻ đã bị tắt' }
    case 'SHARE_LINK_DELETED':
      return { status: 'warning', title: 'Link chia sẻ đã bị xóa' }
    case 'SHARE_LINK_EXPIRED':
      return { status: 'warning', title: 'Link chia sẻ đã hết hạn' }
    case 'SHARE_LINK_NOT_FOUND':
    case 'NOT_FOUND':
      return { status: '404', title: 'Link không tồn tại' }
    case 'FORBIDDEN':
      return { status: '403', title: 'Bạn không có quyền truy cập' }
    case 'NETWORK_TIMEOUT':
    case 'NETWORK_ERROR':
      return { status: '500', title: 'Lỗi kết nối' }
    default:
      return { status: '500', title: 'Đã xảy ra lỗi' }
  }
}

export default function PublicSharePage() {
  const { code = '' } = useParams()
  const query = usePublicShare(code)

  useTrackShareView(code, query.isSuccess)

  if (query.isLoading) {
    return <Skeleton active paragraph={{ rows: 6 }} />
  }

  if (query.isError) {
    const err = query.error
    const errorCode = err instanceof ApiError ? err.code : 'NETWORK_ERROR'
    const view = mapError(errorCode)
    return (
      <Result
        status={view.status}
        title={view.title}
        subTitle={err instanceof Error ? err.message : undefined}
        extra={
          <Button type="primary" onClick={() => void query.refetch()}>
            Thử lại
          </Button>
        }
      />
    )
  }

  if (!query.data) {
    return null
  }

  return <SubmissionViewer data={query.data} />
}
