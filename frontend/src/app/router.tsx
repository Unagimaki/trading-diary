import { createBrowserRouter } from 'react-router-dom'
import { AuthPage } from '@/pages/auth'
import { JournalListPage } from '@/pages/journal-list'
import { JournalPage } from '@/pages/journal'

export const router = createBrowserRouter([
  { path: '/', element: <JournalListPage /> },
  { path: '/login', element: <AuthPage mode="login" /> },
  { path: '/register', element: <AuthPage mode="register" /> },
  { path: '/journals/:id', element: <JournalPage /> },
])
