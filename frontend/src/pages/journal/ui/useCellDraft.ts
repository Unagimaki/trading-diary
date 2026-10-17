import { useEffect, useRef, useState } from "react";

// Keep unfinished edits separate from server refreshes and queued saves.
export function useCellDraft(value: unknown, saving: boolean, failed: boolean) {
  const initial = value == null ? "" : String(value);
  const [draft, setDraft] = useState(initial);
  const dirty = useRef(false);
  const submitted = useRef(initial);
  const focused = useRef(false);

  useEffect(() => {
    if (saving) return;
    if (failed) {
      submitted.current = initial;
      dirty.current = true;
      return;
    }
    if (!dirty.current && !focused.current) {
      setDraft(initial);
      submitted.current = initial;
    }
  }, [initial, saving, failed]);

  return {
    draft,
    change: (next: string) => { dirty.current = true; setDraft(next); },
    focus: () => { focused.current = true; },
    commit: (raw: string, save: () => void) => {
      focused.current = false;
      dirty.current = false;
      if (raw !== submitted.current) {
        submitted.current = raw;
        save();
      }
    },
  };
}
