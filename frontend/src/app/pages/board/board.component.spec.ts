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
    authSpy   = jasmine.createSpyObj('AuthService', ['getCurrentUser']);
    apiSpy    = jasmine.createSpyObj('ApiService', [
      'getProject', 'getStages', 'getTasks', 'getProjects',
    ]);

    authSpy.getCurrentUser.and.returnValue(mockUser);
    apiSpy.getProject.and.returnValue(of({ id: 1, name: 'Board 1', description: '', created_at: '', updated_at: '' }));
    apiSpy.getStages.and.returnValue(of([]));
    apiSpy.getProjects.and.returnValue(of([]));

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
});
