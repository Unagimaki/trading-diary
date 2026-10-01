import { QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { afterEach, expect, it, vi } from 'vitest'
import { AuthPage } from '@/pages/auth'
import { JournalListPage } from '@/pages/journal-list/ui/JournalListPage'
import { authApi } from '@/entities/user/api/auth'
import { ApiError } from './api-error'
import { queryClient } from './query-client'
import { apiRequest } from './http'

afterEach(() => {
  queryClient.clear()
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

it('stops at login after a cached session receives 401', async () => {
  queryClient.setQueryData(['current-user'], { id: '1', name: 'User', email: 'user@example.com' }, { updatedAt: 1 })
  const me = vi.spyOn(authApi, 'me').mockRejectedValue(new ApiError(401, 'AUTHENTICATION_REQUIRED', 'Unauthorized'))
  render(<QueryClientProvider client={queryClient}><MemoryRouter initialEntries={['/']}><Routes>
    <Route path="/" element={<JournalListPage />} />
    <Route path="/login" element={<AuthPage mode="login" />} />
  </Routes></MemoryRouter></QueryClientProvider>)
  expect(await screen.findByRole('heading', { name: 'Вход' })).toBeInTheDocument()
  expect(me).toHaveBeenCalledTimes(1)
})

it('does not retry unauthorized queries', async () => {
  const request = vi.fn().mockRejectedValue(new ApiError(401, 'AUTHENTICATION_REQUIRED', 'Unauthorized'))
  await expect(queryClient.fetchQuery({ queryKey: ['unauthorized'], queryFn: request })).rejects.toMatchObject({ status: 401 })
  expect(request).toHaveBeenCalledTimes(1)
})

it('includes cookies in the shared HTTP client', async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response('{}', { status: 200 }))
  vi.stubGlobal('fetch', fetchMock)
  await apiRequest('/auth/me')
  expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/auth/me'), expect.objectContaining({ credentials: 'include' }))
})
