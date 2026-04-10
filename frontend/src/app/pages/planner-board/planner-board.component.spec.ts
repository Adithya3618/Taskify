import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { By } from '@angular/platform-browser';
import { ActivatedRoute, Router } from '@angular/router';
import { RouterTestingModule } from '@angular/router/testing';
import { of, Subject } from 'rxjs';
import { PlannerBoardComponent, TaskWithStage } from './planner-board.component';
import { ApiService } from '../../services/api.service';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { TaskCompletionStorageService } from '../../services/task-completion-storage.service';
import { buildCardDescription } from '../../utils/task-card-meta';
import { Stage } from '../../models/stage.model';
import { Task } from '../../models/task.model';

const mockUser = { id: '1', name: 'Alice', email: 'alice@example.com' };

function makeStage(id: number, name: string): Stage {
  return {
    id,
    project_id: 1,
    name,
    position: id,
    created_at: '',
    updated_at: '',
    tasks: [],
  };
}

function makeTaskWithStage(overrides: Partial<TaskWithStage> = {}): TaskWithStage {
  return {
    id: 1,
    stage_id: 1,
    title: 'Task',
    description: '',
    position: 0,
    created_at: '',
    updated_at: '',
    stageName: 'To Do',
    stageId: 1,
    ...overrides,
  };
}

/** Push tasks into planner buckets (mirrors loadTasks → applyTaskBuckets). */
function seedTasks(component: PlannerBoardComponent, tasks: TaskWithStage[]): void {
  const c = component as unknown as {
    allTasks: TaskWithStage[];
    applyTaskBuckets: () => void;
    buildMonthWeeks: (d: Date) => Date[][];
    calendarWeeks: Date[][];
  };
  c.allTasks = tasks;
  c.applyTaskBuckets();
  c.calendarWeeks = c.buildMonthWeeks(component.viewMonth);
}

