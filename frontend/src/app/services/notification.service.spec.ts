import { TestBed } from '@angular/core/testing';
import { NotificationService } from './notification.service';
import { AppNotification } from '../models/notification.model';

const STORAGE_KEY = 'taskify.notifications';

describe('NotificationService', () => {
  let service: NotificationService;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({});
    service = TestBed.inject(NotificationService);
  });

  afterEach(() => {
    localStorage.clear();
  });

  // ── Initial state ──────────────────────────────────────────────────────────

  it('should start with an empty notifications list when localStorage is empty', () => {
    expect(service.notifications()).toEqual([]);
  });

  it('should load persisted notifications from localStorage on creation', () => {
    localStorage.clear();
    const stored: AppNotification[] = [{
      id: 'notif-1',
      type: 'task_assigned',
      message: 'Persisted task',
      is_read: false,
      created_at: new Date().toISOString()
    }];
    localStorage.setItem(STORAGE_KEY, JSON.stringify(stored));

    TestBed.resetTestingModule();
    TestBed.configureTestingModule({});
    const freshService = TestBed.inject(NotificationService);

    expect(freshService.notifications().length).toBe(1);
    expect(freshService.notifications()[0].message).toBe('Persisted task');
  });

  it('should return empty list when localStorage contains invalid JSON', () => {
    localStorage.clear();
    localStorage.setItem(STORAGE_KEY, 'not-valid-json{{{');

    TestBed.resetTestingModule();
    TestBed.configureTestingModule({});
    const freshService = TestBed.inject(NotificationService);

    expect(freshService.notifications()).toEqual([]);
  });

  // ── unreadCount ────────────────────────────────────────────────────────────

  it('should report unreadCount of 0 when no notifications exist', () => {
    expect(service.unreadCount()).toBe(0);
  });

  it('should compute unreadCount for multiple unread notifications', () => {
    service.add('task_assigned', 'Task Alpha assigned');
    service.add('project_invite', 'Project Beta invite');
    expect(service.unreadCount()).toBe(2);
  });

  it('should decrease unreadCount after a notification is marked read', () => {
    service.add('task_assigned', 'Task Alpha assigned');
    service.add('task_assigned', 'Task Beta assigned');
    const id = service.notifications()[0].id;
    service.markRead(id);
    expect(service.unreadCount()).toBe(1);
  });

  // ── add() ──────────────────────────────────────────────────────────────────

  it('should add a notification with correct type, message, and is_read=false', () => {
    service.add('task_assigned', 'You have been assigned Task X');
    const notif = service.notifications()[0];
    expect(notif.type).toBe('task_assigned');
    expect(notif.message).toBe('You have been assigned Task X');
    expect(notif.is_read).toBeFalse();
  });

  it('should generate a unique id for each notification', () => {
    service.add('task_assigned', 'Task 1');
    service.add('project_invite', 'Project invite');
    const ids = service.notifications().map(n => n.id);
    expect(new Set(ids).size).toBe(ids.length);
  });

  it('should persist the new notification to localStorage', () => {
    service.add('project_invite', 'Invited to project');
    const stored: AppNotification[] = JSON.parse(localStorage.getItem(STORAGE_KEY)!);
    expect(stored.length).toBe(1);
    expect(stored[0].message).toBe('Invited to project');
  });

  it('should prepend new notifications so the latest appears first', () => {
    service.add('task_assigned', 'First notification');
    service.add('task_assigned', 'Second notification');
    expect(service.notifications()[0].message).toBe('Second notification');
    expect(service.notifications()[1].message).toBe('First notification');
  });

  it('should not add a duplicate if an unread notification with the same message exists', () => {
    service.add('deadline_reminder', '"Task A" is due today.');
    service.add('deadline_reminder', '"Task A" is due today.');
    expect(service.notifications().length).toBe(1);
  });

  it('should allow re-adding the same message once the original has been read', () => {
    service.add('deadline_reminder', '"Task A" is due today.');
    service.markAllRead();
    service.add('deadline_reminder', '"Task A" is due today.');
    expect(service.notifications().length).toBe(2);
  });

  it('should store the optional link when provided', () => {
    service.add('deadline_reminder', '"Task B" is overdue!', '/board/7');
    expect(service.notifications()[0].link).toBe('/board/7');
  });

  it('should cap the notifications list at 50 entries', () => {
    for (let i = 0; i < 55; i++) {
      service.add('project_invite', `Unique notification message ${i}`);
    }
    expect(service.notifications().length).toBe(50);
  });

  // ── markRead() ─────────────────────────────────────────────────────────────

  it('should mark only the targeted notification as read', () => {
    service.add('task_assigned', 'First task');   // → index 1 after second add
    service.add('task_assigned', 'Second task');  // → index 0 (prepended)
    const firstTaskId = service.notifications()[1].id;
    service.markRead(firstTaskId);
    expect(service.notifications()[1].is_read).toBeTrue();
    expect(service.notifications()[0].is_read).toBeFalse();
  });

  it('should persist the read state after markRead', () => {
    service.add('task_assigned', 'Mark me read');
    const id = service.notifications()[0].id;
    service.markRead(id);
    const stored: AppNotification[] = JSON.parse(localStorage.getItem(STORAGE_KEY)!);
    expect(stored[0].is_read).toBeTrue();
  });

  // ── markAllRead() ──────────────────────────────────────────────────────────

  it('should mark all notifications as read and set unreadCount to 0', () => {
    service.add('task_assigned', 'Task 1');
    service.add('project_invite', 'Project invite');
    service.markAllRead();
    expect(service.unreadCount()).toBe(0);
    service.notifications().forEach(n => expect(n.is_read).toBeTrue());
  });

  it('should persist the fully-read state to localStorage', () => {
    service.add('task_assigned', 'Notif to read');
    service.markAllRead();
    const stored: AppNotification[] = JSON.parse(localStorage.getItem(STORAGE_KEY)!);
    expect(stored.every(n => n.is_read)).toBeTrue();
  });

  // ── checkDeadlines() ───────────────────────────────────────────────────────

  it('should add an "overdue" notification for a task with a past deadline', () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    service.checkDeadlines(
      [{ id: 1, title: 'Overdue Task', deadline: yesterday.toISOString() }], 1
    );
    expect(service.notifications().length).toBe(1);
    expect(service.notifications()[0].message).toContain('overdue');
  });

  it('should add a "due today" notification for a task due today', () => {
    const today = new Date();
    service.checkDeadlines(
      [{ id: 2, title: 'Today Task', deadline: today.toISOString() }], 1
    );
    expect(service.notifications()[0].message).toContain('due today');
  });

  it('should add a "due tomorrow" notification for a task due tomorrow', () => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    service.checkDeadlines(
      [{ id: 3, title: 'Tomorrow Task', deadline: tomorrow.toISOString() }], 1
    );
    expect(service.notifications()[0].message).toContain('due tomorrow');
  });

  it('should not add a notification for tasks with no deadline', () => {
    service.checkDeadlines([{ id: 4, title: 'No-Date Task' }], 1);
    expect(service.notifications().length).toBe(0);
  });

  it('should not add a notification for a task due more than one day away', () => {
    const nextWeek = new Date();
    nextWeek.setDate(nextWeek.getDate() + 7);
    service.checkDeadlines(
      [{ id: 5, title: 'Future Task', deadline: nextWeek.toISOString() }], 1
    );
    expect(service.notifications().length).toBe(0);
  });

  it('should set the notification link to the correct /board/:id URL', () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    service.checkDeadlines(
      [{ id: 6, title: 'Linked Task', deadline: yesterday.toISOString() }], 42
    );
    expect(service.notifications()[0].link).toBe('/board/42');
  });

  it('should include the task title in the notification message', () => {
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    service.checkDeadlines(
      [{ id: 7, title: 'My Special Task', deadline: yesterday.toISOString() }], 1
    );
    expect(service.notifications()[0].message).toContain('My Special Task');
  });

  // ── timeAgo() ─────────────────────────────────────────────────────────────

  it('should return "just now" for a timestamp less than 1 minute ago', () => {
    expect(service.timeAgo(new Date().toISOString())).toBe('just now');
  });

  it('should return minutes ago for a timestamp 30 minutes ago', () => {
    const past = new Date(Date.now() - 30 * 60_000).toISOString();
    expect(service.timeAgo(past)).toBe('30m ago');
  });

  it('should return hours ago for a timestamp 3 hours ago', () => {
    const past = new Date(Date.now() - 3 * 60 * 60_000).toISOString();
    expect(service.timeAgo(past)).toBe('3h ago');
  });

  it('should return days ago for a timestamp 2 days ago', () => {
    const past = new Date(Date.now() - 2 * 24 * 60 * 60_000).toISOString();
    expect(service.timeAgo(past)).toBe('2d ago');
  });

  // ── remove() ──────────────────────────────────────────────────────────────

  it('remove() should delete the notification with the given id', () => {
    service.add('task_assigned', 'First notification');
    service.add('deadline_reminder', 'Second notification');
    const idToRemove = service.notifications()[1].id;

    service.remove(idToRemove);

    expect(service.notifications().length).toBe(1);
    expect(service.notifications().every(n => n.id !== idToRemove)).toBeTrue();
  });

  it('remove() should persist the updated list to localStorage', () => {
    service.add('task_assigned', 'Persisted notification');
    const id = service.notifications()[0].id;

    service.remove(id);

    const stored = JSON.parse(localStorage.getItem('taskify.notifications') ?? '[]');
    expect(stored.length).toBe(0);
  });

  it('remove() should not affect other notifications', () => {
    service.add('task_assigned', 'Keep me');
    service.add('deadline_reminder', 'Remove me');
    const removeId = service.notifications()[0].id;

    service.remove(removeId);

    expect(service.notifications()[0].message).toBe('Keep me');
  });

  // ── clearAll() ────────────────────────────────────────────────────────────

  it('clearAll() should empty the notifications list', () => {
    service.add('task_assigned', 'Notification A');
    service.add('project_invite', 'Notification B');

    service.clearAll();

    expect(service.notifications().length).toBe(0);
  });

  it('clearAll() should reset unreadCount to 0', () => {
    service.add('task_assigned', 'Unread one');
    service.add('deadline_reminder', 'Unread two');

    service.clearAll();

    expect(service.unreadCount()).toBe(0);
  });

  it('clearAll() should persist empty list to localStorage', () => {
    service.add('task_assigned', 'Will be cleared');

    service.clearAll();

    const stored = JSON.parse(localStorage.getItem('taskify.notifications') ?? '["not empty"]');
    expect(stored.length).toBe(0);
  });
});
