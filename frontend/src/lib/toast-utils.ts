// Toast 管理工具
type ToastType = 'success' | 'error' | 'info';

interface ToastItem {
  id: number;
  message: string;
  type: ToastType;
}

let toastId = 0;
const toastListeners: Array<(toasts: ToastItem[]) => void> = [];
let toasts: ToastItem[] = [];

function emitChange() {
  toastListeners.forEach((listener) => listener(toasts));
}

export const toast = {
  success: (message: string) => {
    const id = toastId++;
    toasts = [...toasts, { id, message, type: 'success' }];
    emitChange();
  },
  error: (message: string) => {
    const id = toastId++;
    toasts = [...toasts, { id, message, type: 'error' }];
    emitChange();
  },
  info: (message: string) => {
    const id = toastId++;
    toasts = [...toasts, { id, message, type: 'info' }];
    emitChange();
  },
  remove: (id: number) => {
    toasts = toasts.filter((t) => t.id !== id);
    emitChange();
  },
  subscribe: (listener: (toasts: ToastItem[]) => void) => {
    toastListeners.push(listener);
    return () => {
      const index = toastListeners.indexOf(listener);
      if (index > -1) {
        toastListeners.splice(index, 1);
      }
    };
  },
  getToasts: () => toasts,
};

export type { ToastItem, ToastType };
