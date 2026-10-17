import { beforeEach, describe, expect, it } from 'vitest'

import { usePracticeStore } from './practice.store'

describe('practice store', () => {
  beforeEach(() => usePracticeStore.getState().reset())

  it('starts an exam and records answers', () => {
    usePracticeStore.getState().start('demo')

    let latest = usePracticeStore.getState()
    expect(latest.startedAt).not.toBeNull()
    expect(latest.examId).toBe('demo')

    latest.setAnswer(0, 'A')
    latest.setAnswer(1, 'C')
    latest.setCurrentIndex(2)

    latest = usePracticeStore.getState()
    expect(latest.answers[0]).toBe('A')
    expect(latest.answers[1]).toBe('C')
    expect(latest.currentIndex).toBe(2)
  })

  it('submit flags the exam as submitted', () => {
    usePracticeStore.getState().start('demo')
    usePracticeStore.getState().submit()
    expect(usePracticeStore.getState().submitted).toBe(true)
  })

  it('start resets previous answers', () => {
    usePracticeStore.getState().start('demo')
    usePracticeStore.getState().setAnswer(0, 'B')

    usePracticeStore.getState().start('demo')
    expect(usePracticeStore.getState().answers).toEqual({})
    expect(usePracticeStore.getState().submitted).toBe(false)
  })
})
