// Toast notification store
// Usage: import { addToast } from '$lib/stores/toast';
//        addToast({ type: 'success', message: 'Done!' });

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
	id: string;
	type: ToastType;
	message: string;
	title?: string;
	duration?: number;
}

type Subscriber = (toasts: Toast[]) => void;

let _toasts: Toast[] = [];
const subscribers: Set<Subscriber> = new Set();

function notify(): void {
	for (const sub of subscribers) {
		sub([..._toasts]);
	}
}

export function subscribe(fn: Subscriber): () => void {
	subscribers.add(fn);
	fn([..._toasts]);
	return () => {
		subscribers.delete(fn);
	};
}

export function getToasts(): Toast[] {
	return [..._toasts];
}

export function addToast(toast: Omit<Toast, 'id'>): string {
	const id = Math.random().toString(36).slice(2) + Date.now().toString(36);
	const duration = toast.duration ?? 5000;

	_toasts = [..._toasts, { ...toast, id }];
	notify();

	if (duration > 0) {
		setTimeout(() => {
			removeToast(id);
		}, duration);
	}

	return id;
}

export function removeToast(id: string): void {
	_toasts = _toasts.filter((t) => t.id !== id);
	notify();
}

export function clearToasts(): void {
	_toasts = [];
	notify();
}

// Convenience methods
export function toastSuccess(message: string, title?: string): string {
	return addToast({ type: 'success', message, title });
}

export function toastError(message: string, title?: string): string {
	return addToast({ type: 'error', message, title, duration: 8000 });
}

export function toastWarning(message: string, title?: string): string {
	return addToast({ type: 'warning', message, title });
}

export function toastInfo(message: string, title?: string): string {
	return addToast({ type: 'info', message, title });
}
