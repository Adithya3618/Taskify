import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';
import { Stage, CreateStageRequest } from '../../models/stage.model';
import { Task, CreateTaskRequest } from '../../models/task.model';

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

  // New item inputs
  newStageName = '';
  newTaskTitles: { [key: number]: string } = {};

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
      if (this.loading) {
        console.log('Board load timeout - showing board anyway');
        this.loading = false;
      }
    }, 5000);
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
        this.loading = false;
        this.router.navigate(['/']);
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
    this.apiService.getTasks(stage.id).subscribe({
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
    const request: CreateTaskRequest = { title, description: '', position };
    this.apiService.createTask(stageId, request).subscribe({
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