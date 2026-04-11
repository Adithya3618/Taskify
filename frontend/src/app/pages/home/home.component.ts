import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../services/api.service';
import { Project } from '../../models/project.model';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { NotificationBellComponent } from '../../components/notification-bell/notification-bell.component';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, NotificationBellComponent],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
/** Home page: boards list, navbar with profile, login/signup links, footer. */
export class HomeComponent {
  projects: Project[] = [];
  loading = true;
  currentYear = new Date().getFullYear();
  apiError = false;
  apiErrorTitle = '';
  apiErrorBody = '';
  showBackendRunHint = false;
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
    private authService: AuthService,
    public themeService: ThemeService
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
        this.apiErrorTitle = '';
        this.apiErrorBody = '';
        this.showBackendRunHint = false;
        this.useDemoData = false;
      },
      error: (err) => {
        console.error('Failed to load projects:', err);
        this.apiError = true;
        this.useDemoData = true;
        this.setApiErrorState(err);

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
    return projects
      .filter((project) => this.apiService.userHasProjectAccess(project.id, email))
      .map((project) => ({
        ...project,
        member_count: this.apiService.getProjectMemberCount(project.id)
      }));
  }

  private setBoardOwner(projectId: number) {
    this.boardOwners[String(projectId)] = this.userEmail.trim().toLowerCase();
    this.saveBoardOwners();
    this.apiService.seedProjectOwner(projectId);
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
        this.apiErrorTitle = '';
        this.apiErrorBody = '';
        this.showBackendRunHint = false;
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

  private migrateBoardOwnersEmail(oldEmail: string, newEmail: string) {
    if (!oldEmail || oldEmail === newEmail) return;
    Object.keys(this.boardOwners).forEach((projectId) => {
      if ((this.boardOwners[projectId] || '').trim().toLowerCase() === oldEmail) {
        this.boardOwners[projectId] = newEmail;
      }
    });
    this.saveBoardOwners();
  }

  private isValidEmail(email: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  }

  private setApiErrorState(err: any) {
    const status = Number(err?.status ?? 0);
    const serverMessage = String(err?.error?.error || err?.message || '').trim();

    if (status === 0) {
      this.apiErrorTitle = 'Could not connect to backend';
      this.apiErrorBody = 'The app could not reach the API server on localhost:8080.';
      this.showBackendRunHint = true;
      return;
    }

    if (status === 401 || status === 403) {
      this.apiErrorTitle = 'Session expired or unauthorized';
      this.apiErrorBody = 'Please sign in again to continue.';
      this.showBackendRunHint = false;
      return;
    }

    if (status >= 500) {
      this.apiErrorTitle = 'Backend returned an error';
      this.apiErrorBody = serverMessage || `Request failed with status ${status}.`;
      this.showBackendRunHint = false;
      return;
    }

    this.apiErrorTitle = 'Could not load boards';
    this.apiErrorBody = serverMessage || `Request failed with status ${status}.`;
    this.showBackendRunHint = false;
  }
}
