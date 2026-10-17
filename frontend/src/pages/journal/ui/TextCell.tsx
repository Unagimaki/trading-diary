import { useLayoutEffect, useRef } from "react";
import { useCellDraft } from "./useCellDraft";

type TextCellProps = {
  value: unknown;
  wrapText: boolean;
  onSave: (value: string | null) => void;
  saving?: boolean;
  failed?: boolean;
};

export function TextCell({ value, wrapText, onSave, saving = false, failed = false }: TextCellProps) {
  const cell = useCellDraft(value, saving, failed);
  const ref = useRef<HTMLTextAreaElement>(null);
  const resize = () => {
    const element = ref.current;
    if (!element) return;
    element.style.height = "51px";
    if (wrapText) element.style.height = `${Math.max(51, element.scrollHeight)}px`;
  };

  useLayoutEffect(() => {
    const element = ref.current;
    if (!element) return;
    const updateHeight = () => {
      element.style.height = "51px";
      if (wrapText) element.style.height = `${Math.max(51, element.scrollHeight)}px`;
    };
    updateHeight();
    let width = element.getBoundingClientRect().width;
    const observer = new ResizeObserver(() => {
      const nextWidth = element.getBoundingClientRect().width;
      if (nextWidth !== width) {
        width = nextWidth;
        updateHeight();
      }
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, [wrapText, cell.draft]);

  return (
    <textarea
      ref={ref}
      className="cell-input text-cell"
      rows={1}
      wrap={wrapText ? "soft" : "off"}
      value={cell.draft}
      onChange={(event) => cell.change(event.target.value)}
      onFocus={cell.focus}
      onInput={resize}
      onBlur={(event) => {
        const raw = event.target.value;
        cell.commit(raw, () => onSave(raw || null));
      }}
    />
  );
}
