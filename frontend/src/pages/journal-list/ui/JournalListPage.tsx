import { FormEvent, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Edit2, Plus, Trash2 } from 'lucide-react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { journalsApi, type Journal } from '@/entities/journal/api/journals'
import { authApi } from '@/entities/user/api/auth'
import { queryClient } from '@/shared/api/query-client'

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

  if (user.isPending) return <main className="page-loader">Trade Diary</main>
  if (user.isError) return <Navigate to="/login" replace />
  return <main className="workspace">
    <header className="topbar"><div className="brand">Trade Diary</div><div className="account"><span>{user.data.name}</span><button className="text-button" onClick={() => logout.mutate()}>Выйти</button></div></header>
    <section className="content" aria-labelledby="journals-title">
      <div className="section-heading"><div><h1 id="journals-title">Журналы</h1><p>Торговые наблюдения и статистика</p></div><button className="primary-button button-with-icon" onClick={() => setDialog('create')}><Plus size={16}/>Новый журнал</button></div>
      {journals.isPending && <div className="list-status">Загрузка…</div>}
      {journals.isError && <div className="list-status error">Не удалось загрузить журналы</div>}
      {journals.data?.length === 0 && <div className="empty-state"><button className="empty-mark" onClick={() => setDialog('create')} aria-label="Создать журнал"><Plus size={20}/></button><h2>Здесь появятся ваши журналы</h2><p>Создайте первый журнал торговых наблюдений.</p></div>}
      {!!journals.data?.length && <div className="journal-list">{journals.data.map((item) => <div className="journal-row" key={item.id}><Link to={`/journals/${item.id}`}><strong>{item.name}</strong><span>Изменён {new Date(item.updatedAt).toLocaleDateString('ru-RU')}</span></Link><div className="row-actions"><button title="Переименовать" aria-label={`Переименовать ${item.name}`} onClick={() => openRename(item)}><Edit2 size={16}/></button><button title="Удалить" aria-label={`Удалить ${item.name}`} onClick={() => openDelete(item)}><Trash2 size={16}/></button></div></div>)}</div>}
    </section>
    {dialog && <div className="dialog-backdrop" onMouseDown={(e) => e.target === e.currentTarget && close()}><div className="dialog" role="dialog" aria-modal="true" aria-labelledby="dialog-title">{dialog === 'delete' ? <><h2 id="dialog-title">Удалить журнал?</h2><p>«{selected?.name}» будет удалён без возможности восстановления.</p><div className="dialog-actions"><button className="secondary-button" onClick={close}>Отмена</button><button className="danger-button" onClick={() => selected && remove.mutate(selected.id)}>Удалить</button></div></> : <form onSubmit={submit}><h2 id="dialog-title">{dialog === 'create' ? 'Новый журнал' : 'Переименовать журнал'}</h2><label>Название<input autoFocus required maxLength={120} value={name} onChange={(e) => setName(e.target.value)}/></label><div className="dialog-actions"><button type="button" className="secondary-button" onClick={close}>Отмена</button><button className="primary-button" disabled={!name.trim()}>Сохранить</button></div></form>}</div></div>}
  </main>
}
