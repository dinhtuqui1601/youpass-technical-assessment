import { Card, Typography } from 'antd'

import type { PracticeQuestion } from '../data/questions'

interface Props {
  question: PracticeQuestion
  index: number
  value?: string
  onChange: (key: string) => void
}

export function QuestionCard({ question, index, value, onChange }: Props) {
  return (
    <Card styles={{ body: { padding: 24 } }}>
      <Typography.Text type="secondary" style={{ fontSize: 13 }}>
        Câu {index + 1}
      </Typography.Text>
      <Typography.Title level={5} style={{ marginTop: 4, marginBottom: 24, fontSize: 18 }}>
        {question.text}
      </Typography.Title>

      <div style={{ display: 'grid', gap: 12 }}>
        {question.options.map((opt) => {
          const selected = value === opt.key
          return (
            <div
              key={opt.key}
              role="radio"
              aria-checked={selected}
              tabIndex={0}
              onClick={() => onChange(opt.key)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault()
                  onChange(opt.key)
                }
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 14,
                padding: '14px 16px',
                borderRadius: 10,
                border: '1px solid',
                borderColor: selected ? '#2563EB' : '#E5EAF0',
                background: selected ? '#EFF6FF' : '#FFFFFF',
                boxShadow: selected ? '0 0 0 3px rgba(37, 99, 235, 0.10)' : 'none',
                cursor: 'pointer',
                transition: 'border-color 0.15s ease, background-color 0.15s ease, box-shadow 0.15s ease',
              }}
            >
              <span
                style={{
                  width: 28,
                  height: 28,
                  borderRadius: '50%',
                  border: selected ? 'none' : '1px solid #E5EAF0',
                  background: selected ? '#2563EB' : '#FFFFFF',
                  color: selected ? '#FFFFFF' : '#64748B',
                  display: 'grid',
                  placeItems: 'center',
                  fontWeight: 600,
                  fontSize: 13,
                  flexShrink: 0,
                }}
              >
                {opt.key}
              </span>
              <Typography.Text style={{ fontSize: 16 }}>{opt.text}</Typography.Text>
            </div>
          )
        })}
      </div>
    </Card>
  )
}

