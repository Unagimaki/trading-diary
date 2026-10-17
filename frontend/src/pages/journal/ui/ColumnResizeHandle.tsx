import { useEffect, useRef } from "react";
import { MIN_COLUMN_WIDTH, MAX_COLUMN_WIDTH, clampColumnWidth } from "./column-widths";

type Props = { name: string; width: number; onResize: (width: number) => void };
type Drag = {
  pointerId: number;
  startX: number;
  startWidth: number;
  tableWidth: number;
  width: number;
  table: HTMLTableElement;
  column: HTMLTableColElement;
  handle: HTMLDivElement;
  frame: number | null;
};

export function ColumnResizeHandle({ name, width, onResize }: Props) {
  const drag = useRef<Drag | null>(null);
  const apply = (current: Drag) => {
    current.column.style.width = `${current.width}px`;
    current.table.style.width = `${current.tableWidth + current.width - current.startWidth}px`;
    current.handle.setAttribute("aria-valuenow", String(current.width));
  };
  const finish = (cancel = false) => {
    const current = drag.current;
    if (!current) return;
    drag.current = null;
    if (current.frame !== null) cancelAnimationFrame(current.frame);
    if (cancel) current.width = current.startWidth;
    apply(current);
    current.handle.removeAttribute("data-resizing");
    document.body.classList.remove("resizing-column");
    if (current.handle.hasPointerCapture(current.pointerId)) current.handle.releasePointerCapture(current.pointerId);
    if (!cancel) onResize(current.width);
  };

  useEffect(() => () => {
    const current = drag.current;
    if (!current) return;
    if (current.frame !== null) cancelAnimationFrame(current.frame);
    document.body.classList.remove("resizing-column");
    drag.current = null;
  }, []);

  return (
    <div
      className="column-resizer"
      role="separator"
      tabIndex={0}
      aria-label={`Ширина колонки «${name}»`}
      aria-orientation="vertical"
      aria-valuemin={MIN_COLUMN_WIDTH}
      aria-valuemax={MAX_COLUMN_WIDTH}
      aria-valuenow={width}
      title="Потяните для изменения ширины"
      onPointerDown={(event) => {
        if (event.button !== 0 || drag.current) return;
        const header = event.currentTarget.closest("th");
        const table = header?.closest("table");
        const column = table?.querySelectorAll("col")[header!.cellIndex];
        if (!table || !column) return;
        event.preventDefault();
        event.stopPropagation();
        event.currentTarget.focus({ preventScroll: true });
        event.currentTarget.setPointerCapture(event.pointerId);
        drag.current = {
          pointerId: event.pointerId, startX: event.clientX, startWidth: width, width,
          tableWidth: table.getBoundingClientRect().width, table, column, handle: event.currentTarget, frame: null,
        };
        event.currentTarget.setAttribute("data-resizing", "true");
        document.body.classList.add("resizing-column");
      }}
      onPointerMove={(event) => {
        const current = drag.current;
        if (!current || current.pointerId !== event.pointerId) return;
        current.width = clampColumnWidth(current.startWidth + event.clientX - current.startX);
        if (current.frame !== null) return;
        current.frame = requestAnimationFrame(() => { current.frame = null; apply(current); });
      }}
      onPointerUp={() => finish()}
      onPointerCancel={() => finish(true)}
      onLostPointerCapture={() => finish()}
      onKeyDown={(event) => {
        if (event.key === "Escape" && drag.current) { event.preventDefault(); finish(true); return; }
        const step = event.shiftKey ? 40 : 10;
        const next = event.key === "ArrowLeft" ? width - step : event.key === "ArrowRight" ? width + step :
          event.key === "Home" ? MIN_COLUMN_WIDTH : event.key === "End" ? MAX_COLUMN_WIDTH : null;
        if (next !== null) { event.preventDefault(); onResize(clampColumnWidth(next)); }
      }}
    />
  );
}
