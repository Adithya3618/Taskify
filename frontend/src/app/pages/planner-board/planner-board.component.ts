import { ChangeDetectorRef, Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink, RouterLinkActive } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { forkJoin, of, Subscription } from 'rxjs';
import { catchError, map, switchMap } from 'rxjs/operators';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';
import { Stage } from '../../models/stage.model';
import { CreateTaskRequest, Task } from '../../models/task.model';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { TaskCompletionStorageService } from '../../services/task-completion-storage.service';
import { buildCardDescription, parseCardMeta, parseDueToDateKey } from '../../utils/task-card-meta';

export interface TaskWithStage extends Task {
  stageName: string;
  stageId: number;
}

@Component({
  selector: 'app-planner-board',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive, FormsModule],
  templateUrl: './planner-board.component.html',
  styleUrls: ['./planner-board.component.scss'],
})
export class PlannerBoardComponent implements OnInit, OnDestroy {
  private readonly boardOwnersKey = 'taskify.board.owners';
  projectId = 0;
  project: Project | null = null;
  loading = true;
  userDisplayName = '';
  userEmail = '';
  allBoards: Project[] = [];
  showBoardSwitcher = false;

  /** Lists (columns) for add-task and counts */
  stages: Stage[] = [];

  /** Current month being shown */
  viewMonth = new Date();

  /** 6×7 grid of dates (weeks start Sunday) */
  calendarWeeks: Date[][] = [];

  /** Tasks keyed by YYYY-MM-DD */
  tasksByDate = new Map<string, TaskWithStage[]>();
  unscheduled: TaskWithStage[] = [];

  /** Jump to month/year */
  showMonthYearPicker = false;
  monthYearDraft = '';

  /** Add task (same fields as board column form) */
  showAddTaskModal = false;
  addTaskStageId = 0;
  newTaskTitle = '';
  newTaskDesc = '';
  newTaskDue = '';
  newTaskPriority = '';
  newTaskNotes = '';
  showAddTaskDetails = false;

  /** Task detail modal (inline, like board) */
  detailTask: TaskWithStage | null = null;
  detailStageName = '';
  detailTitle = '';
  detailDesc = '';
  detailDue = '';
  detailPriority = '';
  detailNotes = '';
  detailCompleted = false;

  private routeSub?: Subscription;
  private allTasks: TaskWithStage[] = [];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private apiService: ApiService,
    private authService: AuthService,
    public themeService: ThemeService,
    private taskCompletionStorage: TaskCompletionStorageService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    const user = this.authService.getCurrentUser();
    if (!user) {
      this.router.navigate(['/login']);
      return;
    }
    this.userDisplayName = user.name;
    this.userEmail = user.email;
    this.loadAllBoards();

    this.routeSub = this.route.params.subscribe((params) => {
      const id = params['id'];
      if (!id) {
        this.router.navigate(['/']);
        return;
      }
      this.projectId = +id;
      if (!this.canAccessBoard(this.projectId, user.email)) {
        this.router.navigate(['/boards']);
        return;
      }
      this.showBoardSwitcher = false;
      this.loadProject();
    });

