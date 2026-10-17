import { useState } from "react";
import { clampColumnWidth } from "./column-widths";

function readWidths(journalId: string): Record<string, number> {
  try {
    const stored: unknown = JSON.parse(localStorage.getItem(`journal-column-widths:${journalId}`) ?? "{}");
    if (!stored || typeof stored !== "object" || Array.isArray(stored)) return {};
    return Object.fromEntries(Object.entries(stored).filter(([, value]) => typeof value === "number" && Number.isFinite(value))
      .map(([key, value]) => [key, clampColumnWidth(value as number)]));
  } catch { return {}; }
}

export function useColumnWidths(journalId: string) {
  const [state, setState] = useState(() => ({ journalId, widths: readWidths(journalId) }));
  if (state.journalId !== journalId) setState({ journalId, widths: readWidths(journalId) });
  return {
    widths: state.widths,
    resize: (columnId: string, width: number) => {
      const widths = { ...state.widths, [columnId]: clampColumnWidth(width) };
      setState({ journalId, widths });
      try { localStorage.setItem(`journal-column-widths:${journalId}`, JSON.stringify(widths)); } catch { /* Resizing also works when storage is unavailable. */ }
    },
  };
}
