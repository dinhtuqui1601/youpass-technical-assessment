import { App as AntdApp, ConfigProvider } from 'antd'
import type { ReactNode } from 'react'

import { themeConfig } from '../theme'

export function AntdProvider({ children }: { children: ReactNode }) {
  return (
    <ConfigProvider theme={themeConfig}>
      <AntdApp>{children}</AntdApp>
    </ConfigProvider>
  )
}
