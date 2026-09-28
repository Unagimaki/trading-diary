import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router-dom'
import { queryClient } from '@/shared/api/query-client'
import { router } from './router'
import { Notifications } from '@/widgets/notifications'

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Notifications />
    </QueryClientProvider>
  )
}

