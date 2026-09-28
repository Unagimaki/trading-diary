import type { ReactNode } from "react";
import { AlertCircle, Inbox } from "lucide-react";

export function Spinner({ label = "Загрузка" }: { label?: string }) {
  return <span className="ui-spinner" role="status" aria-label={label} />;
}

export function StateView({
  kind,
  title,
  description,
  action,
  compact = false,
}: {
  kind: "loading" | "empty" | "error";
  title: string;
  description?: string;
  action?: ReactNode;
  compact?: boolean;
}) {
  return (
    <div className={`state-view state-view-${kind}${compact ? " state-view-compact" : ""}`} role={kind === "error" ? "alert" : "status"}>
      <div className="state-view-icon" aria-hidden="true">
        {kind === "loading" ? <Spinner /> : kind === "error" ? <AlertCircle size={19} /> : <Inbox size={19} />}
      </div>
      <div className="state-view-copy">
        <strong>{title}</strong>
        {description && <p>{description}</p>}
      </div>
      {action && <div className="state-view-action">{action}</div>}
    </div>
  );
}
