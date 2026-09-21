import { useMutation, useQuery } from '@tanstack/react-query'
import { Navigate, useNavigate } from 'react-router-dom'
import { authApi } from '@/entities/user/api/auth'
import { queryClient } from '@/shared/api/query-client'

export function JournalListPage() {
  const navigate = useNavigate()
  const user = useQuery({ queryKey: ['current-user'], queryFn: authApi.me, retry: false })
  const logout = useMutation({ mutationFn: authApi.logout, onSuccess: () => { queryClient.clear(); navigate('/login') } })

  if (user.isPending) return <main className="page-loader">Trade Diary</main>
  if (user.isError) return <Navigate to="/login" replace />

  return (
    <main className="workspace">
      <header className="topbar">
        <div className="brand">Trade Diary</div>
        <div className="account"><span>{user.data.name}</span><button className="text-button" type="button" onClick={() => logout.mutate()}>Выйти</button></div>
      </header>
      <section className="content" aria-labelledby="journals-title">
        <div className="section-heading">
          <div>
            <h1 id="journals-title">Журналы</h1>
            <p>Торговые наблюдения и статистика</p>
          </div>
        </div>
        <div className="empty-state">
          <button className="empty-mark" type="button" aria-label="Создать журнал">+</button>
          <h2>Здесь появятся ваши журналы</h2>
          <p>Создайте первый журнал торговых наблюдений.</p>
        </div>
      </section>
    </main>
  )
}