    setTimeout(() => {
      if (this.loading) this.loading = false;
    }, 5000);
  }

  ngOnDestroy(): void {
    this.routeSub?.unsubscribe();
  }

  private loadAllBoards(): void {
    this.apiService.getProjects().subscribe({
      next: (projects) => {
        const email = this.userEmail.trim().toLowerCase();
        const raw = localStorage.getItem(this.boardOwnersKey);
        const owners: Record<string, string> = raw ? JSON.parse(raw) : {};
        this.allBoards = (projects || []).filter((p) => owners[String(p.id)] === email);
      },
      error: () => {
        this.allBoards = [];
      },
    });
  }

  toggleBoardSwitcher(): void {
    this.showBoardSwitcher = !this.showBoardSwitcher;
  }

  switchBoard(id: number): void {
    if (id === this.projectId) {
      this.showBoardSwitcher = false;
      return;
    }
    this.showBoardSwitcher = false;
    this.router.navigate(['/board', id, 'planner']);
  }

  private canAccessBoard(projectId: number, email: string): boolean {
    try {
      const raw = localStorage.getItem(this.boardOwnersKey);
      const owners = raw ? (JSON.parse(raw) as Record<string, string>) : {};
      return owners[String(projectId)] === email.trim().toLowerCase();
    } catch {
      return false;
    }
  }

  private loadProject(): void {
    this.loading = true;
    this.apiService.getProject(this.projectId).subscribe({
      next: (p) => {
        this.project = p;
        this.loadTasks();
      },
      error: () => {
        this.project = {
          id: this.projectId,
          name: `Board ${this.projectId}`,
          description: '',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        this.loadTasks();
      },
    });
  }

  private loadTasks(): void {
    this.apiService
      .getStages(this.projectId)
      .pipe(
        switchMap((stages: Stage[]) => {
          const list = stages || [];
          this.stages = list.map((s) => ({ ...s, tasks: s.tasks ?? [] }));
          if (!list.length) {
            return of([] as TaskWithStage[]);
          }
          return forkJoin(
            list.map((stage) =>
              this.apiService.getTasks(this.projectId, stage.id).pipe(
                map((tasks) =>
                  (tasks || []).map(
                    (t): TaskWithStage => ({
                      ...t,
                      stageName: stage.name,
                      stageId: stage.id,
                    })
                  )
                ),
                catchError(() => of([] as TaskWithStage[]))
              )
            )
          ).pipe(map((chunks) => chunks.flat()));
        })
      )
      .subscribe({
        next: (rows) => {
          this.allTasks = rows;
          this.applyTaskBuckets();
          this.calendarWeeks = this.buildMonthWeeks(this.viewMonth);
          this.loading = false;
          this.cdr.markForCheck();
        },
        error: () => {
          this.allTasks = [];
          this.applyTaskBuckets();
          this.calendarWeeks = this.buildMonthWeeks(this.viewMonth);
          this.loading = false;
        },
      });
  }

  /** Bucket tasks by due date (same meta format as the board). */
  private applyTaskBuckets(): void {
    this.tasksByDate.clear();
    this.unscheduled = [];

    for (const t of this.allTasks) {
      const due = parseCardMeta(t.description || '').due;
      const key = parseDueToDateKey(due);
      if (!key) {
        this.unscheduled.push(t);
        continue;
      }
      const list = this.tasksByDate.get(key) ?? [];
      list.push(t);
      this.tasksByDate.set(key, list);
    }

    for (const [, list] of this.tasksByDate) {
      list.sort((a, b) => a.title.localeCompare(b.title));
    }
    this.unscheduled.sort((a, b) => a.title.localeCompare(b.title));
  }

  private buildMonthWeeks(anchor: Date): Date[][] {
    const year = anchor.getFullYear();
    const month = anchor.getMonth();
    const first = new Date(year, month, 1);
    const start = new Date(first);
    start.setDate(first.getDate() - first.getDay());

    const weeks: Date[][] = [];
    const cur = new Date(start);
    for (let w = 0; w < 6; w++) {
      const row: Date[] = [];
      for (let d = 0; d < 7; d++) {
        row.push(new Date(cur));
        cur.setDate(cur.getDate() + 1);
      }
      weeks.push(row);
    }
    return weeks;
  }

  prevMonth(): void {
    const d = new Date(this.viewMonth);
    d.setMonth(d.getMonth() - 1);
    this.viewMonth = d;
    this.calendarWeeks = this.buildMonthWeeks(this.viewMonth);
  }

  nextMonth(): void {
    const d = new Date(this.viewMonth);
    d.setMonth(d.getMonth() + 1);
    this.viewMonth = d;
    this.calendarWeeks = this.buildMonthWeeks(this.viewMonth);
  }

  goToday(): void {
    this.viewMonth = new Date();
    this.calendarWeeks = this.buildMonthWeeks(this.viewMonth);
  }

  /** Show native month picker dialog */
  openMonthYearPicker(): void {
    const y = this.viewMonth.getFullYear();
    const m = String(this.viewMonth.getMonth() + 1).padStart(2, '0');
    this.monthYearDraft = `${y}-${m}`;
    this.showMonthYearPicker = true;
  }

  closeMonthYearPicker(): void {
    this.showMonthYearPicker = false;
  }

  applyMonthYear(): void {
    const m = this.monthYearDraft?.match(/^(\d{4})-(\d{2})$/);
    if (m) {
      const y = +m[1];
      const mo = +m[2] - 1;
      if (!Number.isNaN(y) && mo >= 0 && mo <= 11) {
        this.viewMonth = new Date(y, mo, 1);
        this.calendarWeeks = this.buildMonthWeeks(this.viewMonth);
      }
    }
    this.showMonthYearPicker = false;
  }

  /** Clicking a day outside the visible month jumps to that month (Trello-style). */
  goToMonthContaining(day: Date): void {
    this.viewMonth = new Date(day.getFullYear(), day.getMonth(), 1);
    this.calendarWeeks = this.buildMonthWeeks(this.viewMonth);
  }

  /**
   * Click empty day cell: if day is in another month, navigate there; else open add-task for that date.
   * Task chips stop propagation so they open the detail modal instead.
   */
  onDayCellClick(day: Date, event: MouseEvent): void {
    if ((event.target as HTMLElement).closest('.planner-task-chip')) {
      return;
    }
    if (!this.isCurrentMonth(day)) {
      this.goToMonthContaining(day);
      return;
    }
    this.openAddTaskForDay(day);
  }

  openAddTaskForDay(day: Date): void {
    this.newTaskDue = this.dateKey(day);
    this.newTaskTitle = '';
    this.newTaskDesc = '';
    this.newTaskPriority = '';
    this.newTaskNotes = '';
    this.showAddTaskDetails = false;
    this.addTaskStageId = this.stages[0]?.id ?? 0;
    this.showAddTaskModal = true;
  }

  closeAddTaskModal(): void {
    this.showAddTaskModal = false;
  }

  toggleAddTaskDetails(): void {
    this.showAddTaskDetails = !this.showAddTaskDetails;
  }

  getCreatePriorityClass(): string {
    const priority = (this.newTaskPriority || '').toLowerCase();
    if (priority === 'critical' || priority === 'high') return 'priority-high';
    if (priority === 'medium') return 'priority-mid';
    if (priority === 'low') return 'priority-low';
    return 'priority-none';
  }

  createTaskFromPlanner(): void {
    const title = this.newTaskTitle?.trim();
    if (!title || !this.addTaskStageId) return;

    const position = this.allTasks.filter((t) => t.stageId === this.addTaskStageId).length;
    const description = buildCardDescription(
      this.newTaskDesc || '',
      this.newTaskDue || '',
      this.newTaskPriority || '',
      this.newTaskNotes || ''
    );
    const request: CreateTaskRequest = { title, description, position };

    this.apiService.createTask(this.projectId, this.addTaskStageId, request).subscribe({
      next: () => {
        this.closeAddTaskModal();
        this.loadTasks();
      },
      error: () => {
        this.closeAddTaskModal();
        this.loadTasks();
      },
    });
  }

  openTaskDetail(task: TaskWithStage, event?: Event): void {
    event?.stopPropagation();
    const parsed = parseCardMeta(task.description || '');
    this.detailTask = task;
    this.detailStageName = task.stageName;
    this.detailTitle = task.title || '';
    this.detailDesc = parsed.desc;
    this.detailDue = parsed.due;
    this.detailPriority = parsed.priority;
    this.detailNotes = parsed.notes;
    this.detailCompleted =
      task.completed ?? this.taskCompletionStorage.getCompleted(this.projectId, task.id);
  }

  closeTaskDetail(): void {
    this.detailTask = null;
  }

  saveTaskDetail(): void {
    if (!this.detailTask) return;
    const task = this.detailTask;
    const updatedTitle = this.detailTitle.trim() || task.title;
    const updatedDescription = buildCardDescription(
      this.detailDesc,
      this.detailDue,
      this.detailPriority,
      this.detailNotes
    );

    this.taskCompletionStorage.setCompleted(this.projectId, task.id, this.detailCompleted);
    task.completed = this.detailCompleted;

    this.apiService
      .updateTask(task.id, {
        title: updatedTitle,
        description: updatedDescription,
        position: task.position,
      })
      .subscribe({
        next: () => {
          this.closeTaskDetail();
          this.loadTasks();
        },
        error: () => {
          task.title = updatedTitle;
          task.description = updatedDescription;
          this.closeTaskDetail();
          this.applyTaskBuckets();
        },
      });
  }

  monthLabel(): string {
    return this.viewMonth.toLocaleDateString(undefined, { month: 'long', year: 'numeric' });
  }

  weekdayLabels = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  dateKey(d: Date): string {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }

  isCurrentMonth(d: Date): boolean {
    return d.getMonth() === this.viewMonth.getMonth() && d.getFullYear() === this.viewMonth.getFullYear();
  }

  isToday(d: Date): boolean {
    const t = new Date();
    return (
      d.getDate() === t.getDate() &&
      d.getMonth() === t.getMonth() &&
      d.getFullYear() === t.getFullYear()
    );
  }

  tasksForDay(d: Date): TaskWithStage[] {
    return this.tasksByDate.get(this.dateKey(d)) ?? [];
  }

  isTaskDone(task: Task): boolean {
    return !!(task.completed ?? this.taskCompletionStorage.getCompleted(this.projectId, task.id));
  }

  getTaskPriority(task: Task): string {
    return parseCardMeta(task.description || '').priority;
  }

  getPriorityClass(task: Task): string {
    const priority = this.getTaskPriority(task).toLowerCase();
    if (priority === 'critical' || priority === 'high' || priority === 'highest') return 'priority-high';
    if (priority === 'medium' || priority === 'mid') return 'priority-mid';
    if (priority === 'low' || priority === 'lowest') return 'priority-low';
    return 'priority-none';
  }

  /** Same behavior as board task cards (localStorage completion). */
  toggleTaskCompleted(task: TaskWithStage, event: Event): void {
    event.stopPropagation();
    const input = event.target as HTMLInputElement;
    const next = input.checked;
    this.taskCompletionStorage.setCompleted(this.projectId, task.id, next);
    task.completed = next;
    this.cdr.detectChanges();
  }

  goBack(): void {
    this.router.navigate(['/boards']);
  }
}
