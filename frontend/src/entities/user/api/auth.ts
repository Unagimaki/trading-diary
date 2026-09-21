import { API_URL } from '@/shared/config/env'

export type User = { id: string; email: string; name: string }
type Credentials = { name?: string; email: string; password: string }

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...options?.headers },
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: 'Ошибка сервера' }))
    throw new Error(body.error)
  }
  return response.status === 204 ? (undefined as T) : response.json()
}

export const authApi = {
  me: () => request<User>('/auth/me'),
  login: (data: Credentials) => request<User>('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  register: (data: Credentials) => request<User>('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
}
