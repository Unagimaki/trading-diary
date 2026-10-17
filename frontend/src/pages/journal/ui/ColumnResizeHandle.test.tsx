import { act, cleanup, fireEvent, render, renderHook, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ColumnResizeHandle } from "./ColumnResizeHandle";
import { useColumnWidths } from "./useColumnWidths";

describe("column resizing", () => {
  let frame: FrameRequestCallback;
  beforeEach(() => {
    localStorage.clear();
    vi.stubGlobal("PointerEvent", class extends MouseEvent {
      pointerId: number;
      constructor(type: string, init: PointerEventInit) { super(type, init); this.pointerId = init.pointerId ?? 1; }
    });
    vi.stubGlobal("requestAnimationFrame", vi.fn((callback: FrameRequestCallback) => { frame = callback; return 1; }));
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
  });
  afterEach(() => { cleanup(); vi.unstubAllGlobals(); localStorage.clear(); });

  function setup() {
    const onResize = vi.fn();
    const { container } = render(<table style={{ width: 566 }}>
      <colgroup><col style={{ width: 46 }} /><col style={{ width: 260 }} /><col style={{ width: 260 }} /></colgroup>
      <thead><tr><th>#</th><th><ColumnResizeHandle name="Заметка" width={260} onResize={onResize} /></th><th>Другая колонка</th></tr></thead>
    </table>);
    const table = container.querySelector("table")!;
    table.getBoundingClientRect = () => ({ width: 566 } as DOMRect);
    const handle = screen.getByRole("separator");
    handle.setPointerCapture = vi.fn();
    handle.hasPointerCapture = () => true;
    handle.releasePointerCapture = vi.fn();
    return { table, handle, columns: container.querySelectorAll("col"), onResize };
  }

  it("batches pointer movements and changes only the selected column, committing once on release", () => {
    const { handle, columns, table, onResize } = setup();
    fireEvent.pointerDown(handle, { button: 0, clientX: 100, pointerId: 1 });
    fireEvent.pointerMove(handle, { clientX: 120, pointerId: 1 });
    fireEvent.pointerMove(handle, { clientX: 140, pointerId: 1 });
    expect(requestAnimationFrame).toHaveBeenCalledTimes(1);
    expect(onResize).not.toHaveBeenCalled();
    act(() => frame(0));
    expect(columns[1].style.width).toBe("300px");
    expect(columns[2].style.width).toBe("260px");
    expect(table.style.width).toBe("606px");
    fireEvent.pointerUp(handle, { pointerId: 1 });
    expect(onResize).toHaveBeenCalledExactlyOnceWith(300);
    expect(document.body).not.toHaveClass("resizing-column");
  });

  it("restores the original width on cancellation", () => {
    const { handle, columns, table, onResize } = setup();
    fireEvent.pointerDown(handle, { button: 0, clientX: 100, pointerId: 1 });
    fireEvent.pointerMove(handle, { clientX: 180, pointerId: 1 });
    act(() => frame(0));
    fireEvent.keyDown(handle, { key: "Escape" });
    expect(columns[1].style.width).toBe("260px");
    expect(table.style.width).toBe("566px");
    expect(onResize).not.toHaveBeenCalled();
  });

  it("supports keyboard resizing and bounds", () => {
    const { handle, onResize } = setup();
    fireEvent.keyDown(handle, { key: "ArrowRight" });
    expect(onResize).toHaveBeenLastCalledWith(270);
    fireEvent.keyDown(handle, { key: "Home" });
    expect(onResize).toHaveBeenLastCalledWith(160);
  });

  it("persists widths by column ID separately for each journal", () => {
    const first = renderHook(() => useColumnWidths("j1"));
    act(() => first.result.current.resize("c1", 340));
    first.unmount();
    const restored = renderHook(() => useColumnWidths("j1"));
    expect(restored.result.current.widths.c1).toBe(340);
    const other = renderHook(() => useColumnWidths("j2"));
    expect(other.result.current.widths.c1).toBeUndefined();
  });
});
