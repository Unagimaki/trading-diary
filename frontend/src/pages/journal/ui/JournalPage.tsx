import { FormEvent, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  ImagePlus,
  Plus,
  Settings2,
  Trash2,
  X,
} from "lucide-react";
import { Link, Navigate, useParams } from "react-router-dom";
import {
  columnsApi,
  type ColumnRole,
  type ColumnType,
  type ColumnValues,
  type JournalColumn,
} from "@/entities/column/api/columns";
import {
  attachmentsApi,
  type Attachment,
} from "@/entities/attachment/api/attachments";
import { journalsApi } from "@/entities/journal/api/journals";
import { rowsApi } from "@/entities/observation/api/rows";
import { queryClient } from "@/shared/api/query-client";

const typeLabels: Record<ColumnType, string> = {
  text: "Текст",
  number: "Число",
  select: "Выбор",
  date: "Дата",
  boolean: "Да / нет",
  image: "Изображение",
};
const roleLabels: Record<Exclude<ColumnRole, null>, string> = {
  trade_result: "Результат сделки",
  pnl: "PnL",
  r: "R-множитель",
};
const compatibleRoles = (type: ColumnType): Exclude<ColumnRole, null>[] =>
  type === "select" ? ["trade_result"] : type === "number" ? ["pnl", "r"] : [];

