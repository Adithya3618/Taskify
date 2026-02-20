import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';
import { Stage, CreateStageRequest } from '../../models/stage.model';
import { Task, CreateTaskRequest } from '../../models/task.model';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-board',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss']
})
export class BoardComponent implements OnInit {
  projectId: number = 0;
  project: Project | null = null;
  stages: Stage[] = [];
  loading = true;
  private readonly boardOwnersKey = 'taskify.board.owners';
  private readonly META_SEP = '\n---\n';

  // New item inputs
  newStageName = '';
  newTaskTitles: { [key: number]: string } = {};
  newTaskDescs: { [key: number]: string } = {};
  newTaskDues: { [key: number]: string } = {};
  newTaskPriorities: { [key: number]: string } = {};
  newTaskNotes: { [key: number]: string } = {};
  showTaskDetails: { [key: number]: boolean } = {};

  // Task detail view/edit modal state
  detailTask: Task | null = null;
  detailStageName = '';
  detailTitle = '';
  detailDesc = '';
  detailDue = '';
  detailPriority = '';
  detailNotes = '';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private apiService: ApiService,
    private authService: AuthService
  ) {}

  ngOnInit() {
    console.log('BoardComponent initialized');
    const currentUser = this.authService.getCurrentUser();
    if (!currentUser) {
      this.router.navigate(['/login']);
      return;
    }

    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.projectId = +id;
      if (!this.canAccessBoard(this.projectId, currentUser.email)) {
        this.router.navigate(['/boards']);
        return;
      }
      console.log('Loading project ID:', this.projectId);
      this.loadProject();
    } else {
      console.log('No project ID found, redirecting to home');
      this.router.navigate(['/']);
    }

    // Fallback timeout - show board even if API fails
    setTimeout(() => {
      if (this.loading) {
        console.log('Board load timeout - showing board anyway');
        this.loading = false;
      }
    }, 5000);
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
        this.stages = stages || [];
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
        this.loading = false;
      }
    });
  }

  loadTasks(stage: Stage) {
    this.apiService.getTasks(this.projectId, stage.id).subscribe({
      next: (tasks) => {
        stage.tasks = tasks;
      },
      error: (err) => {
        console.error('Failed to load tasks for stage:', stage.id, err);
      }
    });
  }

  goBack() {
    this.router.navigate(['/']);
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
        alert('Failed to create stage. Check console for details.');
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
      }
    });
  }

  toggleTaskDetails(stageId: number) {
    this.showTaskDetails[stageId] = !this.showTaskDetails[stageId];
  }

  deleteStage(stageId: number) {
    if (confirm('Are you sure you want to delete this stage and all its tasks?')) {
      this.apiService.deleteStage(stageId).subscribe({
        next: () => {
          this.stages = this.stages.filter(s => s.id !== stageId);
        },
        error: (err) => {
          console.error('Failed to delete stage:', err);
        }
      });
    }
  }

  deleteTask(stageId: number, taskId: number, event?: Event) {
    if (event) event.stopPropagation();
    if (confirm('Are you sure you want to delete this task?')) {
      this.apiService.deleteTask(taskId).subscribe({
        next: () => {
          const stage = this.stages.find(s => s.id === stageId);
          if (stage && stage.tasks) {
            stage.tasks = stage.tasks.filter((t: Task) => t.id !== taskId);
          }
        },
        error: (err) => {
          console.error('Failed to delete task:', err);
        }
      });
    }
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
      }
    });
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

  getPriorityClass(task: Task): string {
    const priority = this.getTaskPriority(task).toLowerCase();
    if (priority === 'critical' || priority === 'high' || priority === 'highest') return 'priority-high';
    if (priority === 'medium' || priority === 'mid') return 'priority-mid';
    if (priority === 'low' || priority === 'lowest') return 'priority-low';
    return 'priority-none';
  }

  getCreatePriorityClass(stageId: number): string {
    const priority = (this.newTaskPriorities[stageId] || '').toLowerCase();
    if (priority === 'critical' || priority === 'high') return 'priority-high';
    if (priority === 'medium') return 'priority-mid';
    if (priority === 'low') return 'priority-low';
    return 'priority-none';
  }
}
