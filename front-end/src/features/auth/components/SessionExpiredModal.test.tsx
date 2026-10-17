import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import { SessionExpiredModal } from './SessionExpiredModal'
import { useAuthStore } from '../stores/auth.store'

describe('SessionExpiredModal', () => {
  beforeEach(() => {
    useAuthStore.setState({ sessionExpired: false })
  })

  it('does not render when session is active', () => {
    render(<SessionExpiredModal />)
    expect(screen.queryByText('Phiên đăng nhập đã hết hạn')).not.toBeInTheDocument()
  })

  it('renders when sessionExpired is true', () => {
    useAuthStore.setState({ sessionExpired: true })
    render(<SessionExpiredModal />)
    expect(screen.getByText('Phiên đăng nhập đã hết hạn')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Đăng nhập lại' })).toBeInTheDocument()
  })
})
