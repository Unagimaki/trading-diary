import { FormEvent, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { authApi } from '@/entities/user/api/auth'
import { queryClient } from '@/shared/api/query-client'

export function AuthPage({ mode }: { mode: 'login' | 'register' }) {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const mutation = useMutation({
    mutationFn: mode === 'login' ? authApi.login : authApi.register,
    onSuccess: (user) => { queryClient.setQueryData(['current-user'], user); navigate('/') },
  })
  const submit = (event: FormEvent) => {
    event.preventDefault()
    mutation.mutate({ name, email, password })
  }

  if (queryClient.getQueryData(['current-user'])) return <Navigate to="/" replace />

  return (
    <main className="auth-page">
      <div className="auth-brand">Trade Diary</div>
      <form className="auth-form" onSubmit={submit}>
        <h1>{mode === 'login' ? 'Вход' : 'Создать аккаунт'}</h1>
        <p>{mode === 'login' ? 'Войдите в свой торговый дневник' : 'Начните вести журнал наблюдений'}</p>
        {mode === 'register' && <label>Имя<input required autoComplete="name" value={name} onChange={(e) => setName(e.target.value)} /></label>}
        <label>Email<input required type="email" autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} /></label>
        <label>Пароль<input required minLength={8} type="password" autoComplete={mode === 'login' ? 'current-password' : 'new-password'} value={password} onChange={(e) => setPassword(e.target.value)} /></label>
        {mutation.error && <div className="form-error">Проверьте введённые данные</div>}
        <button className="primary-button auth-submit" disabled={mutation.isPending}>{mutation.isPending ? 'Подождите…' : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}</button>
        <div className="auth-switch">{mode === 'login' ? 'Нет аккаунта?' : 'Уже есть аккаунт?'} <Link to={mode === 'login' ? '/register' : '/login'}>{mode === 'login' ? 'Регистрация' : 'Войти'}</Link></div>
      </form>
    </main>
  )
}

