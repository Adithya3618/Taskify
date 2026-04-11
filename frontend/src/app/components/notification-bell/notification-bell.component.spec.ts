import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';
import { NotificationBellComponent } from './notification-bell.component';
import { NotificationService } from '../../services/notification.service';

describe('NotificationBellComponent', () => {
  let component: NotificationBellComponent;
  let fixture: ComponentFixture<NotificationBellComponent>;
  let notifService: NotificationService;
  let router: Router;

  beforeEach(async () => {
    localStorage.clear();
    await TestBed.configureTestingModule({
      imports: [NotificationBellComponent, RouterTestingModule],
    }).compileComponents();

    fixture = TestBed.createComponent(NotificationBellComponent);
    component = fixture.componentInstance;
    notifService = TestBed.inject(NotificationService);
    router = TestBed.inject(Router);
    fixture.detectChanges();
  });

  afterEach(() => {
    localStorage.clear();
  });

  // ── Creation ───────────────────────────────────────────────────────────────

  it('should create the component', () => {
    expect(component).toBeTruthy();
  });

  // ── Panel toggle ───────────────────────────────────────────────────────────

  it('should start with the dropdown panel closed', () => {
    expect(component.isOpen).toBeFalse();
  });

  it('toggle() should open the panel', () => {
    component.toggle();
    expect(component.isOpen).toBeTrue();
  });

  it('toggle() should close the panel when called again', () => {
    component.toggle();
    component.toggle();
    expect(component.isOpen).toBeFalse();
  });

  it('close() should set isOpen to false', () => {
    component.isOpen = true;
    component.close();
    expect(component.isOpen).toBeFalse();
  });

  // ── markAllRead() ──────────────────────────────────────────────────────────

  it('markAllRead() should delegate to notifService and set unreadCount to 0', () => {
    notifService.add('task_assigned', 'A task has been assigned');
    notifService.add('project_invite', 'You are invited to a project');
    expect(notifService.unreadCount()).toBe(2);

    component.markAllRead();

    expect(notifService.unreadCount()).toBe(0);
  });

  // ── onNotificationClick() ─────────────────────────────────────────────────

  it('onNotificationClick() should mark the notification as read', () => {
    notifService.add('task_assigned', 'Assigned to you');
    const notif = notifService.notifications()[0];
    expect(notif.is_read).toBeFalse();

    component.onNotificationClick(notif);

    expect(notifService.notifications()[0].is_read).toBeTrue();
  });

  it('onNotificationClick() should close the panel', () => {
    notifService.add('task_assigned', 'Another task');
    component.isOpen = true;

    component.onNotificationClick(notifService.notifications()[0]);

    expect(component.isOpen).toBeFalse();
  });

  it('onNotificationClick() should navigate to the notification link when present', () => {
    const navigateSpy = spyOn(router, 'navigateByUrl');
    notifService.add('deadline_reminder', '"Task X" is overdue!', '/board/5');
    const notif = notifService.notifications()[0];

    component.onNotificationClick(notif);

    expect(navigateSpy).toHaveBeenCalledWith('/board/5');
  });

  it('onNotificationClick() should not navigate when the notification has no link', () => {
    const navigateSpy = spyOn(router, 'navigateByUrl');
    notifService.add('task_assigned', 'No link here');

    component.onNotificationClick(notifService.notifications()[0]);

    expect(navigateSpy).not.toHaveBeenCalled();
  });

  // ── iconForType() ──────────────────────────────────────────────────────────

  it('iconForType() should return clock emoji for deadline_reminder', () => {
    expect(component.iconForType('deadline_reminder')).toBe('⏰');
  });

  it('iconForType() should return clipboard emoji for task_assigned', () => {
    expect(component.iconForType('task_assigned')).toBe('📋');
  });

  it('iconForType() should return bell emoji for project_invite', () => {
    expect(component.iconForType('project_invite')).toBe('🔔');
  });
});
