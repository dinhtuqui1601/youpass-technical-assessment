import { create } from 'zustand'
import { createJSONStorage, persist } from 'zustand/middleware'

interface PracticeState {
  examId: string
  answers: Record<number, string>
  currentIndex: number
  startedAt: number | null
  submitted: boolean

  start: (examId: string) => void
  setAnswer: (index: number, key: string) => void
  setCurrentIndex: (index: number) => void
  submit: () => void
  reset: () => void
}

// Practice draft state. Persisted to sessionStorage so a reload (or a token
// refresh, which never reloads) keeps the learner's answers. Only non-sensitive
// demo data is stored here — never tokens.
export const usePracticeStore = create<PracticeState>()(
  persist(
    (set) => ({
      examId: '',
      answers: {},
      currentIndex: 0,
      startedAt: null,
      submitted: false,

      start: (examId) =>
        set({ examId, answers: {}, currentIndex: 0, startedAt: Date.now(), submitted: false }),
      setAnswer: (index, key) =>
        set((s) => ({ answers: { ...s.answers, [index]: key } })),
      setCurrentIndex: (index) => set({ currentIndex: index }),
      submit: () => set({ submitted: true }),
      reset: () =>
        set({ examId: '', answers: {}, currentIndex: 0, startedAt: null, submitted: false }),
    }),
    {
      name: 'youpass-practice-draft',
      storage: createJSONStorage(() => sessionStorage),
    },
  ),
)
