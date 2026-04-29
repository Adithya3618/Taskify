import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { RouterTestingModule } from '@angular/router/testing';
import { of, Subject } from 'rxjs';
import { BoardComponent } from './board.component';
import { ApiService } from '../../services/api.service';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { Stage } from '../../models/stage.model';
import { Task } from '../../models/task.model';

const mockUser = { id: '1', name: 'Alice', email: 'alice@example.com' };

function makeStage(tasks: Partial<Task>[] = []): Stage {
  return {
    id: 1,
    project_id: 1,
    name: 'To Do',
    position: 0,
    created_at: '',
    updated_at: '',
    tasks: tasks as Task[],
  };
}

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: 1,
    stage_id: 1,
    title: 'Task',
    description: '',
    position: 0,
    created_at: '',
    updated_at: '',
    ...overrides,
  };
}

describe('BoardComponent', () => {
  let component: BoardComponent;
  let navigateSpy: jasmine.Spy;
  let authSpy: jasmine.SpyObj<AuthService>;
  let apiSpy: jasmine.SpyObj<ApiService>;
  let paramsSubject: Subject<{ id: string }>;

  beforeEach(async () => {
    localStorage.clear();
    // Put a valid board owner entry so canAccessBoard passes
    localStorage.setItem('taskify.board.owners', JSON.stringify({ '1': 'alice@example.com' }));

    paramsSubject = new Subject();
    authSpy   = jasmine.createSpyObj('AuthService', ['getCurrentUser', 'getKnownUsers']);
    apiSpy    = jasmine.createSpyObj('ApiService', [
      'addProjectMember',
      'createComment',
      'createStage',
      'createSubtask',
      'createTask',
      'deleteComment',
      'deleteStage',
      'deleteSubtask',
      'deleteTask',
      'getCachedProjectMembers',
      'getCachedTaskComments',
      'getComments',
      'getProject',
      'getProjectActivity',
      'getProjectMemberCount',
      'getProjectMembers',
      'getProjects',
      'getStages',
      'getSubtasks',
      'getTaskCommentCount',
      'getTasks',
      'moveTask',
      'primeTaskComments',
      'removeProjectMember',
      'seedProjectOwner',
      'setCachedProjectMembers',
      'updateComment',
      'updateSubtask',
      'updateTask',
      'userHasProjectAccess',
    ]);

    authSpy.getCurrentUser.and.returnValue(mockUser);
    authSpy.getKnownUsers.and.returnValue([mockUser]);
    apiSpy.getProject.and.returnValue(of({ id: 1, name: 'Board 1', description: '', created_at: '', updated_at: '' }));
    apiSpy.getStages.and.returnValue(of([]));
    apiSpy.getProjects.and.returnValue(of([]));
    apiSpy.userHasProjectAccess.and.returnValue(true);
    apiSpy.getProjectMemberCount.and.returnValue(1);
    apiSpy.getProjectMembers.and.returnValue(of([
      {
        project_id: 1,
        user_id: '1',
        user_name: 'Alice',
        user_email: 'alice@example.com',
        role: 'owner',
        joined_at: '',
      }
    ]));
    apiSpy.getCachedProjectMembers.and.returnValue([]);
    apiSpy.getTaskCommentCount.and.returnValue(0);
    apiSpy.getCachedTaskComments.and.returnValue([]);
    apiSpy.getComments.and.returnValue(of([]));
    apiSpy.getSubtasks.and.returnValue(of([]));
    apiSpy.getProjectActivity.and.returnValue(of({ success: true, data: [], page: 1, limit: 20, total: 0 }));

    await TestBed.configureTestingModule({
      imports: [BoardComponent, HttpClientTestingModule, RouterTestingModule.withRoutes([])],
      providers: [
        { provide: AuthService, useValue: authSpy },
        { provide: ApiService, useValue: apiSpy },
        { provide: ActivatedRoute, useValue: { params: paramsSubject.asObservable() } },
        ThemeService,
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(BoardComponent);
    component = fixture.componentInstance;
    navigateSpy = spyOn(TestBed.inject(Router), 'navigate');

    // Trigger ngOnInit then emit a route param
    fixture.detectChanges();
    paramsSubject.next({ id: '1' });
  });

  afterEach(() => {
    localStorage.clear();
    component?.ngOnDestroy();
  });

  // ── toggleFilterPanel ────────────────────────────────────
  it('should start with filter panel closed', () => {
    expect(component.showFilterPanel).toBeFalse();
  });

  it('toggleFilterPanel() should open the filter panel', () => {
    component.toggleFilterPanel();
    expect(component.showFilterPanel).toBeTrue();
  });

  it('toggleFilterPanel() should close the filter panel when called again', () => {
    component.toggleFilterPanel();
    component.toggleFilterPanel();
    expect(component.showFilterPanel).toBeFalse();
  });

  // ── hasActiveFilters ────────────────────────────────────
  it('hasActiveFilters should be false when no filters are set', () => {
    expect(component.hasActiveFilters).toBeFalse();
  });

  it('hasActiveFilters should be true when filterPriority is set', () => {
    component.filterPriority = 'High';
    expect(component.hasActiveFilters).toBeTrue();
  });

  it('hasActiveFilters should be true when filterDue is set', () => {
    component.filterDue = 'today';
    expect(component.hasActiveFilters).toBeTrue();
  });

  // ── clearFilters ────────────────────────────────────────
  it('clearFilters() should reset filterPriority and filterDue', () => {
    component.filterPriority = 'High';
    component.filterDue = 'overdue';
    component.clearFilters();
    expect(component.filterPriority).toBe('');
    expect(component.filterDue).toBe('');
  });

  // ── getFilteredTasks ────────────────────────────────────
  it('getFilteredTasks() should return all tasks when no filters are active', () => {
    const stage = makeStage([makeTask(), makeTask({ id: 2 })]);
    expect(component.getFilteredTasks(stage).length).toBe(2);
  });

  it('getFilteredTasks() should filter tasks by priority', () => {
    const highTask = makeTask({ id: 1, description: '\n---\npriority: High\ndue: \nnotes: ' });
    const lowTask  = makeTask({ id: 2, description: '\n---\npriority: Low\ndue: \nnotes: ' });
    const stage = makeStage([highTask, lowTask]);
    component.filterPriority = 'High';
    const result = component.getFilteredTasks(stage);
    expect(result.length).toBe(1);
    expect(result[0].id).toBe(1);
  });

  it('getFilteredTasks() should return empty array when no tasks match priority filter', () => {
    const lowTask = makeTask({ description: '\n---\npriority: Low\ndue: \nnotes: ' });
    const stage = makeStage([lowTask]);
    component.filterPriority = 'Critical';
    expect(component.getFilteredTasks(stage).length).toBe(0);
  });

  it('getFilteredTasks() should filter by "none" due date (tasks with no due date)', () => {
    const noDue   = makeTask({ id: 1, description: '\n---\npriority: \ndue: \nnotes: ' });
    const withDue = makeTask({ id: 2, description: '\n---\npriority: \ndue: 2099-12-31\nnotes: ' });
    const stage = makeStage([noDue, withDue]);
    component.filterDue = 'none';
    const result = component.getFilteredTasks(stage);
    expect(result.length).toBe(1);
    expect(result[0].id).toBe(1);
  });

  // ── task search ─────────────────────────────────────────
  it('getFilteredTasks() should filter tasks by title search', () => {
    const stage = makeStage([
      makeTask({ id: 1, title: 'Write release notes' }),
      makeTask({ id: 2, title: 'Fix login bug' }),
    ]);
    component.activeSearchQuery = 'release';

    const result = component.getFilteredTasks(stage);

    expect(result.length).toBe(1);
    expect(result[0].id).toBe(1);
  });

  it('getFilteredTasks() should filter tasks by description search', () => {
    const stage = makeStage([
      makeTask({ id: 1, title: 'Docs', description: 'Update onboarding checklist' }),
      makeTask({ id: 2, title: 'Bug', description: 'Investigate auth timeout' }),
    ]);
    component.activeSearchQuery = 'auth';

    const result = component.getFilteredTasks(stage);

    expect(result.length).toBe(1);
    expect(result[0].id).toBe(2);
  });

  it('task search should be case-insensitive and trim whitespace', () => {
    const stage = makeStage([
      makeTask({ id: 1, title: 'Sprint Demo Script' }),
      makeTask({ id: 2, title: 'Backend contract review' }),
    ]);
    component.activeSearchQuery = '  demo  ';

    const result = component.getFilteredTasks(stage);

    expect(result.length).toBe(1);
    expect(result[0].id).toBe(1);
  });

  it('totalMatchCount should count matching tasks across stages', () => {
    component.stages = [
      makeStage([
        makeTask({ id: 1, title: 'Search API' }),
        makeTask({ id: 2, title: 'Calendar polish' }),
      ]),
      { ...makeStage([makeTask({ id: 3, title: 'Search UI' })]), id: 2, name: 'Done' },
    ];
    component.activeSearchQuery = 'search';

    expect(component.totalMatchCount).toBe(2);
    expect(component.hasVisibleTasks).toBeTrue();
  });

  it('clearFilters() should clear task search state', () => {
    component.searchQuery = 'release';
    component.activeSearchQuery = 'release';

    component.clearFilters();

    expect(component.searchQuery).toBe('');
    expect(component.activeSearchQuery).toBe('');
  });

  // ── openShareModal / closeShareModal ────────────────────
  it('openShareModal() should set showShareModal to true', () => {
    component.openShareModal();
    expect(component.showShareModal).toBeTrue();
  });

  it('openShareModal() should reset shareLinkCopied to false', () => {
    component.shareLinkCopied = true;
    component.openShareModal();
    expect(component.shareLinkCopied).toBeFalse();
  });

  it('closeShareModal() should set showShareModal to false', () => {
    component.openShareModal();
    component.closeShareModal();
    expect(component.showShareModal).toBeFalse();
  });

  it('closeShareModal() should reset shareLinkCopied to false', () => {
    component.shareLinkCopied = true;
    component.closeShareModal();
    expect(component.shareLinkCopied).toBeFalse();
  });

  // ── boardUrl ────────────────────────────────────────────
  it('boardUrl should return the current window location href', () => {
    expect(component.boardUrl).toBe(window.location.href);
  });

  // ── toggleBoardSwitcher ─────────────────────────────────
  it('toggleBoardSwitcher() should open the board switcher', () => {
    component.toggleBoardSwitcher();
    expect(component.showBoardSwitcher).toBeTrue();
  });

  it('toggleBoardSwitcher() should close the board switcher when called again', () => {
    component.toggleBoardSwitcher();
    component.toggleBoardSwitcher();
    expect(component.showBoardSwitcher).toBeFalse();
  });

  // ── switchBoard ─────────────────────────────────────────
  it('switchBoard() should navigate to the new board', () => {
    component.projectId = 1;
    component.switchBoard(2);
    expect(navigateSpy).toHaveBeenCalledWith(['/board', 2]);
  });

  it('switchBoard() should not navigate when selecting the current board', () => {
    component.projectId = 1;
    component.switchBoard(1);
    expect(navigateSpy).not.toHaveBeenCalled();
  });

  it('switchBoard() should close the board switcher', () => {
    component.showBoardSwitcher = true;
    component.switchBoard(2);
    expect(component.showBoardSwitcher).toBeFalse();
  });

  // ── goBack ──────────────────────────────────────────────
  it('goBack() should navigate to /boards', () => {
    component.goBack();
    expect(navigateSpy).toHaveBeenCalledWith(['/boards']);
  });

  // ── viewMode ─────────────────────────────────────────────
  it('viewMode should default to dashboard', () => {
    expect(component.viewMode).toBe('dashboard');
  });

  it('setView("dashboard") should switch viewMode to dashboard', () => {
    component.setView('dashboard');
    expect(component.viewMode).toBe('dashboard');
  });

  it('setView("table") should switch viewMode to table', () => {
    component.setView('table');
    expect(component.viewMode).toBe('table');
  });

  it('setView("timeline") should switch viewMode to timeline', () => {
    component.setView('timeline');
    expect(component.viewMode).toBe('timeline');
  });

  it('setView("kanban") should switch back to kanban', () => {
    component.setView('dashboard');
    component.setView('kanban');
    expect(component.viewMode).toBe('kanban');
  });

  // ── dashTotalTasks ────────────────────────────────────────
  it('dashTotalTasks should return 0 when no stages loaded', () => {
    component.stages = [];
    expect(component.dashTotalTasks).toBe(0);
  });

  it('dashTotalTasks should sum tasks across all stages', () => {
    component.stages = [
      makeStage([makeTask(), makeTask()]),
      makeStage([makeTask()])
    ];
    expect(component.dashTotalTasks).toBe(3);
  });

  it('dashCompletedTasks should count only completed tasks', () => {
    component.stages = [
      makeStage([makeTask({ completed: true }), makeTask({ completed: false })]),
      makeStage([makeTask({ completed: true })])
    ];
    expect(component.dashCompletedTasks).toBe(2);
  });

  it('dashStagePercent should return 0 when no tasks exist', () => {
    component.stages = [];
    const stage = makeStage([]);
    expect(component.dashStagePercent(stage)).toBe(0);
  });

  it('timelineTasks should return empty array when no tasks have due dates', () => {
    component.stages = [makeStage([makeTask()])];
    expect(component.timelineTasks.length).toBe(0);
  });
});
