import { Button } from 'antd'

interface Props {
  total: number
  current: number
  answers: Record<number, string>
  onSelect: (index: number) => void
}

export function QuestionNavigator({ total, current, answers, onSelect }: Props) {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 8 }}>
      {Array.from({ length: total }, (_, i) => {
        const answered = Boolean(answers[i])
        const isCurrent = i === current
        return (
          <Button
            key={i}
            type={isCurrent ? 'primary' : 'default'}
            onClick={() => onSelect(i)}
            style={{
              height: 40,
              paddingInline: 0,
              ...(!isCurrent && answered
                ? { background: '#EFF6FF', borderColor: '#BFDBFE', color: '#2563EB' }
                : {}),
            }}
          >
            {i + 1}
          </Button>
        )
      })}
    </div>
  )
}

