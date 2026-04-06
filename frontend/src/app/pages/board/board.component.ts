import { ChangeDetectorRef, Component, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink, RouterLinkActive } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { Subscription } from 'rxjs';
import { CdkDragDrop, DragDropModule, moveItemInArray, transferArrayItem } from '@angular/cdk/drag-drop';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';
import { Stage, CreateStageRequest } from '../../models/stage.model';
import { Task, CreateTaskRequest } from '../../models/task.model';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { TaskCompletionStorageService } from '../../services/task-completion-storage.service';

@Component({
  selector: 'app-board',
  standalone: true,
  imports: [CommonModule, FormsModule, DragDropModule, RouterLink, RouterLinkActive],
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

  // Share state
  showShareModal = false;
  shareLinkCopied = false;

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

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private apiService: ApiService,
    private authService: AuthService,
    public themeService: ThemeService,
    private taskCompletionStorage: TaskCompletionStorageService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit() {
    const currentUser = this.authService.getCurrentUser();
    if (!currentUser) {
      this.router.navigate(['/login']);
      return;
    }
    this.userDisplayName = currentUser.name;
    this.userEmail = currentUser.email;

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
  }

  private loadAllBoards() {
    this.apiService.getProjects().subscribe({
      next: (projects) => {
        const email = this.userEmail.trim().toLowerCase();
        const raw = localStorage.getItem(this.boardOwnersKey);
        const owners: Record<string, string> = raw ? JSON.parse(raw) : {};
        this.allBoards = (projects || []).filter(p => owners[String(p.id)] === email);
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
    try {
      const raw = localStorage.getItem(this.boardOwnersKey);
      const owners = raw ? JSON.parse(raw) as Record<string, string> : {};
      return owners[String(projectId)] === email.trim().toLowerCase();
    } catch {
      return false;
    }
  }

  loadProject() {
    this.loading = true;
    console.log('Loading project from API...');
    this.apiService.getProject(this.projectId).subscribe({
      next: (project) => {
        console.log('Project loaded:', project);
        this.project = project;
        this.loadStages();
      },
      error: (err) => {
        console.error('Failed to load project:', err);
        console.error('Error details:', err.message, err.status, err.url);
        // Keep user on board page even if backend is unavailable.
        this.project = {
          id: this.projectId,
          name: `Board ${this.projectId}`,
          description: 'Demo board (backend unavailable)',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        };
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
        this.newTaskTitles[stageId] = '';
        this.newTaskDescs[stageId] = '';
        this.newTaskDues[stageId] = '';
        this.newTaskPriorities[stageId] = '';
        this.newTaskNotes[stageId] = '';
        this.showTaskDetails[stageId] = false;
      },
      error: (err) => {
        console.error('Failed to create task:', err);
        // Demo fallback: keep board usable even when backend task creation fails
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
        }
        this.newTaskTitles[stageId] = '';
        this.newTaskDescs[stageId] = '';
        this.newTaskDues[stageId] = '';
        this.newTaskPriorities[stageId] = '';
        this.newTaskNotes[stageId] = '';
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
  }

  // ── Filter ───────────────────────────────────
  toggleFilterPanel() {
    this.showFilterPanel = !this.showFilterPanel;
  }

  clearFilters() {
    this.filterCompletion = '';
    this.filterPriority = '';
    this.filterDue = '';
  }

  get hasActiveFilters(): boolean {
    return !!(this.filterCompletion || this.filterPriority || this.filterDue);
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
    return tasks;
  }

  // ── Share ─────────────────────────────────────
  openShareModal() {
    this.showShareModal = true;
    this.shareLinkCopied = false;
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

    this.apiService.updateTask(task.id, {
      title: updatedTitle,
      description: updatedDescription,
      position: task.position
    }).subscribe({
      next: (updated) => {
        task.title = updated.title;
        task.description = updated.description;
        this.closeTaskDetail();
      },
      error: (err) => {
        console.error('Failed to update task:', err);
        // Demo fallback: update locally when backend is unavailable
        task.title = updatedTitle;
        task.description = updatedDescription;
        this.closeTaskDetail();
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
