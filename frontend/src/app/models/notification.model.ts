export type NotificationType = 'project_invite' | 'task_assigned' | 'deadline_reminder';

export interface AppNotification {
  id: string;
  type: NotificationType;
  message: string;
  is_read: boolean;
  created_at: string;
  link?: string;
}
