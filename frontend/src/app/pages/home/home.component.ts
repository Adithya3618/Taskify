import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
/** Home page: boards list, navbar with profile, login/signup links, footer. */
export class HomeComponent {
  projects: Project[] = [];
  loading = true;
  currentYear = new Date().getFullYear();
  apiError = false;
  useDemoData = false;
  private readonly boardOwnersKey = 'taskify.board.owners';
  private boardOwners: Record<string, string> = {};

  // --- 3-dot board menu state ---
  openMenuProjectId: number | null = null;

  // --- rename modal state ---
  renameModalOpen = false;
  renameProjectId: number | null = null;
  renameName = '';
  renameDesc = '';

  // Profile menu
  showProfileMenu = false;
  userDisplayName = '';
  userEmail = '';
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
    private apiService: ApiService,
    private authService: AuthService
  ) {
    const currentUser = this.authService.getCurrentUser();
    if (!currentUser) {
      this.router.navigate(['/login']);
      return;
    }

    this.userDisplayName = currentUser.name;
    this.userEmail = currentUser.email;
    this.loadBoardOwners();
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
        this.projects = this.filterProjectsForCurrentUser(projects || []);
        this.loading = false;
        this.apiError = false;
        this.useDemoData = false;
      },
      error: (err) => {
        console.error('Failed to load projects:', err);
        this.apiError = true;
        this.useDemoData = true;

        // In private mode, keep boards scoped to the authenticated user.
        this.projects = [];
        this.loading = false;
      }
    });
  }

  private loadBoardOwners() {
    try {
      const raw = localStorage.getItem(this.boardOwnersKey);
      this.boardOwners = raw ? JSON.parse(raw) : {};
    } catch {
      this.boardOwners = {};
    }
  }

  private saveBoardOwners() {
    localStorage.setItem(this.boardOwnersKey, JSON.stringify(this.boardOwners));
  }

  private filterProjectsForCurrentUser(projects: Project[]): Project[] {
    const email = this.userEmail.trim().toLowerCase();
    if (!email) return [];
    return projects.filter((project) => this.boardOwners[String(project.id)] === email);
  }

  private setBoardOwner(projectId: number) {
    this.boardOwners[String(projectId)] = this.userEmail.trim().toLowerCase();
    this.saveBoardOwners();
  }

  private removeBoardOwner(projectId: number) {
    delete this.boardOwners[String(projectId)];
    this.saveBoardOwners();
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

  logout() {
    this.closeProfileMenu();
    this.authService.logout();
    this.router.navigate(['/']);
  }

  toggleBoardMenu(projectId: number, event: MouseEvent) {
    event.stopPropagation();
    this.openMenuProjectId = this.openMenuProjectId === projectId ? null : projectId;
  }

  closeBoardMenu() {
    this.openMenuProjectId = null;
  }

  // Close menu when clicking anywhere else on the page
  onPageClick() {
    this.closeBoardMenu();
  }

  /** Menu action: Open board */
  openBoardFromMenu(projectId: number, event: MouseEvent) {
    event.stopPropagation();
    this.closeBoardMenu();
    this.openBoard(projectId);
  }

  /** Menu action: Rename board (open modal) */
  openRenameModal(project: Project, event: MouseEvent) {
    event.stopPropagation();
    this.closeBoardMenu();

    this.renameModalOpen = true;
    this.renameProjectId = project.id;
    this.renameName = project.name || '';
    this.renameDesc = project.description || '';
  }

  closeRenameModal() {
    this.renameModalOpen = false;
    this.renameProjectId = null;
    this.renameName = '';
    this.renameDesc = '';
  }

  /** Save rename (calls backend if available, otherwise updates UI optimistically) */
  saveRename() {
    if (!this.renameProjectId) return;

    const id = this.renameProjectId;
    const payload = { name: this.renameName.trim(), description: this.renameDesc.trim() };

    if (!payload.name) return;

    // Optimistic UI update (works without backend too)
    const idx = this.projects.findIndex(p => p.id === id);
    if (idx !== -1) {
      this.projects[idx] = { ...this.projects[idx], ...payload };
    }

    // If backend is up, also persist:
    this.apiService.updateProject(id, payload).subscribe({
      next: () => this.loadProjects(),
      error: () => {
        // If backend fails, keep the optimistic update for demo
        console.warn('Rename failed (backend not ready). Kept UI change for demo.');
      }
    });

    this.closeRenameModal();
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
        this.setBoardOwner(project.id);
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
        // Demo fallback: still allow creating a private local board for this user.
        const demoProject: Project = {
          id: Date.now(),
          name,
          description,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        };
        this.setBoardOwner(demoProject.id);
        this.projects = [demoProject, ...this.projects];
        this.showCreateModal = false;
        this.useDemoData = true;
        this.creating = false;
      }
    });
  }

  // -------------------------
  // Delete board (confirm)
  // -------------------------
  openDeleteModal(project: Project, event?: MouseEvent) {
    event?.stopPropagation();
    this.closeBoardMenu();
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
        this.removeBoardOwner(this.projectToDelete!.id);
        this.projects = this.projects.filter(p => p.id !== this.projectToDelete!.id);
        this.deleting = false;
        this.closeDeleteModal();
      },
      error: (err) => {
        console.error('Failed to delete project:', err);
        // For demo/local-only boards, remove from UI and ownership map even if API fails.
        this.removeBoardOwner(this.projectToDelete!.id);
        this.projects = this.projects.filter(p => p.id !== this.projectToDelete!.id);
        this.deleting = false;
        this.closeDeleteModal();
      }
    });
  }
}
