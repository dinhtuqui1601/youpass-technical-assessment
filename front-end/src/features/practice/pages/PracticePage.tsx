import { ClockCircleOutlined } from '@ant-design/icons'
import { App, Button, Card, Col, Progress, Result, Row, Space, Tag, Typography } from 'antd'
import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { useCreateSubmission } from '@/features/submissions/hooks/useSubmissions'

import { ExamActions } from '../components/ExamActions'
import { QuestionCard } from '../components/QuestionCard'
import { QuestionNavigator } from '../components/QuestionNavigator'
import { demoQuestions } from '../data/questions'
import { usePracticeTimer } from '../hooks/usePracticeTimer'
import { usePracticeStore } from '../stores/practice.store'

const EXAM_DURATION_SEC = 15 * 60
const EXAM_TITLE = 'Bài kiểm tra mẫu (demo)'

function formatTime(sec: number): string {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

export default function PracticePage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const { modal, message } = App.useApp()
  const createSubmissionMutation = useCreateSubmission()

  const examId = usePracticeStore((s) => s.examId)
  const startedAt = usePracticeStore((s) => s.startedAt)
  const answers = usePracticeStore((s) => s.answers)
  const currentIndex = usePracticeStore((s) => s.currentIndex)
  const submitted = usePracticeStore((s) => s.submitted)
  const start = usePracticeStore((s) => s.start)
  const setAnswer = usePracticeStore((s) => s.setAnswer)
  const setCurrentIndex = usePracticeStore((s) => s.setCurrentIndex)
  const submit = usePracticeStore((s) => s.submit)

  const { remaining } = usePracticeTimer(startedAt, EXAM_DURATION_SEC)

  // Start only once per exam; a reload or token refresh keeps the draft.
  useEffect(() => {
    if (id === 'demo' && (examId !== 'demo' || !startedAt)) {
      start('demo')
    }
  }, [id, examId, startedAt, start])

  if (id !== 'demo') {
    return (
      <Result
        status="404"
        title="Không tìm thấy bài làm"
        extra={
          <Button type="primary" onClick={() => navigate('/dashboard')}>
            Về Dashboard
          </Button>
        }
      />
    )
  }

  if (!startedAt) {
    return <Card loading />
  }

  if (submitted) {
    const answeredCount = Object.keys(answers).length
    return (
      <Result
        status="success"
        title="Đã nộp bài"
        subTitle={`Bạn đã trả lời ${answeredCount}/${demoQuestions.length} câu. Bài làm đã được lưu lên server.`}
        extra={[
          <Button key="again" onClick={() => start('demo')}>
            Làm lại
          </Button>,
          <Button key="back" type="primary" onClick={() => navigate('/submissions')}>
            Xem bài làm của tôi
          </Button>,
        ]}
      />
    )
  }

  const question = demoQuestions[currentIndex]
  const answeredCount = Object.keys(answers).length
  const unansweredCount = demoQuestions.length - answeredCount
  const timerColor = remaining <= 60 ? '#DC2626' : remaining <= 300 ? '#D97706' : '#2563EB'

  const onConfirmSubmit = () => {
    modal.confirm({
      title: 'Nộp bài?',
      content: `Bạn đã trả lời ${answeredCount}/${demoQuestions.length} câu.${
        unansweredCount > 0 ? ` Còn ${unansweredCount} câu chưa trả lời.` : ''
      } Bài làm sẽ được lưu lên server.`,
      okText: 'Nộp bài',
      cancelText: 'Làm tiếp',
      onOk: async () => {
        const content = JSON.stringify({
          examId: 'demo',
          examTitle: EXAM_TITLE,
          submittedAt: new Date().toISOString(),
          answeredCount,
          total: demoQuestions.length,
          answers: demoQuestions.map((q, i) => ({
            question: q.id,
            text: q.text,
            answer: answers[i] ?? null,
          })),
        })

        await createSubmissionMutation.mutateAsync({ title: EXAM_TITLE, content })
        submit()
        message.success('Đã nộp bài thành công')
        navigate('/submissions')
      },
    })
  }

  return (
    <Row gutter={[24, 24]}>
      <Col xs={24} lg={17}>
        <Space orientation="vertical" size={16} style={{ width: '100%' }}>
          <Card styles={{ body: { padding: '16px 24px' } }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                flexWrap: 'wrap',
                gap: 8,
              }}
            >
              <div>
                <Typography.Title level={4} style={{ margin: 0 }}>
                  {EXAM_TITLE}
                </Typography.Title>
                <Typography.Text type="secondary">
                  {demoQuestions.length} câu hỏi · 15 phút
                </Typography.Text>
              </div>
              <Tag color="blue">
                {answeredCount}/{demoQuestions.length} đã trả lời
              </Tag>
            </div>
          </Card>

          <QuestionCard
            question={question}
            index={currentIndex}
            value={answers[currentIndex]}
            onChange={(key) => setAnswer(currentIndex, key)}
          />

          <ExamActions
            current={currentIndex}
            total={demoQuestions.length}
            onPrev={() => setCurrentIndex(Math.max(0, currentIndex - 1))}
            onNext={() => setCurrentIndex(Math.min(demoQuestions.length - 1, currentIndex + 1))}
          />
        </Space>
      </Col>

      <Col xs={24} lg={7}>
        <div style={{ position: 'sticky', top: 88 }}>
          <Space orientation="vertical" size={16} style={{ width: '100%' }}>
            <Card styles={{ body: { padding: 20 } }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Space size={8}>
                  <ClockCircleOutlined style={{ color: timerColor }} />
                  <Typography.Text type="secondary">Thời gian còn lại</Typography.Text>
                </Space>
                <Typography.Text
                  strong
                  style={{ fontSize: 26, color: timerColor, fontVariantNumeric: 'tabular-nums' }}
                >
                  {formatTime(remaining)}
                </Typography.Text>
              </div>
            </Card>

            <Card styles={{ body: { padding: 20 } }}>
              <Space orientation="vertical" size={8} style={{ width: '100%' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography.Text type="secondary">Tiến độ</Typography.Text>
                  <Typography.Text strong>
                    {answeredCount}/{demoQuestions.length}
                  </Typography.Text>
                </div>
                <Progress
                  percent={Math.round((answeredCount / demoQuestions.length) * 100)}
                  size="small"
                />
              </Space>
            </Card>

            <Card title="Câu hỏi" styles={{ body: { padding: 16 } }}>
              <QuestionNavigator
                total={demoQuestions.length}
                current={currentIndex}
                answers={answers}
                onSelect={setCurrentIndex}
              />
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12, marginTop: 16 }}>
                <Space size={6}>
                  <span style={{ width: 10, height: 10, borderRadius: 3, background: '#2563EB', display: 'inline-block' }} />
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    Đang chọn
                  </Typography.Text>
                </Space>
                <Space size={6}>
                  <span style={{ width: 10, height: 10, borderRadius: 3, background: '#BFDBFE', display: 'inline-block' }} />
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    Đã trả lời
                  </Typography.Text>
                </Space>
                <Space size={6}>
                  <span style={{ width: 10, height: 10, borderRadius: 3, background: '#fff', border: '1px solid #E5EAF0', display: 'inline-block' }} />
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    Chưa làm
                  </Typography.Text>
                </Space>
              </div>
            </Card>

            <Button type="primary" block size="large" onClick={onConfirmSubmit}>
              Nộp bài
            </Button>
          </Space>
        </div>
      </Col>
    </Row>
  )
}
