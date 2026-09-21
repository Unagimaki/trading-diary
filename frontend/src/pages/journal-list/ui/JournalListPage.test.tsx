import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { AuthPage } from '@/pages/auth'

describe('AuthPage', () => {
  it('shows the login form', () => {
    render(<QueryClientProvider client={new QueryClient()}><MemoryRouter><AuthPage mode="login" /></MemoryRouter></QueryClientProvider>)

    expect(screen.getByRole('heading', { name: 'Вход' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Войти' })).toBeEnabled()
  })
})
