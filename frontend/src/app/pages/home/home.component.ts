import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']

})
export class HomeComponent {
  projects: Project[] = [];
  loading = true;
  apiError = false;
  useDemoData = false;

  // Avatar colors
  colors = ['#cdb4db', '#bde0fe', '#ffc8dd', '#ffafcc', '#a2d2ff', '#bde0fe'];

  // ----- Create modal state -----
  showCreateModal = false;
  newBoardName = '';
  newBoardDesc = '';
  creating = false;

  // ----- Delete modal state -----
  showDeleteModal = false;
  deleting = false;
  projectToDelete: Project | null = null;

  constructor(
    private router: Router,
    private apiService: ApiService
  ) {
    this.loadProjects();

    // Stop infinite loading if API hangs
    setTimeout(() => {
      if (this.loading) this.loading = false;
    }, 5000);
  }

  loadProjects() {
    this.loading = true;
    this.apiService.getProjects().subscribe({
      next: (projects) => {
        this.projects = projects || [];
        this.loading = false;
        this.apiError = false;
        this.useDemoData = false;
      },
      error: (err) => {
        console.error('Failed to load projects:', err);
        this.apiError = true;
        this.useDemoData = true;

        // Demo fallback
        this.projects = [
          { id: 1, name: 'Demo Project', description: 'This is demo data', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
          { id: 2, name: 'Sample Board', description: 'Click to open board', created_at: new Date().toISOString(), updated_at: new Date().toISOString() }
        ];
        this.loading = false;
      }
    });
  }

  getProjectColor(index: number): string {
    return this.colors[index % this.colors.length];
  }

  openBoard(projectId: number) {
    this.router.navigate(['/board', projectId]);
  }

  // -------------------------
  // Create board (modal)
  // -------------------------
  openCreateModal() {
    this.newBoardName = '';
    this.newBoardDesc = '';
    this.showCreateModal = true;
  }

  closeCreateModal() {
    if (this.creating) return;
    this.showCreateModal = false;
  }

  createNewBoard() {
    const name = this.newBoardName.trim();
    const description = this.newBoardDesc.trim();

    if (!name) return;

    this.creating = true;

    this.apiService.createProject({ name, description }).subscribe({
      next: (project) => {
        // add to top
        this.projects = [project, ...this.projects];
        this.creating = false;
        this.showCreateModal = false;

        // if demo mode was active, switch off now
        this.apiError = false;
        this.useDemoData = false;
      },
      error: (err) => {
        console.error('Failed to create project:', err);
        this.creating = false;
      }
    });
  }

  // -------------------------
  // Delete board (confirm)
  // -------------------------
  openDeleteModal(project: Project) {
    this.projectToDelete = project;
    this.showDeleteModal = true;
  }

  closeDeleteModal() {
    if (this.deleting) return;
    this.showDeleteModal = false;
    this.projectToDelete = null;
  }

  confirmDeleteBoard() {
    if (!this.projectToDelete) return;

    this.deleting = true;

    this.apiService.deleteProject(this.projectToDelete.id).subscribe({
      next: () => {
        this.projects = this.projects.filter(p => p.id !== this.projectToDelete!.id);
        this.deleting = false;
        this.closeDeleteModal();
      },
      error: (err) => {
        console.error('Failed to delete project:', err);
        this.deleting = false;
      }
    });
  }
}
