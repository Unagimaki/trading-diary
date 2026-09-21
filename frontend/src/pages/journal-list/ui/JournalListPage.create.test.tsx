import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { journalsApi } from '@/entities/journal/api/journals'
import { JournalListPage } from './JournalListPage'

vi.mock('@/entities/user/api/auth', () => ({ authApi: { me: vi.fn().mockResolvedValue({ id: 'u1', name: 'Developer', email: 'dev@example.com' }), logout: vi.fn() } }))
vi.mock('@/entities/journal/api/journals', () => ({ journalsApi: { list: vi.fn().mockResolvedValue([]), create: vi.fn().mockResolvedValue({ id: 'j1', name: 'Сделки', createdAt: '', updatedAt: '' }), rename: vi.fn(), remove: vi.fn() } }))

describe('JournalListPage', () => {
  beforeEach(() => vi.clearAllMocks())
  it('creates a journal from the dialog', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(<QueryClientProvider client={client}><MemoryRouter><JournalListPage /></MemoryRouter></QueryClientProvider>)
    fireEvent.click(await screen.findByRole('button', { name: 'Новый журнал' }))
    fireEvent.change(screen.getByLabelText('Название'), { target: { value: 'Сделки' } })
    fireEvent.click(screen.getByRole('button', { name: 'Сохранить' }))
    await waitFor(() => expect(vi.mocked(journalsApi.create).mock.calls[0]?.[0]).toBe('Сделки'))
  })
})
