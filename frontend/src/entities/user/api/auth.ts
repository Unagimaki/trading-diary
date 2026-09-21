import { apiRequest } from '@/shared/api/http'

export type User = { id: string; email: string; name: string }
type Credentials = { name?: string; email: string; password: string }

export const authApi = {
  me: () => apiRequest<User>('/auth/me'),
  login: (data: Credentials) => apiRequest<User>('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  register: (data: Credentials) => apiRequest<User>('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  logout: () => apiRequest<void>('/auth/logout', { method: 'POST' }),
}
