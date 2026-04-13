import { ChangeDetectorRef, Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink, RouterLinkActive } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { Subscription } from 'rxjs';
import { CdkDragDrop, DragDropModule, moveItemInArray, transferArrayItem } from '@angular/cdk/drag-drop';
import { ApiService } from '../../services/api.service';
import { ActivityLog } from '../../models/activity.model';
import { AddProjectMemberRequest, Project, ProjectMember } from '../../models/project.model';
import { Comment, CreateCommentRequest } from '../../models/comment.model';
import { Stage, CreateStageRequest } from '../../models/stage.model';
import { Task, CreateTaskRequest } from '../../models/task.model';
import { Label } from '../../models/label.model';
import { Subtask } from '../../models/subtask.model';
import { AuthService, AuthUser } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { TaskCompletionStorageService } from '../../services/task-completion-storage.service';
import { NotificationService } from '../../services/notification.service';
import { NotificationBellComponent } from '../../components/notification-bell/notification-bell.component';

@Component({
  selector: 'app-board',
  standalone: true,
  imports: [CommonModule, FormsModule, DragDropModule, RouterLink, RouterLinkActive, NotificationBellComponent],
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss']
})
export class BoardComponent implements OnInit, OnDestroy {
  projectId: number = 0;
  project: Project | null = null;
  stages: Stage[] = [];
  loading = true;
  private readonly boardOwnersKey = 'taskify.board.owners';
  private readonly META_SEP = '\n---\n';
  userDisplayName = '';
  userEmail = '';
  showProfileMenu = false;
  showAccountModal = false;
  pendingEmail = '';
  currentPassword = '';
  newPassword = '';
  confirmPassword = '';
  showCurrentPassword = false;
  showNewPassword = false;
  showConfirmPassword = false;
  accountError = '';
  accountSuccess = '';

  // New item inputs
  newStageName = '';
  newTaskTitles: { [key: number]: string } = {};
  newTaskDescs: { [key: number]: string } = {};
  newTaskDues: { [key: number]: string } = {};
  newTaskPriorities: { [key: number]: string } = {};
  newTaskNotes: { [key: number]: string } = {};
  newTaskLabels: { [key: number]: string[] } = {};
  showTaskDetails: { [key: number]: boolean } = {};

  // Board switcher
  allBoards: Project[] = [];
  showBoardSwitcher = false;
  private routeSub?: Subscription;

  /** Collapsed stage columns (narrow vertical strip with vertical title). Persisted per project in sessionStorage. */
  collapsedStages: Record<number, boolean> = {};

  // Filter state
  showFilterPanel = false;
  /** '' = all, 'active' = not completed, 'done' = completed */
  filterCompletion = '';
  filterPriority = '';
  filterDue = '';
  filterLabel = '';

  // Labels
  projectLabels: Label[] = [];
  showLabelManager = false;
  newLabelName = '';
  newLabelColor = '#818CF8';
  taskLabelMap: Record<number, string[]> = {};
  readonly LABEL_COLORS = [
    '#818CF8', '#6366f1', '#22D3EE', '#34D399',
    '#FBBF24', '#F87171', '#E879F9', '#F59E0B',
    '#10B981', '#60A5FA', '#FB923C', '#A78BFA'
  ];

  // Share state
  showShareModal = false;
  shareLinkCopied = false;
  showProjectSettingsModal = false;
  settingsActiveTab: 'general' | 'members' | 'activity' = 'general';
  projectMembers: ProjectMember[] = [];
  membersLoading = false;
  addMemberQuery = '';
  memberError = '';
  memberSuccess = '';
  addingMember = false;
  removingMemberId: string | null = null;
  pendingMemberRemoval: { member: ProjectMember; timeoutId: ReturnType<typeof setTimeout> } | null = null;
  knownUsers: AuthUser[] = [];
  activityLogs: ActivityLog[] = [];
  activityLoading = false;
  activityError = '';
  activityPage = 1;
  activityLimit = 12;
  activityTotal = 0;
  activityHasLoaded = false;
  activityLoadingMore = false;
  activityFilterUserId = '';
  activityFilterFrom = '';
  activityFilterTo = '';

  /** Delete task confirmation (replaces window.confirm). */
  deleteTaskPending: { stageId: number; taskId: number; taskTitle: string } | null = null;

  /** Delete list (stage) confirmation — replaces window.confirm on column ×. */
  deleteStagePending: { stageId: number; stageName: string } | null = null;

