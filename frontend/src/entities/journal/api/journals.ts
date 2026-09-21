import { apiRequest } from '@/shared/api/http'
export type Journal = { id: string; name: string; createdAt: string; updatedAt: string }
export const journalsApi = {
  list: () => apiRequest<Journal[]>('/journals'), get: (id: string) => apiRequest<Journal>(`/journals/${id}`),
  create: (name: string) => apiRequest<Journal>('/journals', { method: 'POST', body: JSON.stringify({ name }) }),
  rename: (id: string, name: string) => apiRequest<Journal>(`/journals/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) }),
  remove: (id: string) => apiRequest<void>(`/journals/${id}`, { method: 'DELETE' }),
}
