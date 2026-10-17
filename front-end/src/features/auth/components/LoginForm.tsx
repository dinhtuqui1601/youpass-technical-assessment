import { LockOutlined, MailOutlined } from '@ant-design/icons'
import { zodResolver } from '@hookform/resolvers/zod'
import { Alert, Button, Form, Input, Typography } from 'antd'
import { Controller, useForm } from 'react-hook-form'
import { useLocation, useNavigate } from 'react-router-dom'
import { z } from 'zod'

import { ApiError } from '@/lib/api-error'

import { useLogin } from '../hooks/useLogin'

const schema = z.object({
  email: z.string().email('Email không hợp lệ'),
  password: z.string().min(1, 'Vui lòng nhập mật khẩu'),
})

type FormValues = z.infer<typeof schema>

export function LoginForm() {
  const navigate = useNavigate()
  const location = useLocation()
  const loginMutation = useLogin()

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = (values: FormValues) => {
    loginMutation.mutate(values, {
      onSuccess: () => {
        const state = location.state as { from?: string } | null
        navigate(state?.from ?? '/share-links', { replace: true })
      },
    })
  }

  const errorMessage =
    loginMutation.error instanceof ApiError
      ? loginMutation.error.message
      : loginMutation.error?.message

  return (
    <form onSubmit={handleSubmit(onSubmit)} noValidate>
      {loginMutation.isError && (
        <Alert
          type="error"
          showIcon
          title={errorMessage ?? 'Đăng nhập thất bại'}
          style={{ marginBottom: 20 }}
        />
      )}

      <Form layout="vertical" requiredMark={false} component={false}>
        <Controller
          name="email"
          control={control}
          render={({ field }) => (
            <Form.Item
              label="Email"
              validateStatus={errors.email ? 'error' : undefined}
              help={errors.email?.message}
            >
              <Input
                {...field}
                size="large"
                prefix={<MailOutlined style={{ color: '#94A3B8' }} />}
                placeholder="you@youpass.vn"
                autoComplete="email"
              />
            </Form.Item>
          )}
        />

        <Controller
          name="password"
          control={control}
          render={({ field }) => (
            <Form.Item
              label="Mật khẩu"
              validateStatus={errors.password ? 'error' : undefined}
              help={errors.password?.message}
            >
              <Input.Password
                {...field}
                size="large"
                prefix={<LockOutlined style={{ color: '#94A3B8' }} />}
                placeholder="Nhập mật khẩu"
                autoComplete="current-password"
              />
            </Form.Item>
          )}
        />
      </Form>

      <Button
        type="primary"
        htmlType="submit"
        block
        size="large"
        loading={loginMutation.isPending}
        style={{ marginTop: 4 }}
      >
        Đăng nhập
      </Button>

      <Typography.Text
        type="secondary"
        style={{ display: 'block', textAlign: 'center', marginTop: 20, fontSize: 13 }}
      >
        Bạn quên mật khẩu? Liên hệ quản trị viên để được hỗ trợ.
      </Typography.Text>
    </form>
  )
}
