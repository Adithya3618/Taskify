import { Injectable, signal, computed } from '@angular/core';
import { AppNotification, NotificationType } from '../models/notification.model';

@Injectable({ providedIn: 'root' })
export class NotificationService {
  private readonly storageKey = 'taskify.notifications';

  private _notifications = signal<AppNotification[]>(this.load());

  notifications = this._notifications.asReadonly();
  unreadCount = computed(() => this._notifications().filter(n => !n.is_read).length);

  private load(): AppNotification[] {
    try {
      const raw = localStorage.getItem(this.storageKey);
      return raw ? JSON.parse(raw) : [];
    } catch { return []; }
  }

  private save(list: AppNotification[]): void {
    localStorage.setItem(this.storageKey, JSON.stringify(list));
  }

  add(type: NotificationType, message: string, link?: string): void {
    const existing = this._notifications();
    // Avoid duplicate deadline reminders for same message
    if (existing.some(n => n.message === message && !n.is_read)) return;

    const n: AppNotification = {
      id: `notif-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      type,
      message,
      is_read: false,
      created_at: new Date().toISOString(),
      link
    };
    const updated = [n, ...existing].slice(0, 50); // keep latest 50
    this._notifications.set(updated);
    this.save(updated);
  }

  markRead(id: string): void {
    const updated = this._notifications().map(n =>
      n.id === id ? { ...n, is_read: true } : n
    );
    this._notifications.set(updated);
    this.save(updated);
  }

  markAllRead(): void {
    const updated = this._notifications().map(n => ({ ...n, is_read: true }));
    this._notifications.set(updated);
    this.save(updated);
  }

  remove(id: string): void {
    const updated = this._notifications().filter(n => n.id !== id);
    this._notifications.set(updated);
    this.save(updated);
  }

  clearAll(): void {
    this._notifications.set([]);
    this.save([]);
  }

  /** Called by board component when tasks load — generates deadline reminders. */
  checkDeadlines(tasks: { id: number; title: string; deadline?: string }[], projectId: number): void {
    const today = new Date(); today.setHours(0, 0, 0, 0);

    tasks.forEach(task => {
      if (!task.deadline) return;
      const due = new Date(task.deadline); due.setHours(0, 0, 0, 0);
      const daysLeft = Math.ceil((due.getTime() - today.getTime()) / 86_400_000);

      if (daysLeft < 0) {
        this.add('deadline_reminder', `"${task.title}" is overdue!`, `/board/${projectId}`);
      } else if (daysLeft === 0) {
        this.add('deadline_reminder', `"${task.title}" is due today.`, `/board/${projectId}`);
      } else if (daysLeft === 1) {
        this.add('deadline_reminder', `"${task.title}" is due tomorrow.`, `/board/${projectId}`);
      }
    });
  }

  timeAgo(isoDate: string): string {
    const diff = Date.now() - new Date(isoDate).getTime();
    const mins = Math.floor(diff / 60_000);
    if (mins < 1)  return 'just now';
    if (mins < 60) return `${mins}m ago`;
    const hrs = Math.floor(mins / 60);
    if (hrs < 24)  return `${hrs}h ago`;
    const days = Math.floor(hrs / 24);
    return `${days}d ago`;
  }
}
