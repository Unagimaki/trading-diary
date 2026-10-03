import { useLayoutEffect, useRef } from "react";

type TextCellProps = {
  value: unknown;
  wrapText: boolean;
  onSave: (value: string | null) => void;
};

export function TextCell({ value, wrapText, onSave }: TextCellProps) {
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
  }, [wrapText, value]);

  return (
    <textarea
      ref={ref}
      key={String(value)}
      className="cell-input text-cell"
      rows={1}
      wrap={wrapText ? "soft" : "off"}
      defaultValue={value == null ? "" : String(value)}
      onInput={resize}
      onBlur={(event) => {
        const next = event.target.value || null;
        if (next !== value) onSave(next);
      }}
    />
  );
}
