import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it } from 'vitest'

import { tokenManager } from '@/lib/token-manager'
import { server } from '@/test/msw/server'

import { useAuthStore } from './stores/auth.store'
import { LoginForm } from './components/LoginForm'

function renderForm() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <LoginForm />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('LoginForm', () => {
  beforeEach(() => {
    tokenManager.clear()
    useAuthStore.setState({ status: 'unauthenticated', user: null, initError: false, sessionVersion: 0 })
  })

  it('logs in and stores the access token in memory', async () => {
    const user = userEvent.setup()
    renderForm()

    await user.type(screen.getByPlaceholderText('you@youpass.vn'), 'demo@youpass.vn')
    await user.type(screen.getByPlaceholderText('Nhập mật khẩu'), 'password123')
    await user.click(screen.getByRole('button', { name: /đăng nhập/i }))

    await waitFor(() => expect(tokenManager.get()).toBe('test-access-token'))
    expect(useAuthStore.getState().status).toBe('authenticated')
  })

  it('shows the backend error on invalid credentials', async () => {
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json(
          {
            success: false,
            data: null,
            error: { code: 'INVALID_CREDENTIALS', message: 'Sai email hoặc mật khẩu' },
            requestId: 'r',
          },
          { status: 401 },
        ),
      ),
    )

    const user = userEvent.setup()
    renderForm()

    await user.type(screen.getByPlaceholderText('you@youpass.vn'), 'demo@youpass.vn')
    await user.type(screen.getByPlaceholderText('Nhập mật khẩu'), 'wrong-password')
    await user.click(screen.getByRole('button', { name: /đăng nhập/i }))

    expect(await screen.findByText('Sai email hoặc mật khẩu')).toBeInTheDocument()
    expect(tokenManager.get()).toBeNull()
  })
})
