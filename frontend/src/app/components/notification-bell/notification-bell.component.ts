import { Component, inject } from '@angular/core';
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

  onNotificationClick(n: AppNotification) {
    this.notifService.markRead(n.id);
    if (n.link) this.router.navigateByUrl(n.link);
    this.close();
  }

  markAllRead() { this.notifService.markAllRead(); }

  iconForType(type: AppNotification['type']): string {
    if (type === 'deadline_reminder') return '⏰';
    if (type === 'task_assigned')     return '📋';
    return '🔔';
  }
}
