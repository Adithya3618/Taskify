import { Component, HostListener, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { NotificationService } from '../../services/notification.service';
import { AppNotification } from '../../models/notification.model';

@Component({
  selector: 'app-notification-bell',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './notification-bell.component.html',
  styleUrl: './notification-bell.component.scss'
})
export class NotificationBellComponent {
  private router = inject(Router);
  notifService = inject(NotificationService);

  isOpen = false;

  toggle() { this.isOpen = !this.isOpen; }
  close()  { this.isOpen = false; }

  @HostListener('document:keydown.escape')
  onEscape() { this.close(); }

  onNotificationClick(n: AppNotification) {
    this.notifService.markRead(n.id);
    if (n.link) this.router.navigateByUrl(n.link);
    this.close();
  }

  dismiss(event: Event, id: string) {
    event.stopPropagation();
    this.notifService.remove(id);
  }

  markAllRead() { this.notifService.markAllRead(); }
  clearAll()    { this.notifService.clearAll(); }

  iconForType(type: AppNotification['type']): string {
    if (type === 'deadline_reminder') return '⏰';
    if (type === 'task_assigned')     return '📋';
    return '🔔';
  }
}
