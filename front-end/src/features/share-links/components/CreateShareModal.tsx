import { CopyOutlined, ExportOutlined } from '@ant-design/icons'
import { Alert, App, Button, Form, Input, Modal, Segmented, Space, Tag, Typography } from 'antd'
import { useEffect, useState } from 'react'

import { toApiError } from '@/lib/api-error'
import { copyText } from '@/utils/clipboard'
import { getShortUrl } from '@/utils/url'

import { useCreateShareLink, useCreateSubmission } from '../hooks/useShareLinkMutations'

type Mode = 'new-submission' | 'existing-id'

interface Props {
  open: boolean
  onClose: () => void
}

export function CreateShareModal({ open, onClose }: Props) {
  const { message } = App.useApp()
  const [form] = Form.useForm()
  const [mode, setMode] = useState<Mode>('new-submission')
  const [createdUrl, setCreatedUrl] = useState<string | null>(null)

  const createSubmissionMutation = useCreateSubmission()
  const createLinkMutation = useCreateShareLink()

  const creating = createSubmissionMutation.isPending || createLinkMutation.isPending

  useEffect(() => {
    if (open) {
      form.resetFields()
      setCreatedUrl(null)
      setMode('new-submission')
    }
  }, [open, form])

  const handleCreate = async () => {
    try {
      const values = await form.validateFields()

      let submissionId: string
      if (mode === 'new-submission') {
        const submission = await createSubmissionMutation.mutateAsync({
          title: values.title,
          content: values.content,
        })
        submissionId = submission.id
      } else {
        submissionId = (values.submissionId as string).trim()
      }

      const link = await createLinkMutation.mutateAsync(submissionId)
      setCreatedUrl(getShortUrl(link.code))
      message.success('Đã tạo link chia sẻ')
    } catch (err) {
      if (err instanceof Error && err.message !== '') {
        message.error(toApiError(err).message)
      }
    }
  }

  const onCopy = async () => {
    if (!createdUrl) return
    const ok = await copyText(createdUrl)
    if (ok) message.success('Đã sao chép link')
    else message.error('Không thể sao chép link')
  }

  return (
    <Modal
      title="Tạo link chia sẻ"
      open={open}
      onCancel={onClose}
      width={560}
      destroyOnHidden
      footer={
        createdUrl ? (
          <Button type="primary" onClick={onClose}>
            Đóng
          </Button>
        ) : (
          <Space>
            <Button onClick={onClose}>Hủy</Button>
            <Button type="primary" loading={creating} onClick={() => void handleCreate()}>
              Tạo link
            </Button>
          </Space>
        )
      }
    >
      {createdUrl ? (
        <Space orientation="vertical" size={16} style={{ width: '100%' }}>
          <Alert
            type="success"
            showIcon
            title="Đã tạo link chia sẻ thành công"
            description="Link của bạn đã sẵn sàng để chia sẻ."
          />
          <div>
            <Typography.Text type="secondary">Link rút gọn</Typography.Text>
            <div style={{ marginTop: 4 }}>
              <Typography.Text strong copyable={{ text: createdUrl }} style={{ fontSize: 16 }}>
                {createdUrl}
              </Typography.Text>
            </div>
          </div>
          <Space wrap>
            <Button type="primary" icon={<CopyOutlined />} onClick={onCopy}>
              Sao chép
            </Button>
            <Button icon={<ExportOutlined />} href={createdUrl} target="_blank" rel="noreferrer">
              Mở link
            </Button>
            <Tag color="green">Hoạt động</Tag>
          </Space>
        </Space>
      ) : (
        <>
          <Typography.Paragraph type="secondary" style={{ marginTop: 0 }}>
            Chọn nguồn bài làm để tạo link chia sẻ rút gọn.
          </Typography.Paragraph>

          <Form form={form} layout="vertical" requiredMark={false}>
            <Form.Item label="Nguồn bài làm">
              <Segmented
                value={mode}
                onChange={(v) => setMode(v as Mode)}
                style={{ width: '100%' }}
                options={[
                  { label: 'Tạo bài làm mới', value: 'new-submission' },
                  { label: 'Nhập Submission ID', value: 'existing-id' },
                ]}
              />
            </Form.Item>

            {mode === 'new-submission' ? (
              <>
                <Form.Item
                  name="title"
                  label="Tiêu đề"
                  rules={[{ required: true, message: 'Vui lòng nhập tiêu đề' }]}
                >
                  <Input placeholder="Tiêu đề bài làm" maxLength={200} />
                </Form.Item>
                <Form.Item
                  name="content"
                  label="Nội dung"
                  rules={[{ required: true, message: 'Vui lòng nhập nội dung' }]}
                >
                  <Input.TextArea rows={4} placeholder="Nội dung bài làm" />
                </Form.Item>
              </>
            ) : (
              <Form.Item
                name="submissionId"
                label="Submission ID"
                rules={[{ required: true, message: 'Vui lòng nhập Submission ID' }]}
                extra="Nhập ID của một bài làm đã có (xem tại mục Bài làm)."
              >
                <Input placeholder="Ví dụ: 665f1a2b3c4d5e6f7a8b9c0d" />
              </Form.Item>
            )}
          </Form>
        </>
      )}
    </Modal>
  )
}