export function JournalPage() {
  const { id = "" } = useParams();
  const [editing, setEditing] = useState<JournalColumn | null | undefined>(
    undefined,
  );
  const [deleting, setDeleting] = useState<JournalColumn | null>(null);
  const journal = useQuery({
    queryKey: ["journal", id],
    queryFn: () => journalsApi.get(id),
    retry: false,
  });
  const columns = useQuery({
    queryKey: ["columns", id],
    queryFn: () => columnsApi.list(id),
    enabled: journal.isSuccess,
  });
  const rows = useQuery({
    queryKey: ["rows", id],
    queryFn: () => rowsApi.list(id),
    enabled: journal.isSuccess,
  });
  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: ["columns", id] });
  const save = useMutation({
    mutationFn: (data: ColumnValues) =>
      editing
        ? columnsApi.update(id, editing.id, data)
        : columnsApi.create(id, data),
    onSuccess: async () => {
      await refresh();
      setEditing(undefined);
    },
  });
  const remove = useMutation({
    mutationFn: (columnId: string) => columnsApi.remove(id, columnId),
    onSuccess: async () => {
      await refresh();
      setDeleting(null);
    },
  });
  const move = useMutation({
    mutationFn: ({
      item,
      position,
    }: {
      item: JournalColumn;
      position: number;
    }) =>
      columnsApi.update(id, item.id, {
        name: item.name,
        type: item.type,
        role: item.role,
        options: item.options,
        position,
      }),
    onSuccess: refresh,
  });
  const refreshRows = () =>
    queryClient.invalidateQueries({ queryKey: ["rows", id] });
  const addRow = useMutation({
    mutationFn: () => rowsApi.create(id),
    onSuccess: refreshRows,
  });
  const deleteRow = useMutation({
    mutationFn: (rowId: string) => rowsApi.remove(id, rowId),
    onSuccess: refreshRows,
  });
  const saveCell = useMutation({
    mutationFn: ({
      rowId,
      columnId,
      value,
    }: {
      rowId: string;
      columnId: string;
      value: unknown;
    }) => rowsApi.setCell(id, rowId, columnId, value),
    onSuccess: refreshRows,
  });
  if (journal.isError) return <Navigate to="/" replace />;
  return (
    <main className="workspace">
      <header className="topbar">
        <Link className="back-link" to="/">
          <ArrowLeft size={17} />
          Журналы
        </Link>
        <div className="brand">Trade Diary</div>
        <span />
      </header>
      <section className="journal-content">
        <div className="journal-heading">
          <div>
            <h1>{journal.data?.name ?? "Загрузка…"}</h1>
            <p>Торговые наблюдения</p>
          </div>
          <div className="journal-actions">
            <button
              className="secondary-button button-with-icon"
              onClick={() => addRow.mutate()}
              disabled={!columns.data?.length}
            >
              <Plus size={16} />
              Строка
            </button>
            <button
              className="primary-button button-with-icon"
              onClick={() => setEditing(null)}
            >
              <Plus size={16} />
              Колонка
            </button>
          </div>
        </div>
        {columns.isPending && <div className="list-status">Загрузка…</div>}
        {columns.isError && (
          <div className="list-status error">Не удалось загрузить колонки</div>
        )}
        {columns.data?.length === 0 && (
          <div className="empty-state journal-empty">
            <button
              className="empty-mark"
              onClick={() => setEditing(null)}
              aria-label="Добавить колонку"
            >
              <Plus size={20} />
            </button>
            <h2>Настройте структуру журнала</h2>
            <p>Добавьте характеристики, которые хотите фиксировать.</p>
          </div>
        )}
        {!!columns.data?.length && (
          <div className="table-shell">
            <table className="journal-table">
              <thead>
                <tr>
                  <th className="row-number">#</th>
                  {columns.data.map((item, index) => (
                    <th key={item.id}>
                      <div className="column-header">
                        <button
                          className="column-name"
                          onClick={() => setEditing(item)}
                        >
                          <span>{item.name}</span>
                          <small>
                            {typeLabels[item.type]}
                            {item.role ? ` · ${roleLabels[item.role]}` : ""}
                          </small>
                        </button>
                        <div className="column-tools">
                          <button
                            title="Переместить влево"
                            disabled={index === 0}
                            onClick={() =>
                              move.mutate({ item, position: index - 1 })
                            }
                          >
                            <ChevronLeft size={14} />
                          </button>
                          <button
                            title="Переместить вправо"
                            disabled={index === columns.data.length - 1}
                            onClick={() =>
                              move.mutate({ item, position: index + 1 })
                            }
                          >
                            <ChevronRight size={14} />
                          </button>
                          <button
                            title="Настроить"
                            onClick={() => setEditing(item)}
                          >
                            <Settings2 size={14} />
                          </button>
                          <button
                            title="Удалить"
                            onClick={() => setDeleting(item)}
                          >
                            <Trash2 size={14} />
                          </button>
                        </div>
                      </div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows.data?.map((row, index) => (
                  <tr key={row.id}>
                    <td className="row-number row-control">
                      <span>{index + 1}</span>
                      <button
                        title="Удалить строку"
                        onClick={() => deleteRow.mutate(row.id)}
                      >
                        <Trash2 size={13} />
                      </button>
                    </td>
                    {columns.data.map((item) => (
                      <td key={item.id}>
                        <CellEditor
                          journalId={id}
                          rowId={row.id}
                          column={item}
                          value={row.values[item.id]}
                          onRefresh={refreshRows}
                          onSave={(value) =>
                            saveCell.mutate({
                              rowId: row.id,
                              columnId: item.id,
                              value,
                            })
                          }
                        />
                      </td>
                    ))}
                  </tr>
                ))}
                {rows.data?.length === 0 && (
                  <tr>
                    <td
                      className="empty-table-message"
                      colSpan={columns.data.length + 1}
                    >
                      <button
                        className="text-button"
                        onClick={() => addRow.mutate()}
                      >
                        Добавить первую строку
                      </button>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </section>
      {editing !== undefined && (
        <ColumnDialog
          column={editing}
          pending={save.isPending}
          onClose={() => setEditing(undefined)}
          onSave={(values) => save.mutate(values)}
        />
      )}{" "}
      {deleting && (
        <div className="dialog-backdrop">
          <div className="dialog" role="dialog" aria-modal="true">
            <h2>Удалить колонку?</h2>
            <p>«{deleting.name}» и все её значения будут удалены.</p>
            <div className="dialog-actions">
              <button
                className="secondary-button"
                onClick={() => setDeleting(null)}
              >
                Отмена
              </button>
              <button
                className="danger-button"
                onClick={() => remove.mutate(deleting.id)}
              >
                Удалить
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}

function CellEditor({
  journalId,
  rowId,
  column,
  value,
  onSave,
  onRefresh,
}: {
  journalId: string;
  rowId: string;
  column: JournalColumn;
  value: unknown;
  onSave: (value: unknown) => void;
  onRefresh: () => Promise<unknown>;
}) {
  if (column.type === "select")
    return (
      <select
        className="cell-input"
        value={typeof value === "string" ? value : ""}
        onChange={(e) => onSave(e.target.value || null)}
      >
        <option value="">—</option>
        {column.options.map((option) => (
          <option key={option}>{option}</option>
        ))}
      </select>
    );
  if (column.type === "boolean")
    return (
      <label className="cell-checkbox">
        <input
          type="checkbox"
          checked={value === true}
          onChange={(e) => onSave(e.target.checked)}
        />
      </label>
    );
  if (column.type === "image")
    return <ImageCell journalId={journalId} rowId={rowId} columnId={column.id} value={value} onRefresh={onRefresh} />;
  const type =
    column.type === "number"
      ? "number"
      : column.type === "date"
        ? "date"
        : "text";
  return (
    <input
      key={String(value)}
      className="cell-input"
      type={type}
      defaultValue={value == null ? "" : String(value)}
      onBlur={(e) => {
        const raw = e.target.value;
        const next =
          raw === "" ? null : column.type === "number" ? Number(raw) : raw;
        if (next !== value) onSave(next);
      }}
    />
  );
}

function ImageCell({ journalId, rowId, columnId, value, onRefresh }: { journalId: string; rowId: string; columnId: string; value: unknown; onRefresh: () => Promise<unknown> }) {
  const current = value && typeof value === "object" && "id" in value ? value as Attachment : null;
  const upload = useMutation({ mutationFn: (file: File) => attachmentsApi.upload(journalId, rowId, columnId, file), onSuccess: onRefresh });
  const remove = useMutation({ mutationFn: () => attachmentsApi.remove(journalId, rowId, columnId), onSuccess: onRefresh });
  if (current) return <div className="image-cell"><img src={attachmentsApi.contentUrl(current.id)} alt={current.name} /><span title={current.name}>{current.name}</span><a href={attachmentsApi.contentUrl(current.id)} target="_blank" rel="noreferrer" title="Открыть"><ExternalLink size={14} /></a><button title="Удалить" onClick={() => remove.mutate()}><X size={14} /></button></div>;
  return <label className="image-upload"><ImagePlus size={15} /><span>{upload.isPending ? "Загрузка…" : "Добавить"}</span><input type="file" accept="image/jpeg,image/png,image/webp" disabled={upload.isPending} onChange={(event) => { const file = event.target.files?.[0]; if (file) upload.mutate(file); }} /></label>;
}

function ColumnDialog({
  column,
  pending,
  onClose,
  onSave,
}: {
  column: JournalColumn | null;
  pending: boolean;
  onClose: () => void;
  onSave: (v: ColumnValues) => void;
}) {
  const [name, setName] = useState(column?.name ?? "");
  const [type, setType] = useState<ColumnType>(column?.type ?? "text");
  const [role, setRole] = useState<ColumnRole>(column?.role ?? null);
  const [options, setOptions] = useState(column?.options.join(", ") ?? "");
  const roles = compatibleRoles(type);
  const changeType = (value: ColumnType) => {
    setType(value);
    if (role && !compatibleRoles(value).includes(role)) setRole(null);
  };
  const submit = (e: FormEvent) => {
    e.preventDefault();
    onSave({
      name,
      type,
      role,
      options: type === "select" ? options.split(",") : [],
    });
  };
  return (
    <div
      className="dialog-backdrop"
      onMouseDown={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className="dialog column-dialog" role="dialog" aria-modal="true">
        <form onSubmit={submit}>
          <h2>{column ? "Настройка колонки" : "Новая колонка"}</h2>
          <label>
            Название
            <input
              autoFocus
              required
              maxLength={120}
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </label>
          <label>
            Тип
            <select
              value={type}
              onChange={(e) => changeType(e.target.value as ColumnType)}
            >
              {Object.entries(typeLabels).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </label>
          <label>
            Системная роль
            <select
              value={role ?? ""}
              onChange={(e) => setRole((e.target.value || null) as ColumnRole)}
            >
              <option value="">Без роли</option>
              {roles.map((value) => (
                <option key={value!} value={value!}>
                  {roleLabels[value!]}
                </option>
              ))}
            </select>
          </label>
          {type === "select" && (
            <label>
              Варианты выбора
              <textarea
                rows={3}
                value={options}
                onChange={(e) => setOptions(e.target.value)}
                placeholder="Win, Loss, Breakeven"
              />
              <small>Разделяйте варианты запятыми</small>
            </label>
          )}
          <div className="dialog-actions">
            <button
              type="button"
              className="secondary-button"
              onClick={onClose}
            >
              Отмена
            </button>
            <button
              className="primary-button"
              disabled={pending || !name.trim()}
            >
              Сохранить
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
