import { CopyOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons'
import { App, Button, Popconfirm, Space, Switch, Tooltip } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { toApiError } from '@/lib/api-error'
import { copyText } from '@/utils/clipboard'
import { getShortUrl } from '@/utils/url'

import { useDeleteShareLink, useUpdateShareLinkStatus } from '../hooks/useShareLinkMutations'
import type { ShareLink } from '../types'

export function ShareLinkActions({ link }: { link: ShareLink }) {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [confirming, setConfirming] = useState(false)

  const updateMutation = useUpdateShareLinkStatus(link.code)
  const deleteMutation = useDeleteShareLink()

  const onCopy = async () => {
    const ok = await copyText(getShortUrl(link.code))
    if (ok) {
      message.success('Đã sao chép link')
    } else {
      message.error('Không thể sao chép link')
    }
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
    deleteMutation.mutate(link.code, {
      onSuccess: () => message.success('Đã xóa link'),
      onError: (e) => message.error(toApiError(e).message),
      onSettled: () => setConfirming(false),
    })
  }

  return (
    <Space size={2} onClick={(e) => e.stopPropagation()}>
      <Tooltip title="Sao chép link">
        <Button size="small" type="text" icon={<CopyOutlined />} onClick={onCopy} />
      </Tooltip>
      <Tooltip title="Thống kê">
        <Button
          size="small"
          type="text"
          icon={<EyeOutlined />}
          onClick={() => navigate(`/share-links/${link.code}`)}
        />
      </Tooltip>

      {link.status !== 'DELETED' ? (
        <>
          <Tooltip title={link.status === 'ACTIVE' ? 'Tắt chia sẻ' : 'Bật chia sẻ'}>
            <Switch
              checked={link.status === 'ACTIVE'}
              loading={updateMutation.isPending}
              onChange={onToggle}
              size="small"
            />
          </Tooltip>
          <Popconfirm
            title="Xóa link này?"
            description="Link đã xóa không thể khôi phục."
            okText="Xóa"
            okButtonProps={{ danger: true }}
            cancelText="Hủy"
            open={confirming}
            onOpenChange={setConfirming}
            onConfirm={onDelete}
          >
            <Button
              size="small"
              type="text"
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
            />
          </Popconfirm>
        </>
      ) : null}
    </Space>
  )
}