  // Task detail view/edit modal state
  detailTask: Task | null = null;
  detailStageName = '';
  detailTitle = '';
  detailDesc = '';
  detailDue = '';
  detailPriority = '';
  detailNotes = '';
  /** Done flag — stored in browser only (TaskCompletionStorageService). */
  detailCompleted = false;
  detailSubtasks: Subtask[] = [];
  subtasksLoading = false;
  subtaskError = '';
  newSubtaskTitle = '';
  subtaskSaving = false;
  subtaskDeletingId: number | null = null;
  deleteSubtaskPending: Subtask | null = null;
  detailComments: Comment[] = [];
  commentsLoading = false;
  newCommentContent = '';
  commentError = '';
  commentSaving = false;
  editingCommentId: number | string | null = null;
  editingCommentContent = '';
  deletingCommentId: number | string | null = null;
  deleteCommentPending: Comment | null = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private apiService: ApiService,
    private authService: AuthService,
    public themeService: ThemeService,
    private taskCompletionStorage: TaskCompletionStorageService,
    private cdr: ChangeDetectorRef,
    private notificationService: NotificationService
  ) {}

  ngOnInit() {
    const currentUser = this.authService.getCurrentUser();
    if (!currentUser) {
      this.router.navigate(['/login']);
      return;
    }
    this.userDisplayName = currentUser.name;
    this.userEmail = currentUser.email;
    this.knownUsers = this.authService.getKnownUsers();

    this.loadAllBoards();

    this.routeSub = this.route.params.subscribe(params => {
      const id = params['id'];
      if (id) {
        this.projectId = +id;
        if (!this.canAccessBoard(this.projectId, currentUser.email)) {
          this.router.navigate(['/boards']);
          return;
        }
        this.showBoardSwitcher = false;
        this.loadProject();
      } else {
        this.router.navigate(['/']);
      }
    });

    // Fallback timeout - show board even if API fails
    setTimeout(() => {
      if (this.loading) this.loading = false;
    }, 5000);
  }

  ngOnDestroy() {
    this.routeSub?.unsubscribe();
    if (this.pendingMemberRemoval) {
      clearTimeout(this.pendingMemberRemoval.timeoutId);
    }
  }

  private loadAllBoards() {
    this.apiService.getProjects().subscribe({
      next: (projects) => {
        const email = this.userEmail.trim().toLowerCase();
        this.allBoards = (projects || []).filter((project) => this.apiService.userHasProjectAccess(project.id, email));
      },
      error: () => { this.allBoards = []; }
    });
  }

  toggleBoardSwitcher() {
    this.showBoardSwitcher = !this.showBoardSwitcher;
  }

  switchBoard(id: number) {
    if (id === this.projectId) { this.showBoardSwitcher = false; return; }
    this.showBoardSwitcher = false;
    this.router.navigate(['/board', id]);
  }

  private canAccessBoard(projectId: number, email: string): boolean {
    return this.apiService.userHasProjectAccess(projectId, email);
  }

  loadProject() {
    this.loading = true;
    this.loadLabels();
    const currentUser = this.authService.getCurrentUser();
    console.log('Loading project from API...');
    this.apiService.getProject(this.projectId).subscribe({
      next: (project) => {
        console.log('Project loaded:', project);
        this.project = project;
        this.apiService.seedProjectOwner(this.projectId, project);
        this.loadStages();
      },
      error: (err) => {
        console.error('Failed to load project:', err);
        console.error('Error details:', err.message, err.status, err.url);
        // Keep user on board page even if backend is unavailable.
        this.project = {
          id: this.projectId,
          owner_id: currentUser?.id,
          name: `Board ${this.projectId}`,
          description: 'Demo board (backend unavailable)',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        };
        this.apiService.seedProjectOwner(this.projectId, this.project);
        this.loadStages();
      }
    });
  }

  loadStages() {
    console.log('Loading stages from API...');
    this.apiService.getStages(this.projectId).subscribe({
      next: (stages) => {
        console.log('Stages loaded:', stages);
        this.stages = (stages || []).map((s) => ({ ...s, tasks: s.tasks ?? [] }));
        this.loadCollapsedColumnState();
        // Load tasks for each stage
        if (this.stages.length > 0) {
          this.stages.forEach(stage => this.loadTasks(stage));
        }
        this.loading = false;
      },
      error: (err) => {
        console.error('Failed to load stages:', err);
        console.error('Error details:', err.message, err.status, err.url);
        // Add demo stages if API fails
        this.stages = [
          { id: 1, project_id: this.projectId, name: 'To Do', position: 0, created_at: new Date().toISOString(), updated_at: new Date().toISOString(), tasks: [] },
          { id: 2, project_id: this.projectId, name: 'In Progress', position: 1, created_at: new Date().toISOString(), updated_at: new Date().toISOString(), tasks: [] },
          { id: 3, project_id: this.projectId, name: 'Done', position: 2, created_at: new Date().toISOString(), updated_at: new Date().toISOString(), tasks: [] }
        ];
        this.loadCollapsedColumnState();
        this.loading = false;
      }
    });
  }

  loadTasks(stage: Stage) {
    this.apiService.getTasks(this.projectId, stage.id).subscribe({
      next: (tasks) => {
        stage.tasks = this.taskCompletionStorage.mergeTasks(this.projectId, tasks || []);
        this.sortStageTasks(stage);
        this.apiService.primeTaskComments((stage.tasks || []).map((task) => task.id));
        // Check for upcoming/overdue deadlines and push notifications
        const withDeadlines = (tasks || [])
          .map(t => ({ id: t.id, title: t.title, deadline: this.parseCardMeta(t.description || '').due }))
          .filter(t => !!t.deadline);
        if (withDeadlines.length) {
          this.notificationService.checkDeadlines(withDeadlines, this.projectId);
        }
      },
      error: (err) => {
        console.error('Failed to load tasks for stage:', stage.id, err);
      }
    });
  }

  private sortStageTasks(stage: Stage): void {
    if (!stage.tasks?.length) return;
    stage.tasks.sort((a, b) => a.position - b.position);
  }

  /** CDK drop list ids for connecting columns (drag tasks between lists). */
  get dropListIds(): string[] {
    return this.stages.map((s) => this.stageDropListId(s.id));
  }

  stageDropListId(stageId: number): string {
    return `stage-drop-${stageId}`;
  }

  /** Stable task array for CDK drop lists (must not allocate a new [] each change detection). */
  stageTasks(stage: Stage): Task[] {
    if (!stage.tasks) stage.tasks = [];
    return stage.tasks;
  }

  private parseDropListId(dropListElementId: string): number | null {
    const m = dropListElementId?.match(/^stage-drop-(\d+)$/);
    return m ? +m[1] : null;
  }

  onTaskDrop(event: CdkDragDrop<Task[]>): void {
    if (this.hasActiveFilters) return;

    const prevStageId = this.parseDropListId(event.previousContainer.id);
    const nextStageId = this.parseDropListId(event.container.id);
    if (prevStageId === null || nextStageId === null) return;

    const task = event.item.data as Task | undefined;
    if (!task) return;

    if (event.previousContainer === event.container) {
      moveItemInArray(event.container.data, event.previousIndex, event.currentIndex);
    } else {
      transferArrayItem(
        event.previousContainer.data,
        event.container.data,
        event.previousIndex,
        event.currentIndex
      );
    }

    const newPos = event.currentIndex;
    this.apiService.moveTask(task.id, { newStageId: nextStageId, newPos }).subscribe({
      next: () => {
        const prev = this.stages.find((s) => s.id === prevStageId);
        const next = this.stages.find((s) => s.id === nextStageId);
        if (prev) this.loadTasks(prev);
        if (next && next.id !== prev?.id) this.loadTasks(next);
      },
      error: () => {
        const prev = this.stages.find((s) => s.id === prevStageId);
        const next = this.stages.find((s) => s.id === nextStageId);
        if (prev) this.loadTasks(prev);
        if (next) this.loadTasks(next);
      }
    });
  }

  goBack() {
    this.router.navigate(['/boards']);
  }

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

  closeProfileMenu() {
    this.showProfileMenu = false;
  }

  logout() {
    this.closeProfileMenu();
    this.authService.logout();
    this.router.navigate(['/']);
  }

  openAccountSettings() {
    this.closeProfileMenu();
    this.showAccountModal = true;
    this.pendingEmail = this.userEmail;
    this.currentPassword = '';
    this.newPassword = '';
    this.confirmPassword = '';
    this.showCurrentPassword = false;
    this.showNewPassword = false;
    this.showConfirmPassword = false;
    this.accountError = '';
    this.accountSuccess = '';
  }

  closeAccountModal() {
    this.showAccountModal = false;
    this.accountError = '';
    this.accountSuccess = '';
  }

  saveEmailChange() {
    this.accountError = '';
    this.accountSuccess = '';

    const oldEmail = this.userEmail.trim().toLowerCase();
    const nextEmail = this.pendingEmail.trim().toLowerCase();

    if (!nextEmail) {
      this.accountError = 'Email is required.';
      return;
    }
    if (!this.isValidEmail(nextEmail)) {
      this.accountError = 'Please enter a valid email address.';
      return;
    }
    if (nextEmail === oldEmail) {
      this.accountError = 'New email must be different from current email.';
      return;
    }

    this.migrateBoardOwnersEmail(oldEmail, nextEmail);
    const updatedUser = this.authService.updateCurrentUser({ email: nextEmail });
    this.userEmail = updatedUser?.email || nextEmail;
    this.accountSuccess = 'Email updated in UI session.';
  }

  savePasswordChange() {
    this.accountError = '';
    this.accountSuccess = '';

    if (!this.currentPassword.trim()) {
      this.accountError = 'Current password is required.';
      return;
    }
    if (this.newPassword.length < 8) {
      this.accountError = 'New password must be at least 8 characters.';
      return;
    }
    if (this.newPassword !== this.confirmPassword) {
      this.accountError = 'New password and confirm password do not match.';
      return;
    }

    this.currentPassword = '';
    this.newPassword = '';
    this.confirmPassword = '';
    this.showCurrentPassword = false;
    this.showNewPassword = false;
    this.showConfirmPassword = false;
    this.accountSuccess = 'Password change UI is ready. Backend endpoint can be connected later.';
  }

  createStage() {
    const name = this.newStageName.trim();
    if (!name) {
      console.warn('Stage name is empty');
      return;
    }

    const position = this.stages.length;
    const request: CreateStageRequest = { name, position };
    console.log('Creating stage for project', this.projectId, 'with name:', name, 'position:', position);
    
    this.apiService.createStage(this.projectId, request).subscribe({
      next: (stage: Stage) => {
        console.log('Stage created successfully:', stage);
        this.stages.push(stage);
        this.newStageName = '';
      },
      error: (err) => {
        console.error('Failed to create stage:', err);
        console.error('Error status:', err.status);
        console.error('Error message:', err.message);
        // Demo fallback: create stage locally so board remains usable
        const localStage: Stage = {
          id: Date.now(),
          project_id: this.projectId,
          name,
          position,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          tasks: []
        };
        this.stages.push(localStage);
        this.newStageName = '';
      }
    });
  }

  createTask(stageId: number) {
    const title = this.newTaskTitles[stageId]?.trim();
    if (!title) return;

    const position = this.stages.find(s => s.id === stageId)?.tasks?.length || 0;
    const description = this.buildCardDescription(
      this.newTaskDescs[stageId] || '',
      this.newTaskDues[stageId] || '',
      this.newTaskPriorities[stageId] || '',
      this.newTaskNotes[stageId] || ''
    );
    const request: CreateTaskRequest = { title, description, position };
    this.apiService.createTask(this.projectId, stageId, request).subscribe({

      next: (task: Task) => {
        const stage = this.stages.find(s => s.id === stageId);
        if (stage) {
          if (!stage.tasks) stage.tasks = [];
          stage.tasks.push(task);
        }
        // Apply selected labels to the new task
        const selectedLabels = this.newTaskLabels[stageId] || [];
        if (selectedLabels.length) {
          this.taskLabelMap[task.id] = selectedLabels;
          localStorage.setItem(this.taskLabelsKey(task.id), JSON.stringify(selectedLabels));
        }
        this.newTaskTitles[stageId] = '';
        this.newTaskDescs[stageId] = '';
        this.newTaskDues[stageId] = '';
        this.newTaskPriorities[stageId] = '';
        this.newTaskNotes[stageId] = '';
        this.newTaskLabels[stageId] = [];
        this.showTaskDetails[stageId] = false;
      },
      error: (err) => {
        console.error('Failed to create task:', err);
        const stage = this.stages.find(s => s.id === stageId);
        if (stage) {
          if (!stage.tasks) stage.tasks = [];
          const localTask: Task = {
            id: Date.now(),
            stage_id: stageId,
            title,
            description,
            position,
            completed: false,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          };
          stage.tasks.push(localTask);
          const selectedLabels = this.newTaskLabels[stageId] || [];
          if (selectedLabels.length) {
            this.taskLabelMap[localTask.id] = selectedLabels;
            localStorage.setItem(this.taskLabelsKey(localTask.id), JSON.stringify(selectedLabels));
          }
        }
        this.newTaskTitles[stageId] = '';
        this.newTaskDescs[stageId] = '';
        this.newTaskDues[stageId] = '';
        this.newTaskPriorities[stageId] = '';
        this.newTaskNotes[stageId] = '';
        this.newTaskLabels[stageId] = [];
        this.showTaskDetails[stageId] = false;
      }
    });
  }

  toggleTaskDetails(stageId: number) {
    this.showTaskDetails[stageId] = !this.showTaskDetails[stageId];
  }

  private collapsedColumnsStorageKey(): string {
    return `taskify.board.collapsedColumns.v1.${this.projectId}`;
  }

  private loadCollapsedColumnState(): void {
    this.collapsedStages = {};
    try {
      const raw = sessionStorage.getItem(this.collapsedColumnsStorageKey());
      if (!raw) return;
      const ids = JSON.parse(raw) as number[];
      if (!Array.isArray(ids)) return;
      ids.forEach((id) => {
        if (typeof id === 'number' && !Number.isNaN(id)) {
          this.collapsedStages[id] = true;
        }
      });
    } catch {
      /* ignore */
    }
  }

  private saveCollapsedColumnState(): void {
    try {
      const ids = Object.keys(this.collapsedStages)
        .map(Number)
        .filter((id) => this.collapsedStages[id]);
      sessionStorage.setItem(this.collapsedColumnsStorageKey(), JSON.stringify(ids));
    } catch {
      /* ignore */
    }
  }

  isColumnCollapsed(stageId: number): boolean {
    return !!this.collapsedStages[stageId];
  }

  toggleColumnCollapsed(stageId: number, event?: Event): void {
    event?.stopPropagation();
    this.collapsedStages[stageId] = !this.collapsedStages[stageId];
    this.saveCollapsedColumnState();
  }

  requestDeleteStage(stageId: number, stageName: string): void {
    this.deleteTaskPending = null;
    this.deleteStagePending = { stageId, stageName };
  }

  cancelDeleteStage(): void {
    this.deleteStagePending = null;
  }

  confirmDeleteStage(): void {
    if (!this.deleteStagePending) return;
    const { stageId } = this.deleteStagePending;
    this.deleteStagePending = null;
    this.performDeleteStage(stageId);
  }

  private performDeleteStage(stageId: number): void {
    this.apiService.deleteStage(stageId).subscribe({
      next: () => {
        delete this.collapsedStages[stageId];
        this.saveCollapsedColumnState();
        this.stages = this.stages.filter((s) => s.id !== stageId);
      },
      error: (err) => {
        console.error('Failed to delete stage:', err);
        delete this.collapsedStages[stageId];
        this.saveCollapsedColumnState();
        this.stages = this.stages.filter((s) => s.id !== stageId);
      }
    });
  }

  requestDeleteTask(stageId: number, taskId: number, taskTitle: string, event: Event): void {
    event.stopPropagation();
    this.deleteStagePending = null;
    this.deleteTaskPending = { stageId, taskId, taskTitle };
  }

  cancelDeleteTask(): void {
    this.deleteTaskPending = null;
  }

  confirmDeleteTask(): void {
    if (!this.deleteTaskPending) return;
    const { stageId, taskId } = this.deleteTaskPending;
    this.deleteTaskPending = null;
    this.performDeleteTask(stageId, taskId);
  }

  private performDeleteTask(stageId: number, taskId: number): void {
    this.apiService.deleteTask(taskId).subscribe({
      next: () => {
        const stage = this.stages.find(s => s.id === stageId);
        if (stage && stage.tasks) {
          stage.tasks = stage.tasks.filter((t: Task) => t.id !== taskId);
        }
        this.taskCompletionStorage.setCompleted(this.projectId, taskId, false);
      },
      error: (err) => {
        console.error('Failed to delete task:', err);
        const stage = this.stages.find(s => s.id === stageId);
        if (stage && stage.tasks) {
          stage.tasks = stage.tasks.filter((t: Task) => t.id !== taskId);
        }
        this.taskCompletionStorage.setCompleted(this.projectId, taskId, false);
      }
    });
  }

  /** Resolved completion (task field + localStorage). */
  isTaskDone(task: Task): boolean {
    return !!(task.completed ?? this.taskCompletionStorage.getCompleted(this.projectId, task.id));
  }

  getStageTotalCount(stage: Stage): number {
    return stage.tasks?.length ?? 0;
  }

  getStageDoneCount(stage: Stage): number {
    const tasks = stage.tasks || [];
    return tasks.filter((t) => this.isTaskDone(t)).length;
  }

  trackByTaskId(_index: number, task: Task): number {
    return task.id;
  }

  /** Open task modal — edit flow (same as clicking card body). */
  editTask(task: Task, event: Event): void {
    event.stopPropagation();
    this.openTaskDetail(task);
  }

  openTaskDetail(task: Task, event?: Event) {
    if (event) event.stopPropagation();
    const stage = this.stages.find(s => (s.tasks || []).some(t => t.id === task.id));
    const parsed = this.parseCardMeta(task.description || '');

    this.detailTask = task;
    this.detailStageName = stage?.name || '';
    this.detailTitle = task.title || '';
    this.detailDesc = parsed.desc;
    this.detailDue = parsed.due;
    this.detailPriority = parsed.priority;
    this.detailNotes = parsed.notes;
    this.detailCompleted =
      task.completed ?? this.taskCompletionStorage.getCompleted(this.projectId, task.id);
    this.detailSubtasks = [];
    this.subtasksLoading = false;
    this.subtaskError = '';
    this.newSubtaskTitle = '';
    this.subtaskSaving = false;
    this.subtaskDeletingId = null;
    this.deleteSubtaskPending = null;
    this.newCommentContent = '';
    this.commentError = '';
    this.editingCommentId = null;
    this.editingCommentContent = '';
    this.deletingCommentId = null;
    this.deleteCommentPending = null;
    this.loadTaskSubtasks(task.id);
    this.loadTaskComments(task.id, true);
  }

  // ── Filter ───────────────────────────────────
  toggleFilterPanel() {
    this.showFilterPanel = !this.showFilterPanel;
  }

  clearFilters() {
    this.filterCompletion = '';
    this.filterPriority = '';
    this.filterDue = '';
    this.filterLabel = '';
  }

  get hasActiveFilters(): boolean {
    return !!(this.filterCompletion || this.filterPriority || this.filterDue || this.filterLabel);
  }

  getFilteredTasks(stage: Stage): Task[] {
    let tasks = stage.tasks || [];
    if (this.filterCompletion === 'active') {
      tasks = tasks.filter((t) => !this.isTaskDone(t));
    } else if (this.filterCompletion === 'done') {
      tasks = tasks.filter((t) => this.isTaskDone(t));
    }
    if (this.filterPriority) {
      tasks = tasks.filter(t => {
        const p = this.getEffectivePriority(t).toLowerCase();
        return p === this.filterPriority.toLowerCase();
      });
    }
    if (this.filterDue) {
      const today = new Date(); today.setHours(0, 0, 0, 0);
      tasks = tasks.filter(t => {
        const due = this.getTaskDue(t);
        if (!due) return this.filterDue === 'none';
        const dueDate = new Date(due); dueDate.setHours(0, 0, 0, 0);
        if (this.filterDue === 'none')    return false;
        if (this.filterDue === 'overdue') return dueDate < today;
        if (this.filterDue === 'today')   return dueDate.getTime() === today.getTime();
        if (this.filterDue === 'week') {
          const week = new Date(today); week.setDate(today.getDate() + 7);
          return dueDate >= today && dueDate <= week;
        }
        return true;
      });
    }
    if (this.filterLabel) {
      tasks = tasks.filter(t => this.getTaskLabelIds(t.id).includes(this.filterLabel));
    }
    return tasks;
  }

  // ── Labels ────────────────────────────────────

  private labelsKey(): string { return `taskify.labels.${this.projectId}`; }
  private taskLabelsKey(taskId: number): string { return `taskify.task-labels.${this.projectId}.${taskId}`; }

  loadLabels(): void {
    try {
      const raw = localStorage.getItem(this.labelsKey());
      this.projectLabels = raw ? JSON.parse(raw) : [];
    } catch { this.projectLabels = []; }
  }

  private saveLabels(): void {
    localStorage.setItem(this.labelsKey(), JSON.stringify(this.projectLabels));
  }

  createLabel(): void {
    const name = this.newLabelName.trim();
    if (!name) return;
    const label: Label = { id: `label-${Date.now()}`, name, color: this.newLabelColor };
    this.projectLabels = [...this.projectLabels, label];
    this.saveLabels();
    this.newLabelName = '';
  }

  deleteLabel(labelId: string): void {
    this.projectLabels = this.projectLabels.filter(l => l.id !== labelId);
    this.saveLabels();
    if (this.filterLabel === labelId) this.filterLabel = '';
    Object.keys(this.taskLabelMap).forEach(tid => {
      const id = +tid;
      this.taskLabelMap[id] = (this.taskLabelMap[id] || []).filter(i => i !== labelId);
      localStorage.setItem(this.taskLabelsKey(id), JSON.stringify(this.taskLabelMap[id]));
    });
  }

  getTaskLabelIds(taskId: number): string[] {
    if (!this.taskLabelMap[taskId]) {
      try {
        const raw = localStorage.getItem(this.taskLabelsKey(taskId));
        this.taskLabelMap[taskId] = raw ? JSON.parse(raw) : [];
      } catch { this.taskLabelMap[taskId] = []; }
    }
    return this.taskLabelMap[taskId];
  }

  getTaskLabels(taskId: number): Label[] {
    return this.projectLabels.filter(l => this.getTaskLabelIds(taskId).includes(l.id));
  }

  isLabelOnTask(taskId: number, labelId: string): boolean {
    return this.getTaskLabelIds(taskId).includes(labelId);
  }

  toggleLabelOnTask(taskId: number, labelId: string): void {
    const current = this.getTaskLabelIds(taskId);
    this.taskLabelMap[taskId] = current.includes(labelId)
      ? current.filter(id => id !== labelId)
      : [...current, labelId];
    localStorage.setItem(this.taskLabelsKey(taskId), JSON.stringify(this.taskLabelMap[taskId]));
  }

  toggleNewTaskLabel(stageId: number, labelId: string): void {
    const current = this.newTaskLabels[stageId] || [];
    this.newTaskLabels[stageId] = current.includes(labelId)
      ? current.filter(id => id !== labelId)
      : [...current, labelId];
  }

  isLabelSelectedForNew(stageId: number, labelId: string): boolean {
    return (this.newTaskLabels[stageId] || []).includes(labelId);
  }

  // ── Share ─────────────────────────────────────
  openShareModal() {
    this.showShareModal = true;
    this.shareLinkCopied = false;
  }

  openProjectSettings(tab: 'general' | 'members' | 'activity' = 'general'): void {
    this.showProjectSettingsModal = true;
    this.settingsActiveTab = tab;
    this.memberError = '';
    this.memberSuccess = '';
    this.addMemberQuery = '';
    this.loadProjectMembers();
    if (tab === 'activity' && this.isProjectOwner()) {
      this.loadActivityLogs(true);
    }
  }

  closeProjectSettings(): void {
    if (this.addingMember || this.removingMemberId) return;
    this.showProjectSettingsModal = false;
    this.memberError = '';
    this.memberSuccess = '';
    this.addMemberQuery = '';
    this.activityError = '';
  }

  setSettingsTab(tab: 'general' | 'members' | 'activity'): void {
    this.settingsActiveTab = tab;
    if ((tab === 'members' || tab === 'activity') && !this.projectMembers.length) {
      this.loadProjectMembers();
    }
    if (tab === 'activity' && this.isProjectOwner() && !this.activityHasLoaded) {
      this.loadActivityLogs(true);
    }
  }

  loadProjectMembers(): void {
    this.membersLoading = true;
    this.apiService.getProjectMembers(this.projectId).subscribe({
      next: (members) => {
        this.projectMembers = this.sortMembers(members || []);
        this.syncProjectMemberCount();
        this.membersLoading = false;
      },
      error: () => {
        this.projectMembers = this.sortMembers(this.apiService.getCachedProjectMembers(this.projectId));
        this.syncProjectMemberCount();
        this.membersLoading = false;
      }
    });
  }

  get memberSearchResults(): AuthUser[] {
    const query = this.addMemberQuery.trim().toLowerCase();
    if (!query) return [];

    return this.knownUsers
      .filter((user) => {
        const name = user.name?.trim().toLowerCase() || '';
        const email = user.email?.trim().toLowerCase() || '';
        const notAlreadyAdded = !this.projectMembers.some((member) => {
          const memberEmail = (member.user_email || '').trim().toLowerCase();
          return memberEmail === email || (!!user.id && member.user_id === user.id);
        });

        return notAlreadyAdded && (name.includes(query) || email.includes(query));
      })
      .slice(0, 5);
  }

  get memberCount(): number {
    return this.projectMembers.length || this.apiService.getProjectMemberCount(this.projectId);
  }

  get canViewActivityTab(): boolean {
    return this.isProjectOwner();
  }

  get activityMemberOptions(): ProjectMember[] {
    return this.projectMembers.filter((member) => !!member.user_id);
  }

  get hasMoreActivity(): boolean {
    return this.activityLogs.length < this.activityTotal;
  }

  isProjectOwner(): boolean {
    const currentEmail = this.userEmail.trim().toLowerCase();
    const currentId = this.authService.getCurrentUser()?.id;

    if (currentId && this.project?.owner_id && currentId === this.project.owner_id) {
      return true;
    }

    return this.projectMembers.some((member) => {
      const memberEmail = (member.user_email || '').trim().toLowerCase();
      return member.role === 'owner' && (memberEmail === currentEmail || (!!currentId && member.user_id === currentId));
    });
  }

  canRemoveMember(member: ProjectMember): boolean {
    return this.isProjectOwner() && member.role !== 'owner';
  }

  applyActivityFilters(): void {
    if (!this.canViewActivityTab) return;
    this.loadActivityLogs(true);
  }

  resetActivityFilters(): void {
    this.activityFilterUserId = '';
    this.activityFilterFrom = '';
    this.activityFilterTo = '';
    this.loadActivityLogs(true);
  }

  loadMoreActivity(): void {
    if (this.activityLoading || this.activityLoadingMore || !this.hasMoreActivity) return;
    this.loadActivityLogs(false);
  }

  getTaskCommentCount(taskId: number): number {
    return this.apiService.getTaskCommentCount(taskId);
  }

  addMember(user?: AuthUser): void {
    const query = (user?.email || this.addMemberQuery).trim();
    const normalizedQuery = query.toLowerCase();
    this.memberError = '';
    this.memberSuccess = '';

    if (!query) {
      this.memberError = 'Enter a name or email to add a member.';
      return;
    }

    const matchedUser = user || this.knownUsers.find((knownUser) => {
      const name = knownUser.name?.trim().toLowerCase() || '';
      const email = knownUser.email?.trim().toLowerCase() || '';
      return email === normalizedQuery || name === normalizedQuery;
    });

    const request: AddProjectMemberRequest = {
      user_id: matchedUser?.id,
      email: matchedUser?.email || (this.isValidEmail(normalizedQuery) ? normalizedQuery : undefined)
    };

    if (!request.user_id && !request.email) {
      this.memberError = 'Select a matching user or enter a valid email address.';
      return;
    }

    const duplicate = this.projectMembers.some((member) => {
      const memberEmail = (member.user_email || '').trim().toLowerCase();
      return memberEmail === (request.email || '').trim().toLowerCase() || (!!request.user_id && member.user_id === request.user_id);
    });

    if (duplicate) {
      this.memberError = 'This member is already on the project.';
      return;
    }

    this.addingMember = true;

    this.apiService.addProjectMember(this.projectId, request).subscribe({
      next: (member) => {
        this.projectMembers = this.sortMembers([...this.projectMembers, member]);
        this.syncProjectMemberCount();
        this.addMemberQuery = '';
        this.memberSuccess = `${member.user_name || member.user_email || 'Member'} added to the project.`;
        this.addingMember = false;
        this.loadAllBoards();
      },
      error: (error) => {
        this.memberError = error?.error?.error || 'Could not add this member.';
        this.addingMember = false;
      }
    });
  }

  removeMember(member: ProjectMember): void {
    if (!this.canRemoveMember(member)) return;
    if (!window.confirm(`Remove ${member.user_name || member.user_email || 'this member'} from the project?`)) {
      return;
    }

    if (this.pendingMemberRemoval) {
      this.finalizePendingMemberRemoval();
    }

    this.memberError = '';
    this.memberSuccess = `${member.user_name || member.user_email || 'Member'} will be removed. Undo if needed.`;
    this.removingMemberId = member.user_id;

    this.projectMembers = this.sortMembers(
      this.projectMembers.filter((projectMember) => projectMember.user_id !== member.user_id)
    );
    this.apiService.setCachedProjectMembers(this.projectId, this.projectMembers);
    this.syncProjectMemberCount();

    const timeoutId = setTimeout(() => this.finalizePendingMemberRemoval(), 5000);
    this.pendingMemberRemoval = { member, timeoutId };
  }

  undoRemoveMember(): void {
    if (!this.pendingMemberRemoval) return;

    clearTimeout(this.pendingMemberRemoval.timeoutId);
    this.projectMembers = this.sortMembers([...this.projectMembers, this.pendingMemberRemoval.member]);
    this.apiService.setCachedProjectMembers(this.projectId, this.projectMembers);
    this.syncProjectMemberCount();
    this.memberSuccess = `${this.pendingMemberRemoval.member.user_name || this.pendingMemberRemoval.member.user_email || 'Member'} restored.`;
    this.removingMemberId = null;
    this.pendingMemberRemoval = null;
    this.loadAllBoards();
  }

  private finalizePendingMemberRemoval(): void {
    if (!this.pendingMemberRemoval) return;

    const pendingRemoval = this.pendingMemberRemoval;
    this.pendingMemberRemoval = null;
    this.apiService.removeProjectMember(this.projectId, pendingRemoval.member.user_id).subscribe({
      next: () => {
        this.memberSuccess = `${pendingRemoval.member.user_name || pendingRemoval.member.user_email || 'Member'} removed from the project.`;
        this.removingMemberId = null;
        this.loadAllBoards();
      },
      error: (error) => {
        this.projectMembers = this.sortMembers([...this.projectMembers, pendingRemoval.member]);
        this.apiService.setCachedProjectMembers(this.projectId, this.projectMembers);
        this.syncProjectMemberCount();
        this.memberError = error?.error?.error || 'Could not remove this member.';
        this.removingMemberId = null;
      }
    });
  }

  getMemberInitial(member: ProjectMember): string {
    return (member.user_name || member.user_email || 'M').charAt(0).toUpperCase();
  }

  private syncProjectMemberCount(): void {
    if (this.project) {
      this.project = { ...this.project, member_count: this.projectMembers.length || 1 };
    }
  }

  private sortMembers(members: ProjectMember[]): ProjectMember[] {
    return [...members].sort((a, b) => {
      if (a.role !== b.role) return a.role === 'owner' ? -1 : 1;
      return (a.user_name || a.user_email || '').localeCompare(b.user_name || b.user_email || '');
    });
  }

  closeShareModal() {
    this.showShareModal = false;
    this.shareLinkCopied = false;
  }

  get boardUrl(): string {
    return window.location.href;
  }

  copyShareLink() {
    navigator.clipboard.writeText(this.boardUrl).then(() => {
      this.shareLinkCopied = true;
      setTimeout(() => { this.shareLinkCopied = false; }, 2500);
    });
  }

  closeTaskDetail() {
    this.detailTask = null;
    this.detailSubtasks = [];
    this.subtasksLoading = false;
    this.subtaskError = '';
    this.newSubtaskTitle = '';
    this.subtaskSaving = false;
    this.subtaskDeletingId = null;
    this.deleteSubtaskPending = null;
    this.detailComments = [];
    this.newCommentContent = '';
    this.commentError = '';
    this.editingCommentId = null;
    this.editingCommentContent = '';
    this.deletingCommentId = null;
    this.deleteCommentPending = null;
  }

  saveTaskDetail() {
    if (!this.detailTask) return;

    const task = this.detailTask;
    const updatedTitle = this.detailTitle.trim() || task.title;
    const updatedDescription = this.buildCardDescription(
      this.detailDesc,
      this.detailDue,
      this.detailPriority,
      this.detailNotes
    );

    this.taskCompletionStorage.setCompleted(this.projectId, task.id, this.detailCompleted);
    task.completed = this.detailCompleted;
    this.updateTaskInStages(task.id, { completed: this.detailCompleted });
    this.cdr.detectChanges();

    this.apiService.updateTask(task.id, {
      title: updatedTitle,
      description: updatedDescription,
      position: task.position
    }).subscribe({
      next: (updated) => {
        task.title = updated.title;
        task.description = updated.description;
        this.updateTaskInStages(task.id, {
          title: updated.title,
          description: updated.description,
          completed: this.detailCompleted
        });
        this.cdr.detectChanges();
        this.closeTaskDetail();
      },
      error: (err) => {
        console.error('Failed to update task:', err);
        // Demo fallback: update locally when backend is unavailable
        task.title = updatedTitle;
        task.description = updatedDescription;
        this.updateTaskInStages(task.id, {
          title: updatedTitle,
          description: updatedDescription,
          completed: this.detailCompleted
        });
        this.cdr.detectChanges();
        this.closeTaskDetail();
      }
    });
  }

  loadTaskSubtasks(taskId: number): void {
    this.subtasksLoading = true;
    this.subtaskError = '';
    this.apiService.getSubtasks(taskId).subscribe({
      next: (subtasks) => {
        this.detailSubtasks = this.sortSubtasks(subtasks || []);
        this.syncTaskSubtaskCounts(taskId, this.detailSubtasks);
        this.subtasksLoading = false;
      },
      error: () => {
        this.detailSubtasks = [];
        this.subtaskError = 'Could not load checklist items.';
        this.subtasksLoading = false;
      }
    });
  }

  addSubtask(): void {
    const taskId = this.detailTask?.id;
    const title = this.newSubtaskTitle.trim();
    if (!taskId || !title) return;

    this.subtaskSaving = true;
    this.subtaskError = '';
    this.apiService.createSubtask(taskId, { title }).subscribe({
      next: (subtask) => {
        this.detailSubtasks = this.sortSubtasks([...this.detailSubtasks, subtask]);
        this.syncTaskSubtaskCounts(taskId, this.detailSubtasks);
        this.newSubtaskTitle = '';
        this.subtaskSaving = false;
      },
      error: () => {
        this.subtaskError = 'Could not add checklist item.';
        this.subtaskSaving = false;
      }
    });
  }

  toggleSubtask(subtask: Subtask, event: Event): void {
    const input = event.target as HTMLInputElement;
    const nextCompleted = input.checked;
    const previousSubtasks = this.detailSubtasks;
    const updatedSubtasks = this.detailSubtasks.map((item) =>
      item.id === subtask.id ? { ...item, is_completed: nextCompleted } : item
    );

    this.detailSubtasks = updatedSubtasks;
    this.syncTaskSubtaskCounts(subtask.task_id, updatedSubtasks);
    this.subtaskError = '';

    this.apiService.updateSubtask(subtask.id, { is_completed: nextCompleted }).subscribe({
      next: (updated) => {
        this.detailSubtasks = this.sortSubtasks(
          updatedSubtasks.map((item) => (item.id === updated.id ? updated : item))
        );
        this.syncTaskSubtaskCounts(subtask.task_id, this.detailSubtasks);
      },
      error: () => {
        this.detailSubtasks = previousSubtasks;
        this.syncTaskSubtaskCounts(subtask.task_id, previousSubtasks);
        this.subtaskError = 'Could not update checklist item.';
      }
    });
  }

  requestDeleteSubtask(subtask: Subtask): void {
    this.deleteSubtaskPending = subtask;
  }

  cancelDeleteSubtask(): void {
    this.deleteSubtaskPending = null;
  }

  confirmDeleteSubtask(): void {
    if (!this.deleteSubtaskPending) return;
    const subtask = this.deleteSubtaskPending;
    this.deleteSubtaskPending = null;

    const previousSubtasks = this.detailSubtasks;
    const updatedSubtasks = this.detailSubtasks
      .filter((item) => item.id !== subtask.id)
      .map((item, index) => ({ ...item, position: index }));

    this.subtaskDeletingId = subtask.id;
    this.subtaskError = '';
    this.detailSubtasks = updatedSubtasks;
    this.syncTaskSubtaskCounts(subtask.task_id, updatedSubtasks);

    this.apiService.deleteSubtask(subtask.id).subscribe({
      next: () => {
        this.subtaskDeletingId = null;
      },
      error: () => {
        this.detailSubtasks = previousSubtasks;
        this.syncTaskSubtaskCounts(subtask.task_id, previousSubtasks);
        this.subtaskDeletingId = null;
        this.subtaskError = 'Could not delete checklist item.';
      }
    });
  }

  loadTaskComments(taskId: number, scrollToBottom = false): void {
    this.commentsLoading = true;
    this.apiService.getComments(taskId).subscribe({
      next: (comments) => {
        this.detailComments = comments || [];
        this.commentsLoading = false;
        if (scrollToBottom) {
          this.scrollCommentsToLatest();
        }
      },
      error: () => {
        this.detailComments = this.apiService.getCachedTaskComments(taskId);
        this.commentsLoading = false;
        if (scrollToBottom) {
          this.scrollCommentsToLatest();
        }
      }
    });
  }

  canManageComment(comment: Comment): boolean {
    const currentUser = this.authService.getCurrentUser();
    if (!currentUser) return false;

    return comment.user_id === currentUser.id || comment.user_id === currentUser.email;
  }

  postComment(): void {
    const taskId = this.detailTask?.id;
    const content = this.newCommentContent.trim();
    if (!taskId) return;

    this.commentError = '';
    if (!content) {
      this.commentError = 'Comment cannot be empty.';
      return;
    }

    this.commentSaving = true;
    const request: CreateCommentRequest = { content };

    this.apiService.createComment(taskId, request).subscribe({
      next: (comment) => {
        const exists = this.detailComments.some((existingComment) => String(existingComment.id) === String(comment.id));
        this.detailComments = exists
          ? this.detailComments.map((existingComment) =>
              String(existingComment.id) === String(comment.id) ? comment : existingComment
            )
          : [...this.detailComments, comment];
        this.detailComments.sort(
          (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
        );
        this.newCommentContent = '';
        this.commentSaving = false;
        this.scrollCommentsToLatest();
      },
      error: (error) => {
        this.commentError = error?.error || 'Could not post comment.';
        this.commentSaving = false;
      }
    });
  }

  startCommentEdit(comment: Comment): void {
    this.commentError = '';
    this.editingCommentId = comment.id;
    this.editingCommentContent = comment.content;
  }

  cancelCommentEdit(): void {
    this.editingCommentId = null;
    this.editingCommentContent = '';
  }

  saveCommentEdit(comment: Comment): void {
    const content = this.editingCommentContent.trim();
    if (!content || !this.detailTask) {
      this.commentError = 'Comment cannot be empty.';
      return;
    }

    this.commentError = '';
    this.commentSaving = true;

    this.apiService.updateComment(comment.id, this.detailTask.id, { content }).subscribe({
      next: (updatedComment) => {
        this.detailComments = this.detailComments.map((existingComment) =>
          String(existingComment.id) === String(comment.id) ? updatedComment : existingComment
        );
        this.commentSaving = false;
        this.cancelCommentEdit();
      },
      error: (error) => {
        this.commentError = error?.error || 'Could not update comment.';
        this.commentSaving = false;
      }
    });
  }

  confirmDeleteComment(comment: Comment): void {
    this.deleteCommentPending = comment;
  }

  cancelDeleteComment(): void {
    this.deleteCommentPending = null;
  }

  executeDeleteComment(): void {
    if (!this.detailTask || !this.deleteCommentPending) return;

    const comment = this.deleteCommentPending;
    this.deleteCommentPending = null;
    this.commentError = '';
    this.deletingCommentId = comment.id;
    this.apiService.deleteComment(comment.id, this.detailTask.id).subscribe({
      next: () => {
        this.detailComments = this.detailComments.filter(
          (existingComment) => String(existingComment.id) !== String(comment.id)
        );
        this.deletingCommentId = null;
        if (String(this.editingCommentId) === String(comment.id)) {
          this.cancelCommentEdit();
        }
      },
      error: (error) => {
        this.commentError = error?.error || 'Could not delete comment.';
        this.deletingCommentId = null;
      }
    });
  }

  formatCommentTimestamp(iso: string): string {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short'
    });
  }

  formatActivityTimestamp(iso: string): string {
    const diffMs = Date.now() - new Date(iso).getTime();
    const minuteMs = 60 * 1000;
    const hourMs = 60 * minuteMs;
    const dayMs = 24 * hourMs;

    if (diffMs < minuteMs) return 'Just now';
    if (diffMs < hourMs) return `${Math.max(1, Math.floor(diffMs / minuteMs))}m ago`;
    if (diffMs < dayMs) return `${Math.floor(diffMs / hourMs)}h ago`;
    if (diffMs < 7 * dayMs) return `${Math.floor(diffMs / dayMs)}d ago`;

    return new Date(iso).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  getActivityIcon(activity: ActivityLog): string {
    if (activity.action.includes('moved')) return '↔';
    if (activity.action.includes('comment')) return '◦';
    if (activity.action.includes('member')) return '+';
    if (activity.action.includes('label')) return '#';
    if (activity.action.includes('deleted')) return '−';
    return '•';
  }

  getCommentInitial(comment: Comment): string {
    return (comment.author_name || 'U').charAt(0).toUpperCase();
  }

  trackBySubtaskId(_index: number, subtask: Subtask): number {
    return subtask.id;
  }

  getTaskSubtaskCount(task: Task): number {
    return task.subtask_count ?? 0;
  }

  getTaskCompletedSubtaskCount(task: Task): number {
    return task.completed_count ?? 0;
  }

  getTaskSubtaskProgress(task: Task): number {
    const total = this.getTaskSubtaskCount(task);
    if (!total) return 0;
    return Math.round((this.getTaskCompletedSubtaskCount(task) / total) * 100);
  }

  getDetailSubtaskCompletedCount(): number {
    return this.detailSubtasks.filter((subtask) => subtask.is_completed).length;
  }

  private sortSubtasks(subtasks: Subtask[]): Subtask[] {
    return [...subtasks].sort((a, b) => {
      const positionDiff = a.position - b.position;
      return positionDiff !== 0 ? positionDiff : a.id - b.id;
    });
  }

  private syncTaskSubtaskCounts(taskId: number, subtasks: Subtask[]): void {
    const total = subtasks.length;
    const completed = subtasks.filter((subtask) => subtask.is_completed).length;

    this.stages = this.stages.map((stage) => ({
      ...stage,
      tasks: (stage.tasks || []).map((task) =>
        task.id === taskId
          ? { ...task, subtask_count: total, completed_count: completed }
          : task
      )
    }));

    if (this.detailTask?.id === taskId) {
      this.detailTask = {
        ...this.detailTask,
        subtask_count: total,
        completed_count: completed
      };
    }
  }

  private updateTaskInStages(taskId: number, changes: Partial<Task>): void {
    this.stages = this.stages.map((stage) => ({
      ...stage,
      tasks: (stage.tasks || []).map((existingTask) =>
        existingTask.id === taskId ? { ...existingTask, ...changes } : existingTask
      )
    }));

    if (this.detailTask?.id === taskId) {
      this.detailTask = { ...this.detailTask, ...changes };
    }
  }

  private loadActivityLogs(reset: boolean): void {
    if (!this.canViewActivityTab) return;

    const nextPage = reset ? 1 : this.activityPage + 1;
    this.activityLoading = reset;
    this.activityLoadingMore = !reset;
    this.activityError = '';

    this.apiService.getProjectActivity(this.projectId, {
      page: nextPage,
      limit: this.activityLimit,
      user_id: this.activityFilterUserId || undefined,
      from: this.toActivityBoundary(this.activityFilterFrom, 'start'),
      to: this.toActivityBoundary(this.activityFilterTo, 'end')
    }).subscribe({
      next: (response) => {
        const incoming = response.data || [];
        this.activityLogs = reset ? incoming : [...this.activityLogs, ...incoming];
        this.activityPage = response.page || nextPage;
        this.activityTotal = response.total || this.activityLogs.length;
        this.activityHasLoaded = true;
        this.activityLoading = false;
        this.activityLoadingMore = false;
      },
      error: () => {
        if (reset) {
          this.activityLogs = [];
          this.activityTotal = 0;
        }
        this.activityError = 'Could not load activity history.';
        this.activityHasLoaded = true;
        this.activityLoading = false;
        this.activityLoadingMore = false;
      }
    });
  }

  private toActivityBoundary(value: string, boundary: 'start' | 'end'): string | undefined {
    const trimmed = value.trim();
    if (!trimmed) return undefined;
    return boundary === 'start'
      ? `${trimmed}T00:00:00.000Z`
      : `${trimmed}T23:59:59.999Z`;
  }

  private scrollCommentsToLatest(): void {
    setTimeout(() => {
      const commentContainer = document.querySelector('.commentsList') as HTMLElement | null;
      const latestComment = commentContainer?.querySelector('.commentCard:last-child') as HTMLElement | null;
      if (commentContainer) {
        commentContainer.scrollIntoView({ behavior: 'smooth', block: 'end' });
        commentContainer.scrollTop = commentContainer.scrollHeight;
      }
      if (latestComment) {
        latestComment.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      }
    });
  }

  /** Checkbox on task card — localStorage only. */
  toggleTaskCompleted(_stageId: number, task: Task, event: Event): void {
    event.stopPropagation();
    const input = event.target as HTMLInputElement;
    const next = input.checked;
    this.taskCompletionStorage.setCompleted(this.projectId, task.id, next);
    task.completed = next;
    this.cdr.detectChanges();
  }

  private parseCardMeta(description: string): { desc: string; due: string; priority: string; notes: string } {
    const idx = description.indexOf(this.META_SEP);
    let desc = description;
    let due = '';
    let priority = '';
    let notes = '';
    if (idx >= 0) {
      desc = description.slice(0, idx).trim();
      const meta = description.slice(idx + this.META_SEP.length);
      meta.split('\n').forEach(line => {
        if (line.startsWith('due:')) due = line.slice(4).trim();
        else if (line.startsWith('priority:')) priority = line.slice(9).trim();
        else if (line.startsWith('notes:')) notes = line.slice(6).trim();
      });
    }
    return { desc, due, priority, notes };
  }

  private buildCardDescription(desc: string, due: string, priority: string, notes: string): string {
    const baseDesc = desc.trim();
    const parts: string[] = [];
    if (due.trim()) parts.push('due:' + due.trim());
    if (priority.trim()) parts.push('priority:' + priority.trim());
    if (notes.trim()) parts.push('notes:' + notes.trim());
    if (parts.length === 0) return baseDesc;
    return baseDesc + this.META_SEP + parts.join('\n');
  }

  getDisplayDescription(task: Task): string {
    return this.parseCardMeta(task.description || '').desc;
  }

  getTaskPriority(task: Task): string {
    return this.parseCardMeta(task.description || '').priority;
  }

  getTaskDue(task: Task): string {
    return this.parseCardMeta(task.description || '').due;
  }

  /** Display due as `Due Date: MM/DD/YYYY` (matches priority pill row). */
  formatTaskDueDisplay(task: Task): string {
    const raw = this.getTaskDue(task);
    if (!raw?.trim()) return '';
    const s = raw.trim();
    const ymd = s.match(/^(\d{4})-(\d{2})-(\d{2})/);
    if (ymd) {
      const [, yyyy, mm, dd] = ymd;
      return `Due Date: ${mm}/${dd}/${yyyy}`;
    }
    const d = new Date(s);
    if (Number.isNaN(d.getTime())) return s;
    const mm = String(d.getMonth() + 1).padStart(2, '0');
    const dd = String(d.getDate()).padStart(2, '0');
    const yyyy = d.getFullYear();
    return `Due Date: ${mm}/${dd}/${yyyy}`;
  }

  /** Returns the visually escalated priority based on deadline proximity.
   *  The stored priority is never changed — this is display-only. */
  getEffectivePriority(task: Task): string {
    const set = this.getTaskPriority(task).toLowerCase();
    const raw = this.getTaskDue(task);

    if (!raw?.trim()) return set; // no deadline → show as set

    const today = new Date(); today.setHours(0, 0, 0, 0);
    const due = new Date(raw); due.setHours(0, 0, 0, 0);
    const daysLeft = Math.ceil((due.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));

    const rank: Record<string, number> = { '': 0, 'low': 1, 'lowest': 1, 'medium': 2, 'mid': 2, 'high': 3, 'highest': 3, 'urgent': 4, 'critical': 4 };
    const label: Record<number, string> = { 1: 'Low', 2: 'Medium', 3: 'High', 4: 'Urgent' };

    let minRank = rank[set] ?? 0;
    if (daysLeft < 0)       minRank = Math.max(minRank, 4); // overdue → Urgent
    else if (daysLeft === 0) minRank = Math.max(minRank, 3); // today   → High
    else if (daysLeft <= 2)  minRank = Math.max(minRank, 2); // ≤2 days → Medium

    return label[minRank] ?? set;
  }

  getPriorityClass(task: Task): string {
    const priority = this.getEffectivePriority(task).toLowerCase();
    if (priority === 'urgent' || priority === 'critical' || priority === 'high' || priority === 'highest') return 'priority-high';
    if (priority === 'medium' || priority === 'mid') return 'priority-mid';
    if (priority === 'low' || priority === 'lowest') return 'priority-low';
    return 'priority-none';
  }

  getCreatePriorityClass(stageId: number): string {
    const priority = (this.newTaskPriorities[stageId] || '').toLowerCase();
    if (priority === 'urgent' || priority === 'critical' || priority === 'high') return 'priority-high';
    if (priority === 'medium') return 'priority-mid';
    if (priority === 'low') return 'priority-low';
    return 'priority-none';
  }

  getDetailPriorityClass(): string {
    const priority = (this.detailPriority || '').toLowerCase();
    if (priority === 'urgent' || priority === 'critical' || priority === 'high') return 'priority-high';
    if (priority === 'medium') return 'priority-mid';
    if (priority === 'low') return 'priority-low';
    return 'priority-none';
  }

  getDueDateClass(task: Task): string {
    const raw = this.getTaskDue(task);
    if (!raw?.trim()) return '';
    const today = new Date(); today.setHours(0, 0, 0, 0);
    const due = new Date(raw); due.setHours(0, 0, 0, 0);
    if (due < today) return 'due-overdue';
    if (due.getTime() === today.getTime()) return 'due-today';
    return '';
  }

  private migrateBoardOwnersEmail(oldEmail: string, newEmail: string) {
    if (!oldEmail || oldEmail === newEmail) return;
    try {
      const raw = localStorage.getItem(this.boardOwnersKey);
      const owners = raw ? JSON.parse(raw) as Record<string, string> : {};
      Object.keys(owners).forEach((projectId) => {
        if ((owners[projectId] || '').trim().toLowerCase() === oldEmail) {
          owners[projectId] = newEmail;
        }
      });
      localStorage.setItem(this.boardOwnersKey, JSON.stringify(owners));
    } catch {
      // noop: keep UI responsive even if localStorage parse fails
    }
  }

  private isValidEmail(email: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  }
}
