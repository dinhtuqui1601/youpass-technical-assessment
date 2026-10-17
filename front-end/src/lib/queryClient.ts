import { QueryClient } from '@tanstack/react-query'

import { ApiError } from './api-error'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Do not retry client errors (4xx); retry network/5xx up to 2 times.
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status > 0 && error.status < 500) {
          return false
        }
        return failureCount < 2
      },
      staleTime: 30_000,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: false,
    },
  },
})
