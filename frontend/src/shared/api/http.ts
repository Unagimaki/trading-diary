import { API_URL } from '@/shared/config/env'

export async function apiRequest<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, { ...options, credentials: 'include', headers: { 'Content-Type': 'application/json', ...options?.headers } })
  if (!response.ok) { const body = await response.json().catch(() => ({ error: 'Ошибка сервера' })); throw new Error(body.error) }
  return response.status === 204 ? (undefined as T) : response.json()
}
