import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

type ToastType = 'info' | 'success' | 'error';

export interface ToastMessage {
  id: number;
  type: ToastType;
  message: string;
}

@Injectable({
  providedIn: 'root',
})
export class ToastService {
  private nextId = 1;
  private toasts = new BehaviorSubject<ToastMessage[]>([]);
  private expiryTimers = new Map<number, number>();

  readonly toasts$ = this.toasts.asObservable();

  show(message: string, type: ToastType = 'info', ttlMs = 5000) {
    const item: ToastMessage = {
      id: this.nextId++,
      type,
      message,
    };

    const current = this.toasts.value;
    this.toasts.next([...current, item]);

    const timer = window.setTimeout(() => {
      this.remove(item.id);
    }, Math.max(1200, ttlMs));
    this.expiryTimers.set(item.id, timer);

    return item.id;
  }

  success(message: string, ttlMs = 5000) {
    return this.show(message, 'success', ttlMs);
  }

  error(message: string, ttlMs = 5000) {
    return this.show(message, 'error', ttlMs);
  }

  remove(id: number) {
    const timer = this.expiryTimers.get(id);
    if (timer) {
      window.clearTimeout(timer);
      this.expiryTimers.delete(id);
    }
    const current = this.toasts.value.filter((item) => item.id !== id);
    if (current.length !== this.toasts.value.length) {
      this.toasts.next(current);
    }
  }
}
