import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';
import { ErrorBannerComponent } from '../../components/error-banner/error-banner.component';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, ErrorBannerComponent],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
/** Home page: boards list, navbar with profile, login/signup links, footer. */
export class HomeComponent {
  projects: Project[] = [];
  isLoading = true;
  currentYear = new Date().getFullYear();
  errorMsg = '';
  apiError = false;
  useDemoData = false;

  // Profile menu (placeholder user until auth is wired)
  showProfileMenu = false;
  userDisplayName = 'Guest User';
  userEmail = 'guest@taskify.com';
  get userInitial(): string {
    return (this.userDisplayName || 'U').charAt(0).toUpperCase();
  }

  // Trello-style board colors (each board gets a unique color)
  boardColors = [
    '#0079bf', '#70b500', '#ff9f1a', '#eb5a46', '#c377e0',
    '#00c2e0', '#51e898', '#ff78cb', '#344563', '#b3b9c4',
    '#026aa7', '#4bce97', '#f5cd47', '#f87168', '#9f8fef',
    '#4dadee', '#7bc86c', '#fad29c', '#ef7560', '#cd8de5'
  ];

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
      if (this.isLoading) this.isLoading = false;
    }, 5000);
  }

  loadProjects() {
    this.isLoading = true;
    this.errorMsg = '';
    this.apiService.getProjects().subscribe({
      next: (projects) => {
        this.projects = projects || [];
        this.isLoading = false;
        this.apiError = false;
        this.useDemoData = false;
      },
      error: (err) => {
        console.error('Failed to load projects:', err);
        this.apiError = true;
        this.useDemoData = true;
        this.errorMsg = 'Failed to load boards';
        // Demo fallback (your branch): show sample boards when API is down
        this.projects = [
          { id: 1, name: 'Demo Project', description: 'This is demo data', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
          { id: 2, name: 'Sample Board', description: 'Click to open board', created_at: new Date().toISOString(), updated_at: new Date().toISOString() }
        ];
        this.isLoading = false;
      }
    });
  }

  getProjectColor(index: number): string {
    return this.boardColors[index % this.boardColors.length];
  }

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

  closeProfileMenu() {
    this.showProfileMenu = false;
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
        this.errorMsg = '';
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
