import { create } from "zustand";

export type Notification = { id: number; message: string; requestId?: string };
type NotificationState = {
  notifications: Notification[];
  show: (message: string, requestId?: string) => void;
  dismiss: (id: number) => void;
};
let nextId = 1;

export const useNotificationStore = create<NotificationState>((set, get) => ({
  notifications: [],
  show: (message, requestId) => {
    const id = nextId++;
    set((state) => ({
      notifications: [...state.notifications, { id, message, requestId }].slice(
        -3,
      ),
    }));
    window.setTimeout(() => get().dismiss(id), 7000);
  },
  dismiss: (id) =>
    set((state) => ({
      notifications: state.notifications.filter((item) => item.id !== id),
    })),
}));
