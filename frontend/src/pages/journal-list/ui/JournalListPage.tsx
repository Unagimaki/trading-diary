import { FormEvent, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { BookOpen, Edit2, LogOut, Plus, Trash2 } from 'lucide-react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { journalsApi, type Journal } from '@/entities/journal/api/journals'
import { authApi } from '@/entities/user/api/auth'
import { queryClient } from '@/shared/api/query-client'
import { Button, IconButton, StateView } from '@/shared/ui'

export function JournalListPage() {
  const navigate = useNavigate()
  const [dialog, setDialog] = useState<'create' | 'rename' | 'delete' | null>(null)
  const [selected, setSelected] = useState<Journal | null>(null)
  const [name, setName] = useState('')
  const user = useQuery({ queryKey: ['current-user'], queryFn: authApi.me, retry: false })
  const journals = useQuery({ queryKey: ['journals'], queryFn: journalsApi.list, enabled: user.isSuccess })
  const close = () => { setDialog(null); setSelected(null); setName('') }
  const refresh = async () => { await queryClient.invalidateQueries({ queryKey: ['journals'] }); close() }
  const create = useMutation({ mutationFn: journalsApi.create, onSuccess: refresh })
  const rename = useMutation({ mutationFn: ({ id, value }: { id: string; value: string }) => journalsApi.rename(id, value), onSuccess: refresh })
  const remove = useMutation({ mutationFn: journalsApi.remove, onSuccess: refresh })
  const logout = useMutation({ mutationFn: authApi.logout, onSuccess: () => { queryClient.clear(); navigate('/login') } })
  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (dialog === 'create') create.mutate(name)
    else if (selected) rename.mutate({ id: selected.id, value: name })
  }
  const openRename = (item: Journal) => { setSelected(item); setName(item.name); setDialog('rename') }
  const openDelete = (item: Journal) => { setSelected(item); setDialog('delete') }

  if (user.isPending) return <main className="page-loader"><span className="brand-mark">TD</span><span>Trade Diary</span></main>
  if (user.isError) return <Navigate to="/login" replace />
  return <main className="workspace">
    <header className="topbar"><div className="brand"><span className="brand-mark">TD</span><span>Trade Diary</span></div><div className="account"><span className="account-name">{user.data.name}</span><Button variant="ghost" icon={<LogOut size={15}/>} onClick={() => logout.mutate()}>Выйти</Button></div></header>
    <section className="content" aria-labelledby="journals-title">
      <div className="section-heading"><div><span className="page-eyebrow">Рабочее пространство</span><h1 id="journals-title">Журналы</h1><p>Торговые наблюдения, сделки и статистика</p></div><Button variant="primary" icon={<Plus size={16}/>} onClick={() => setDialog('create')}>Новый журнал</Button></div>
      {journals.isPending && <StateView kind="loading" title="Загружаем журналы" compact />}
      {journals.isError && <StateView kind="error" title="Не удалось загрузить журналы" description="Проверьте соединение и повторите попытку." compact />}
      {journals.data?.length === 0 && <StateView kind="empty" title="Журналов пока нет" description="Создайте первый журнал, настройте колонки и начните фиксировать сделки." action={<Button variant="primary" icon={<Plus size={16}/>} onClick={() => setDialog('create')}>Создать журнал</Button>} />}
      {!!journals.data?.length && <div className="journal-list">{journals.data.map((item) => <div className="journal-row" key={item.id}><Link to={`/journals/${item.id}`}><span className="journal-icon"><BookOpen size={17}/></span><span className="journal-row-copy"><strong>{item.name}</strong><small>Изменён {new Date(item.updatedAt).toLocaleDateString('ru-RU')}</small></span></Link><div className="row-actions"><IconButton label={`Переименовать ${item.name}`} onClick={() => openRename(item)}><Edit2 size={15}/></IconButton><IconButton label={`Удалить ${item.name}`} tone="danger" onClick={() => openDelete(item)}><Trash2 size={15}/></IconButton></div></div>)}</div>}
    </section>
    {dialog && <div className="dialog-backdrop" onMouseDown={(e) => e.target === e.currentTarget && close()}><div className="dialog" role="dialog" aria-modal="true" aria-labelledby="dialog-title">{dialog === 'delete' ? <><h2 id="dialog-title">Удалить журнал?</h2><p>«{selected?.name}» будет удалён без возможности восстановления.</p><div className="dialog-actions"><Button onClick={close}>Отмена</Button><Button variant="danger" loading={remove.isPending} onClick={() => selected && remove.mutate(selected.id)}>Удалить</Button></div></> : <form onSubmit={submit}><h2 id="dialog-title">{dialog === 'create' ? 'Новый журнал' : 'Переименовать журнал'}</h2><label>Название<input autoFocus required maxLength={120} value={name} onChange={(e) => setName(e.target.value)}/></label><div className="dialog-actions"><Button type="button" onClick={close}>Отмена</Button><Button variant="primary" loading={create.isPending || rename.isPending} disabled={!name.trim()}>Сохранить</Button></div></form>}</div></div>}
  </main>
}
