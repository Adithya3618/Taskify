import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { RouterModule } from '@angular/router';
import { AuthService, AuthUser } from '../../services/auth.service';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.scss']
})
export class ProfileComponent {
  currentUser: AuthUser | null = null;

  constructor(private authService: AuthService) {
    this.currentUser = this.authService.getCurrentUser();
  }

  get displayName(): string {
    return this.currentUser?.name?.trim() || 'Taskify User';
  }

  get email(): string {
    return this.currentUser?.email?.trim() || 'No email on file';
  }

  get role(): string {
    const role = this.currentUser?.role?.trim();
    return role ? this.toTitleCase(role) : 'Member';
  }

  get userInitial(): string {
    return this.displayName.charAt(0).toUpperCase();
  }

  private toTitleCase(value: string): string {
    return value
      .split(/[\s_-]+/)
      .filter(Boolean)
      .map(part => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
      .join(' ');
  }
}
