import { FormEvent, useEffect, useRef, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  ImagePlus,
  Plus,
  RotateCcw,
  Settings2,
  Trash2,
  TriangleAlert,
  X,
  ZoomIn,
  ZoomOut,
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
import { journalsApi, type Journal } from "@/entities/journal/api/journals";
import { rowsApi } from "@/entities/observation/api/rows";
import type { DataWarning } from "@/entities/observation/api/rows";
import { queryClient } from "@/shared/api/query-client";
import { errorMessage, errorRequestId } from "@/shared/api/api-error";
import { JournalAnalytics } from "@/widgets/journal-analytics";

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
  r: "RR (Risk/Reward)",
  risk: "Риск, %",
};
const roleHints: Partial<Record<Exclude<ColumnRole, null>, string>> = {
  r: "Положительное отношение прибыли к риску: 3 означает 1 : 3.",
  pnl: "PnL рассчитывается по результату, риску и RR; его можно переопределить вручную.",
  risk: "Процент депозита под риском. По умолчанию берётся из настроек журнала.",
  trade_result: "Win, Loss или Breakeven запускает расчёт PnL.",
};
const compatibleRoles = (type: ColumnType): Exclude<ColumnRole, null>[] =>
  type === "select" ? ["trade_result"] : type === "number" ? ["pnl", "r", "risk"] : [];
const warningLabels: Record<DataWarning, string> = {
  result_pnl_conflict: "Результат сделки противоречит знаку PnL",
  result_r_conflict: "Результат сделки противоречит знаку R",
  pnl_risk_r_conflict: "PnL не совпадает с расчётом Risk × R",
  risk_not_positive: "Риск должен быть больше нуля",
  rr_not_positive: "RR должен быть больше нуля",
};

