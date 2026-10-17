import { Card, Space, Tag, Typography } from 'antd'

import type { PublicShare } from '../types'

function formatContent(content: string): { pretty: string; isJson: boolean } {
  try {
    return { pretty: JSON.stringify(JSON.parse(content), null, 2), isJson: true }
  } catch {
    return { pretty: content, isJson: false }
  }
}

export function SubmissionViewer({ data }: { data: PublicShare }) {
  const { pretty, isJson } = formatContent(data.content)

  return (
    <Card styles={{ body: { padding: 24 } }}>
      <Space size={8} style={{ marginBottom: 16 }} wrap>
        <Tag color="green">Đang chia sẻ</Tag>
        <Typography.Text type="secondary" style={{ fontSize: 13 }}>
          Mã link: {data.code}
        </Typography.Text>
      </Space>

      <Typography.Title level={3} style={{ marginTop: 0 }}>
        {data.title}
      </Typography.Title>

      {isJson ? (
        <pre
          style={{
            whiteSpace: 'pre-wrap',
            background: '#F8FAFD',
            border: '1px solid #E5EAF0',
            padding: 16,
            borderRadius: 8,
            fontSize: 13,
            margin: 0,
          }}
        >
          {pretty}
        </pre>
      ) : (
        <Typography.Paragraph style={{ whiteSpace: 'pre-wrap', fontSize: 15, marginBottom: 0 }}>
          {data.content}
        </Typography.Paragraph>
      )}
    </Card>
  )
}