describe('PlannerBoardComponent', () => {
  let component: PlannerBoardComponent;
  let fixture: ComponentFixture<PlannerBoardComponent>;
  let router: Router;
  let authSpy: jasmine.SpyObj<AuthService>;
  let apiSpy: jasmine.SpyObj<ApiService>;
  let taskStorageSpy: jasmine.SpyObj<TaskCompletionStorageService>;
  let paramsSubject: Subject<{ id: string }>;

  beforeEach(async () => {
    localStorage.clear();
    localStorage.setItem('taskify.board.owners', JSON.stringify({ '1': 'alice@example.com' }));

    paramsSubject = new Subject();
    authSpy = jasmine.createSpyObj('AuthService', ['getCurrentUser']);
    apiSpy = jasmine.createSpyObj('ApiService', [
      'getProject',
      'getStages',
      'getTasks',
      'getProjects',
      'createTask',
      'updateTask',
    ]);
    taskStorageSpy = jasmine.createSpyObj('TaskCompletionStorageService', ['getCompleted', 'setCompleted']);
    taskStorageSpy.getCompleted.and.returnValue(false);

    authSpy.getCurrentUser.and.returnValue(mockUser);
    apiSpy.getProject.and.returnValue(of({ id: 1, name: 'Board 1', description: '', created_at: '', updated_at: '' }));
    apiSpy.getStages.and.returnValue(of([makeStage(1, 'To Do'), makeStage(2, 'Doing')]));
    apiSpy.getTasks.and.callFake((_pid: number, stageId: number) => {
      if (stageId === 1) {
        return of([
          makeTaskWithStage({
            id: 10,
            title: 'With due',
            description: buildCardDescription('', '2026-06-15', 'High', ''),
            stageId: 1,
            stageName: 'To Do',
          }),
          makeTaskWithStage({
            id: 11,
            title: 'No due A',
            description: '',
            stageId: 1,
            stageName: 'To Do',
          }),
        ]);
      }
      if (stageId === 2) {
        return of([
          makeTaskWithStage({
            id: 12,
            title: 'No due B',
            description: '',
            stageId: 2,
            stageName: 'Doing',
          }),
        ]);
      }
      return of([]);
    });
    apiSpy.getProjects.and.returnValue(of([]));
    const stubTask: Task = {
      id: 99,
      stage_id: 1,
      title: 't',
      description: '',
      position: 0,
      created_at: '',
      updated_at: '',
    };
    apiSpy.createTask.and.returnValue(of(stubTask));
    apiSpy.updateTask.and.returnValue(of(stubTask));

    await TestBed.configureTestingModule({
      imports: [PlannerBoardComponent, HttpClientTestingModule, RouterTestingModule.withRoutes([])],
      providers: [
        { provide: AuthService, useValue: authSpy },
        { provide: ApiService, useValue: apiSpy },
        { provide: TaskCompletionStorageService, useValue: taskStorageSpy },
        { provide: ActivatedRoute, useValue: { params: paramsSubject.asObservable() } },
        ThemeService,
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(PlannerBoardComponent);
    component = fixture.componentInstance;
    router = TestBed.inject(Router);
    spyOn(router, 'navigate');

    fixture.detectChanges();
    paramsSubject.next({ id: '1' });
    fixture.detectChanges();
    // Do not await whenStable(): ngOnInit schedules a 5s setTimeout fallback that keeps
    // the zone unstable and causes Jasmine's 5s hook timeout. API stubs complete sync.
    fixture.detectChanges();
  });

  afterEach(() => {
    localStorage.clear();
    component.ngOnDestroy();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
    expect(component.loading).toBeFalse();
    expect(component.projectId).toBe(1);
  });

  // ── Collapsible: Scheduled (button click) ─────────────────────────────
  it('should open Scheduled panel when its header button is clicked', () => {
    expect(component.scheduledOpen).toBeFalse();
    const btn = fixture.debugElement.query(By.css('#planner-scheduled-toggle'));
    expect(btn).toBeTruthy();
    btn.nativeElement.click();
    fixture.detectChanges();
    expect(component.scheduledOpen).toBeTrue();
  });

  it('should close Scheduled panel when header button is clicked again', () => {
    component.scheduledOpen = true;
    fixture.detectChanges();
    fixture.debugElement.query(By.css('#planner-scheduled-toggle'))!.nativeElement.click();
    fixture.detectChanges();
    expect(component.scheduledOpen).toBeFalse();
  });

  it('toggleScheduledOpen() should flip scheduledOpen', () => {
    component.toggleScheduledOpen();
    expect(component.scheduledOpen).toBeTrue();
    component.toggleScheduledOpen();
    expect(component.scheduledOpen).toBeFalse();
  });

  // ── Collapsible: No due date (button click) ──────────────────────────
  it('should open No due date panel when its header button is clicked', () => {
    expect(component.noDueOpen).toBeFalse();
    fixture.debugElement.query(By.css('#planner-nodue-toggle'))!.nativeElement.click();
    fixture.detectChanges();
    expect(component.noDueOpen).toBeTrue();
  });

  it('toggleNoDueOpen() should flip noDueOpen', () => {
    component.toggleNoDueOpen();
    expect(component.noDueOpen).toBeTrue();
  });

  // ── No due date: list filter ──────────────────────────────────────────
  it('unscheduledFiltered should include all no-due tasks when filter is null', () => {
    expect(component.unscheduled.length).toBe(2);
    expect(component.unscheduledFiltered.length).toBe(2);
  });

  it('unscheduledFiltered should only include tasks from selected list', () => {
    component.noDueListFilterStageId = 2;
    expect(component.unscheduledFiltered.length).toBe(1);
    expect(component.unscheduledFiltered[0].title).toBe('No due B');

    component.noDueListFilterStageId = 1;
    expect(component.unscheduledFiltered.length).toBe(1);
    expect(component.unscheduledFiltered[0].title).toBe('No due A');
  });

  it('should show list filter select when No due date is open and stages exist', () => {
    component.noDueOpen = true;
    fixture.detectChanges();
    const select = fixture.debugElement.query(By.css('#noDueListFilter'));
    expect(select).toBeTruthy();
    expect(component.stages.length).toBeGreaterThan(0);
  });

  // ── Month navigation: header ‹ Today › (button clicks) ───────────────
  it('clicking previous month in main header should call prevMonth', () => {
    const before = new Date(component.viewMonth);
    const navBtns = fixture.debugElement.queryAll(By.css('.planner-calendar-header-nav .btn-nav-month'));
    expect(navBtns.length).toBe(2);
    navBtns[0].nativeElement.click();
    expect(component.viewMonth.getMonth()).toBe((before.getMonth() + 11) % 12);
  });

  it('clicking next month in main header should call nextMonth', () => {
    const before = new Date(component.viewMonth);
    const navBtns = fixture.debugElement.queryAll(By.css('.planner-calendar-header-nav .btn-nav-month'));
    navBtns[1].nativeElement.click();
    expect(component.viewMonth.getMonth()).toBe((before.getMonth() + 1) % 12);
  });

  it('clicking Today in main header should reset view to current month', () => {
    component.prevMonth();
    component.prevMonth();
    fixture.debugElement.query(By.css('.planner-calendar-header-nav .btn-today'))!.nativeElement.click();
    const now = new Date();
    expect(component.viewMonth.getMonth()).toBe(now.getMonth());
    expect(component.viewMonth.getFullYear()).toBe(now.getFullYear());
  });

  // ── Mini calendar ‹ › (button clicks) ─────────────────────────────────
  it('clicking first mini nav button should go to previous month', () => {
    const before = component.viewMonth.getMonth();
    const mini = fixture.debugElement.queryAll(By.css('.btn-mini-nav'));
    expect(mini.length).toBe(2);
    mini[0].nativeElement.click();
    expect(component.viewMonth.getMonth()).toBe((before + 11) % 12);
  });

  it('clicking second mini nav button should go to next month', () => {
    const before = component.viewMonth.getMonth();
    fixture.debugElement.queryAll(By.css('.btn-mini-nav'))[1].nativeElement.click();
    expect(component.viewMonth.getMonth()).toBe((before + 1) % 12);
  });

  // ── Month / year picker ───────────────────────────────────────────────
  it('openMonthYearPicker should show dialog and set draft', () => {
    component.viewMonth = new Date(2026, 2, 1);
    component.openMonthYearPicker();
    expect(component.showMonthYearPicker).toBeTrue();
    expect(component.monthYearDraft).toBe('2026-03');
  });

  it('clicking main month label button should open month/year picker', () => {
    component.showMonthYearPicker = false;
    fixture.detectChanges();
    fixture.debugElement.query(By.css('#planner-cal-title'))!.nativeElement.click();
    expect(component.showMonthYearPicker).toBeTrue();
  });

  it('closeMonthYearPicker should hide dialog', () => {
    component.showMonthYearPicker = true;
    component.closeMonthYearPicker();
    expect(component.showMonthYearPicker).toBeFalse();
  });

  it('applyMonthYear should update viewMonth for valid draft', () => {
    component.monthYearDraft = '2025-07';
    component.applyMonthYear();
    expect(component.viewMonth.getFullYear()).toBe(2025);
    expect(component.viewMonth.getMonth()).toBe(6);
    expect(component.showMonthYearPicker).toBeFalse();
  });

  // ── +N tasks labels ───────────────────────────────────────────────────
  it('extraTasksOnSameDateLabel should return +1 task for two tasks', () => {
    expect(component.extraTasksOnSameDateLabel(2)).toBe('+1 task');
  });

  it('extraTasksOnSameDateLabel should return +2 tasks for three tasks', () => {
    expect(component.extraTasksOnSameDateLabel(3)).toBe('+2 tasks');
  });

  it('calendarDayToggleLabel should show +N tasks when multiple tasks and collapsed', () => {
    component.viewMonth = new Date(2026, 5, 1);
    seedTasks(component, [
      makeTaskWithStage({
        id: 1,
        title: 'A',
        description: buildCardDescription('', '2026-06-10', '', ''),
        stageId: 1,
        stageName: 'To Do',
      }),
      makeTaskWithStage({
        id: 2,
        title: 'B',
        description: buildCardDescription('', '2026-06-10', '', ''),
        stageId: 1,
        stageName: 'To Do',
      }),
      makeTaskWithStage({
        id: 3,
        title: 'C',
        description: buildCardDescription('', '2026-06-10', '', ''),
        stageId: 1,
        stageName: 'To Do',
      }),
    ]);
    const day = new Date(2026, 5, 10);
    expect(component.calendarDayToggleLabel(day)).toBe('+2 tasks');
    component.calendarDateExpanded['2026-06-10'] = true;
    expect(component.calendarDayToggleLabel(day)).toBe('Show less');
  });

  it('tasksForCalendarCell should return one task until expanded', () => {
    component.viewMonth = new Date(2026, 7, 1);
    seedTasks(component, [
      makeTaskWithStage({
        id: 1,
        title: 'A',
        description: buildCardDescription('', '2026-08-05', 'High', ''),
        stageId: 1,
        stageName: 'To Do',
      }),
      makeTaskWithStage({
        id: 2,
        title: 'B',
        description: buildCardDescription('', '2026-08-05', 'Low', ''),
        stageId: 1,
        stageName: 'To Do',
      }),
    ]);
    const day = new Date(2026, 7, 5);
    expect(component.tasksForCalendarCell(day).length).toBe(1);
    expect(component.tasksForCalendarCell(day)[0].title).toBe('A');
    const ev = new Event('click');
    spyOn(ev, 'stopPropagation');
    component.toggleCalendarDayExpand(day, ev);
    expect(component.tasksForCalendarCell(day).length).toBe(2);
    expect(ev.stopPropagation).toHaveBeenCalled();
  });

  // ── Scheduled group expand ──────────────────────────────────────────────
  it('visibleScheduledTasks should show one task until group expanded', () => {
    component.viewMonth = new Date(2026, 4, 1);
    seedTasks(component, [
      makeTaskWithStage({
        id: 1,
        title: 'X',
        description: buildCardDescription('', '2026-05-20', '', ''),
        stageId: 1,
        stageName: 'To Do',
      }),
      makeTaskWithStage({
        id: 2,
        title: 'Y',
        description: buildCardDescription('', '2026-05-20', '', ''),
        stageId: 1,
        stageName: 'To Do',
      }),
    ]);
    const g = component.scheduledGroups.find((x) => x.dateKey === '2026-05-20')!;
    expect(g.tasks.length).toBe(2);
    expect(component.visibleScheduledTasks(g).length).toBe(1);
    const ev = new Event('click');
    spyOn(ev, 'stopPropagation');
    component.toggleScheduledGroupExpand('2026-05-20', ev);
    expect(component.visibleScheduledTasks(g).length).toBe(2);
    expect(component.scheduledGroupExpandLabel(g)).toBe('Show less');
  });

  // ── Board switcher ─────────────────────────────────────────────────────
  it('toggleBoardSwitcher should flip showBoardSwitcher', () => {
    expect(component.showBoardSwitcher).toBeFalse();
    component.toggleBoardSwitcher();
    expect(component.showBoardSwitcher).toBeTrue();
  });

  it('clicking theme toggle button should call themeService.toggle', () => {
    const theme = TestBed.inject(ThemeService);
    spyOn(theme, 'toggle');
    fixture.debugElement.query(By.css('.themeToggle'))!.nativeElement.click();
    expect(theme.toggle).toHaveBeenCalled();
  });

  it('goBack should navigate to /boards', () => {
    (router.navigate as jasmine.Spy).calls.reset();
    component.goBack();
    expect(router.navigate).toHaveBeenCalledWith(['/boards']);
  });

  // ── Day cell click ────────────────────────────────────────────────────
  it('onDayCellClick on current-month empty cell should open add-task modal', () => {
    component.viewMonth = new Date(2026, 5, 1);
    component.showAddTaskModal = false;
    const day = new Date(2026, 5, 20);
    const ev = { target: document.createElement('div') } as unknown as MouseEvent;
    component.onDayCellClick(day, ev);
    expect(component.showAddTaskModal).toBeTrue();
    expect(component.newTaskDue).toBe('2026-06-20');
  });

  it('onDayCellClick on other-month cell should change view month', () => {
    component.viewMonth = new Date(2026, 5, 1);
    const day = new Date(2026, 6, 4);
    const ev = { target: document.createElement('div') } as unknown as MouseEvent;
    component.onDayCellClick(day, ev);
    expect(component.viewMonth.getMonth()).toBe(6);
  });

  // ── Helpers ────────────────────────────────────────────────────────────
  it('formatDueLabel should format YYYY-MM-DD', () => {
    expect(component.formatDueLabel('2026-03-22')).toMatch(/2026/);
    expect(component.formatDueLabel('2026-03-22')).toMatch(/22/);
  });

  it('should redirect to login when user is missing', async () => {
    authSpy.getCurrentUser.and.returnValue(null);
    (router.navigate as jasmine.Spy).calls.reset();
    const fix = TestBed.createComponent(PlannerBoardComponent);
    fix.detectChanges();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
    fix.destroy();
    authSpy.getCurrentUser.and.returnValue(mockUser);
  });
});
