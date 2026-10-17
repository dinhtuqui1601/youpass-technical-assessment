import type { ReactNode } from 'react'

import { AntdProvider } from './AntdProvider'
import { QueryProvider } from './QueryProvider'

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <QueryProvider>
      <AntdProvider>{children}</AntdProvider>
    </QueryProvider>
  )
}