export function JournalPage() {
  const { id = "" } = useParams();
  const [view, setView] = useState<"journal" | "analytics">("journal");
  const [editing, setEditing] = useState<JournalColumn | null | undefined>(
    undefined,
  );
  const [deleting, setDeleting] = useState<JournalColumn | null>(null);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [hoveredImage, setHoveredImage] = useState<Attachment | null>(null);
  const [pinnedImage, setPinnedImage] = useState<Attachment | null>(null);
  const [previewSize, setPreviewSize] = useState(25);
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
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
  const refresh = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["columns", id] }),
      queryClient.invalidateQueries({ queryKey: ["rows", id] }),
      queryClient.invalidateQueries({ queryKey: ["analytics", id] }),
    ]);
  };
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
  const refreshRows = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["rows", id] }),
      queryClient.invalidateQueries({ queryKey: ["analytics", id] }),
    ]);
  };
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
  const saveSettings = useMutation({
    mutationFn: ({ initialDeposit, riskPercent, defaultRR }: { initialDeposit: number; riskPercent: number; defaultRR: number }) =>
      journalsApi.updateSettings(id, initialDeposit, riskPercent, defaultRR),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["journal", id] }),
        queryClient.invalidateQueries({ queryKey: ["rows", id] }),
        queryClient.invalidateQueries({ queryKey: ["analytics", id] }),
      ]);
      setSettingsOpen(false);
    },
  });
  useEffect(() => {
    const closePinnedImage = (event: KeyboardEvent) => {
      if (event.key === "Escape") setPinnedImage(null);
    };
    window.addEventListener("keydown", closePinnedImage);
    return () => {
      window.removeEventListener("keydown", closePinnedImage);
      if (hoverTimer.current) clearTimeout(hoverTimer.current);
    };
  }, []);
  const showHoveredImage = (image: Attachment) => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current);
    setHoveredImage(image);
  };
  const keepHoveredImage = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current);
  };
  const hideHoveredImage = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setHoveredImage(null), 250);
  };
  const resizePreview = (change: number) =>
    setPreviewSize((size) => Math.min(42, Math.max(18, size + change)));
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
              className="secondary-button icon-button"
              title="Настройки расчётов"
              aria-label="Настройки расчётов"
              onClick={() => setSettingsOpen(true)}
            >
              <Settings2 size={16} />
            </button>
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
        <div className="journal-tabs" role="tablist" aria-label="Раздел журнала">
          <button
            role="tab"
            aria-selected={view === "journal"}
            className={view === "journal" ? "active" : ""}
            onClick={() => setView("journal")}
          >
            Журнал
          </button>
          <button
            role="tab"
            aria-selected={view === "analytics"}
            className={view === "analytics" ? "active" : ""}
            onClick={() => setView("analytics")}
          >
            Аналитика
          </button>
        </div>
        {view === "analytics" && <JournalAnalytics journalId={id} />}
        {view === "journal" && columns.isPending && <div className="list-status">Загрузка…</div>}
        {view === "journal" && columns.isError && (
          <div className="list-status error">Не удалось загрузить колонки</div>
        )}
        {view === "journal" && columns.data?.length === 0 && (
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
        {view === "journal" && !!columns.data?.length && (
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
                  <tr key={row.id} className={row.warnings?.length ? "row-warning" : undefined}>
                    <td className="row-number row-control">
                      <span>{index + 1}</span>
                      {!!row.warnings?.length && (
                        <TriangleAlert
                          className="row-warning-icon"
                          size={13}
                          aria-label="Есть несогласованные данные"
                        >
                          <title>{row.warnings.map((code) => warningLabels[code]).join("; ")}</title>
                        </TriangleAlert>
                      )}
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
                          onImageEnter={showHoveredImage}
                          onImageLeave={hideHoveredImage}
                          onImagePin={setPinnedImage}
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
      {(hoveredImage || pinnedImage) && (
        <div
          className={`image-preview-dock${hoveredImage && pinnedImage && hoveredImage.id !== pinnedImage.id ? " comparing" : ""}`}
        >
          {hoveredImage && hoveredImage.id !== pinnedImage?.id && (
            <ImagePreview
              image={hoveredImage}
              label={pinnedImage ? "Сравнение" : "Предпросмотр"}
              size={previewSize}
              onIncrease={() => resizePreview(5)}
              onDecrease={() => resizePreview(-5)}
              onControlsEnter={keepHoveredImage}
              onControlsLeave={hideHoveredImage}
            />
          )}
          {pinnedImage && (
            <ImagePreview
              image={pinnedImage}
              label="Закреплено"
              size={previewSize}
              onIncrease={() => resizePreview(5)}
              onDecrease={() => resizePreview(-5)}
              onClose={() => setPinnedImage(null)}
              onControlsEnter={keepHoveredImage}
              onControlsLeave={hideHoveredImage}
            />
          )}
        </div>
      )}
      {editing !== undefined && (
        <ColumnDialog
          column={editing}
          pending={save.isPending}
          onClose={() => setEditing(undefined)}
          onSave={(values) => save.mutate(values)}
        />
      )}{" "}
      {settingsOpen && journal.data && (
        <JournalSettingsDialog
          journal={journal.data}
          pending={saveSettings.isPending}
          onClose={() => setSettingsOpen(false)}
          onSave={(initialDeposit, riskPercent, defaultRR) => saveSettings.mutate({ initialDeposit, riskPercent, defaultRR })}
        />
      )}
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
  onImageEnter,
  onImageLeave,
  onImagePin,
}: {
  journalId: string;
  rowId: string;
  column: JournalColumn;
  value: unknown;
  onSave: (value: unknown) => void;
  onRefresh: () => Promise<unknown>;
  onImageEnter: (image: Attachment) => void;
  onImageLeave: () => void;
  onImagePin: (image: Attachment) => void;
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
    return (
      <ImageCell
        journalId={journalId}
        rowId={rowId}
        columnId={column.id}
        value={value}
        onRefresh={onRefresh}
        onImageEnter={onImageEnter}
        onImageLeave={onImageLeave}
        onImagePin={onImagePin}
      />
    );
  const type =
    column.type === "number"
      ? "number"
      : column.type === "date"
        ? "date"
        : "text";
  const input = (
    <input
      key={String(value)}
      className="cell-input"
      type={type}
      min={column.role === "risk" || column.role === "r" ? 0.01 : undefined}
      max={column.role === "risk" ? 100 : undefined}
      step={column.type === "number" ? "any" : undefined}
      defaultValue={value == null ? "" : String(value)}
      onBlur={(e) => {
        const raw = e.target.value;
        if (raw !== "" && !e.target.checkValidity()) {
          e.target.reportValidity();
          e.target.value = value == null ? "" : String(value);
          return;
        }
        const next =
          raw === "" ? null : column.type === "number" ? Number(raw) : raw;
        if (next !== value) onSave(next);
      }}
    />
  );
  if (column.role === "r") {
    return <div className="rr-cell"><span>1 :</span>{input}</div>;
  }
  if (column.role === "risk") {
    return <div className="risk-cell">{input}<span>%</span></div>;
  }
  return input;
}

function ImageCell({
  journalId,
  rowId,
  columnId,
  value,
  onRefresh,
  onImageEnter,
  onImageLeave,
  onImagePin,
}: {
  journalId: string;
  rowId: string;
  columnId: string;
  value: unknown;
  onRefresh: () => Promise<unknown>;
  onImageEnter: (image: Attachment) => void;
  onImageLeave: () => void;
  onImagePin: (image: Attachment) => void;
}) {
  const [lastFile, setLastFile] = useState<File | null>(null);
  const current =
    value && typeof value === "object" && "id" in value
      ? (value as Attachment)
      : null;
  const upload = useMutation({
    mutationFn: (file: File) =>
      attachmentsApi.upload(journalId, rowId, columnId, file),
    onSuccess: onRefresh,
  });
  const remove = useMutation({
    mutationFn: () => attachmentsApi.remove(journalId, rowId, columnId),
    onSuccess: onRefresh,
  });
  if (current)
    return (
      <div className="image-cell">
        <button
          className="image-thumbnail"
          type="button"
          title="Закрепить изображение"
          aria-label={`Закрепить изображение ${current.name}`}
          onMouseEnter={() => onImageEnter(current)}
          onMouseLeave={onImageLeave}
          onFocus={() => onImageEnter(current)}
          onBlur={onImageLeave}
          onClick={() => onImagePin(current)}
        >
          <img src={attachmentsApi.contentUrl(current.id)} alt="" />
        </button>
        <span title={current.name}>{current.name}</span>
        <a
          href={attachmentsApi.contentUrl(current.id)}
          target="_blank"
          rel="noreferrer"
          title="Открыть"
        >
          <ExternalLink size={14} />
        </a>
        <button title="Удалить" onClick={() => remove.mutate()}>
          <X size={14} />
        </button>
      </div>
    );
  if (upload.isError)
    return (
      <div
        className="image-error"
        title={
          errorRequestId(upload.error)
            ? `Код: ${errorRequestId(upload.error)}`
            : undefined
        }
      >
        <span>{errorMessage(upload.error)}</span>
        <button
          title="Повторить"
          disabled={!lastFile}
          onClick={() => lastFile && upload.mutate(lastFile)}
        >
          <RotateCcw size={14} />
        </button>
      </div>
    );
  return (
    <label className="image-upload">
      <ImagePlus size={15} />
      <span>{upload.isPending ? "Загрузка…" : "Добавить"}</span>
      <input
        type="file"
        accept="image/jpeg,image/png,image/webp"
        disabled={upload.isPending}
        onChange={(event) => {
          const file = event.target.files?.[0];
          if (file) {
            setLastFile(file);
            upload.mutate(file);
          }
        }}
      />
    </label>
  );
}

function ImagePreview({
  image,
  label,
  size,
  onIncrease,
  onDecrease,
  onClose,
  onControlsEnter,
  onControlsLeave,
}: {
  image: Attachment;
  label: string;
  size: number;
  onIncrease: () => void;
  onDecrease: () => void;
  onClose?: () => void;
  onControlsEnter: () => void;
  onControlsLeave: () => void;
}) {
  return (
    <aside
      className="image-preview"
      aria-label={`${label}: ${image.name}`}
      style={{ width: `clamp(240px, ${size}vw, 680px)` }}
    >
      <img src={attachmentsApi.contentUrl(image.id)} alt={image.name} />
      <div className="image-preview-caption" title={image.name}>
        <span>{label}</span>
        {image.name}
      </div>
      <div
        className="image-preview-controls"
        onMouseEnter={onControlsEnter}
        onMouseLeave={onControlsLeave}
      >
        <button
          type="button"
          title="Уменьшить"
          aria-label="Уменьшить изображение"
          onClick={onDecrease}
          disabled={size <= 18}
        >
          <ZoomOut size={15} />
        </button>
        <button
          type="button"
          title="Увеличить"
          aria-label="Увеличить изображение"
          onClick={onIncrease}
          disabled={size >= 42}
        >
          <ZoomIn size={15} />
        </button>
        {onClose && (
          <button
            className="image-preview-close"
            type="button"
            title="Закрыть"
            aria-label="Закрыть закреплённое изображение"
            onClick={onClose}
          >
            <X size={16} />
          </button>
        )}
      </div>
    </aside>
  );
}

function JournalSettingsDialog({
  journal,
  pending,
  onClose,
  onSave,
}: {
  journal: Journal;
  pending: boolean;
  onClose: () => void;
  onSave: (initialDeposit: number, riskPercent: number, defaultRR: number) => void;
}) {
  const [initialDeposit, setInitialDeposit] = useState(String(journal.initialDeposit));
  const [riskPercent, setRiskPercent] = useState(String(journal.riskPercent));
  const [defaultRR, setDefaultRR] = useState(String(journal.defaultRR));
  const deposit = Number(initialDeposit);
  const risk = Number(riskPercent);
  const rr = Number(defaultRR);
  const valid = Number.isFinite(deposit) && deposit > 0 && Number.isFinite(risk) && risk > 0 && risk <= 100 && Number.isFinite(rr) && rr > 0;
  const defaultRisk = valid ? deposit * risk / 100 : 0;
  return (
    <div className="dialog-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <div className="dialog journal-settings-dialog" role="dialog" aria-modal="true" aria-labelledby="journal-settings-title">
        <form onSubmit={(event) => { event.preventDefault(); if (valid) onSave(deposit, risk, rr); }}>
          <h2 id="journal-settings-title">Настройки расчётов</h2>
          <p>Эти значения используются, когда в строке не указан индивидуальный риск.</p>
          <label>
            Начальный депозит
            <input type="number" min="0.01" step="0.01" required value={initialDeposit} onChange={(event) => setInitialDeposit(event.target.value)} />
          </label>
          <label>
            Риск на сделку, %
            <input type="number" min="0.01" max="100" step="0.01" required value={riskPercent} onChange={(event) => setRiskPercent(event.target.value)} />
          </label>
          <label>
            RR по умолчанию (1 : N)
            <input type="number" min="0.01" step="any" required value={defaultRR} onChange={(event) => setDefaultRR(event.target.value)} />
          </label>
          <div className="default-risk-value">
            <span>Риск в деньгах</span>
            <strong>{valid ? defaultRisk.toLocaleString("ru-RU", { maximumFractionDigits: 2 }) : "—"}</strong>
          </div>
          <div className="dialog-actions">
            <button type="button" className="secondary-button" onClick={onClose}>Отмена</button>
            <button className="primary-button" disabled={pending || !valid}>Сохранить</button>
          </div>
        </form>
      </div>
    </div>
  );
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
              onChange={(e) => {
                const nextRole = (e.target.value || null) as ColumnRole;
                setRole(nextRole);
                if (nextRole === "trade_result" && !options.trim()) {
                  setOptions("Win, Loss, Breakeven");
                }
              }}
            >
              <option value="">Без роли</option>
              {roles.map((value) => (
                <option key={value!} value={value!}>
                  {roleLabels[value!]}
                </option>
              ))}
            </select>
            {role && <small>{roleHints[role]}</small>}
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
