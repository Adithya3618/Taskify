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
      if (this.isLoading) {
        console.log('Board load timeout - showing board anyway');
        this.isLoading = false;
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
}
