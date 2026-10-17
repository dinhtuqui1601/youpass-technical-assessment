import { ArrowLeftOutlined, ArrowRightOutlined } from '@ant-design/icons'
import { Button, Typography } from 'antd'

interface Props {
  current: number
  total: number
  onPrev: () => void
  onNext: () => void
}

export function ExamActions({ current, total, onPrev, onNext }: Props) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <Button size="large" disabled={current === 0} onClick={onPrev} icon={<ArrowLeftOutlined />}>
        Câu trước
      </Button>
      <Typography.Text type="secondary">
        {current + 1} / {total}
      </Typography.Text>
      <Button
        size="large"
        type="primary"
        disabled={current === total - 1}
        onClick={onNext}
      >
        Câu tiếp theo <ArrowRightOutlined />
      </Button>
    </div>
  )
}

