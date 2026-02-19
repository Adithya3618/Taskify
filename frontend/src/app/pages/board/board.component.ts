import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';
import { Stage, CreateStageRequest } from '../../models/stage.model';
import { Task, CreateTaskRequest } from '../../models/task.model';
import { ErrorBannerComponent } from '../../components/error-banner/error-banner.component';

@Component({
  selector: 'app-board',
  standalone: true,
  imports: [CommonModule, FormsModule, ErrorBannerComponent],
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss']
})
export class BoardComponent implements OnInit {
  projectId: number = 0;
  project: Project | null = null;
  stages: Stage[] = [];
  isLoading = false;
  errorMsg = '';

  // New item inputs
  newStageName = '';
  newTaskTitles: { [key: number]: string } = {};

  // Card detail panel (click card → expand panel)
  detailTask: Task | null = null;
  detailStageName = '';
  detailTitle = '';
  detailDesc = '';
  detailDue = '';
  detailPriority = '';
  detailNotes = '';

  private readonly META_SEP = '\n---\n';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private apiService: ApiService
  ) {}

  ngOnInit() {
    console.log('BoardComponent initialized');
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.projectId = +id;
      console.log('Loading project ID:', this.projectId);
      this.loadProject();
    } else {
      console.log('No project ID found, redirecting to home');
      this.router.navigate(['/']);
    }

    // Fallback timeout - show board even if API fails
    setTimeout(() => {
      if (this.isLoading) {
        console.log('Board load timeout - showing board anyway');
        this.isLoading = false;
      }
    }, 5000);
  }

  loadProject() {
    this.isLoading = true;
    this.errorMsg = '';
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
        this.stages = [];
        this.isLoading = false;
        this.errorMsg = err?.error?.message || err?.message || 'Failed to load board.';
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
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Failed to load stages:', err);
        console.error('Error details:', err.message, err.status, err.url);
        this.stages = [];
        this.isLoading = false;
        this.errorMsg = err?.error?.message || err?.message || 'Failed to load board.';
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

  openAddListDialog() {
    const name = prompt('Enter stage title...');
    if (!name || !name.trim()) return;
    this.newStageName = name.trim();
    this.createStage();
  }

  createTask(stageId: number) {
    const title = this.newTaskTitles[stageId]?.trim();
    if (!title) return;

    const position = this.stages.find(s => s.id === stageId)?.tasks?.length || 0;
    const request: CreateTaskRequest = { title, description: '', position };
    this.apiService.createTask(this.projectId, stageId, request).subscribe({

      next: (task: Task) => {
        const stage = this.stages.find(s => s.id === stageId);
        if (stage) {
          if (!stage.tasks) stage.tasks = [];
          stage.tasks.push(task);
        }
        this.newTaskTitles[stageId] = '';
      },
      error: (err) => {
        console.error('Failed to create task:', err);
      }
    });
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

  deleteTask(stageId: number, taskId: number) {
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

  openCardDetail(task: Task, event?: Event) {
    if (event) (event as Event).stopPropagation();
    const stage = this.stages.find(s => (s.tasks || []).some((t: Task) => t.id === task.id));
    this.detailTask = task;
    this.detailStageName = stage?.name || '';
    this.detailTitle = task.title || '';
    const parsed = this.parseCardMeta(task.description || '');
    this.detailDesc = parsed.desc;
    this.detailDue = parsed.due;
    this.detailPriority = parsed.priority;
    this.detailNotes = parsed.notes;
  }

  closeCardDetail() {
    this.detailTask = null;
  }

  saveCardDetail() {
    if (!this.detailTask) return;
    const task = this.detailTask;
    const title = this.detailTitle.trim() || task.title;
    const description = this.buildCardDescription(
      this.detailDesc,
      this.detailDue,
      this.detailPriority,
      this.detailNotes
    );
    this.apiService.updateTask(task.id, { title, description, position: task.position }).subscribe({
      next: (updated) => {
        task.title = updated.title;
        task.description = updated.description;
        this.closeCardDetail();
      },
      error: (err) => console.error('Failed to update task:', err)
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
    const parts: string[] = [];
    if (due) parts.push('due:' + due);
    if (priority) parts.push('priority:' + priority);
    if (notes) parts.push('notes:' + notes);
    if (parts.length === 0) return desc.trim();
    return (desc.trim() + this.META_SEP + parts.join('\n'));
  }

  getDisplayDescription(task: Task): string {
    return this.parseCardMeta(task.description || '').desc || 'Click to add description';
  }
}
