import { AlertCircle, X } from "lucide-react";
import { useNotificationStore } from "@/shared/model/notification-store";

export function Notifications() {
  const { notifications, dismiss } = useNotificationStore();
  return (
    <div className="notifications" aria-live="polite">
      {notifications.map((item) => (
        <div className="notification" role="alert" key={item.id}>
          <AlertCircle size={18} />
          <div>
            <strong>{item.message}</strong>
            {item.requestId && <small>Код обращения: {item.requestId}</small>}
          </div>
          <button aria-label="Закрыть" onClick={() => dismiss(item.id)}>
            <X size={16} />
          </button>
        </div>
      ))}
    </div>
  );
}
