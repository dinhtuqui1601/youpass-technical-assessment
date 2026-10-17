import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, it } from 'vitest'

import { NetworkRecoveryAlert } from './NetworkRecoveryAlert'
import { useAuthStore } from '../stores/auth.store'

function renderWithProvider(ui: ReactNode) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>)
}

describe('NetworkRecoveryAlert', () => {
  beforeEach(() => {
    useAuthStore.setState({ networkError: false })
  })

  it('renders nothing when connected', () => {
    renderWithProvider(<NetworkRecoveryAlert />)
    expect(screen.queryByText('Mất kết nối')).not.toBeInTheDocument()
  })

  it('renders recovery banner when networkError is true', () => {
    useAuthStore.setState({ networkError: true })
    renderWithProvider(<NetworkRecoveryAlert />)
    expect(screen.getByText('Mất kết nối')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Thử lại' })).toBeInTheDocument()
  })
})
