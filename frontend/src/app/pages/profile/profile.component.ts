import { CommonModule } from '@angular/common';
import { Component } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { AuthService, AuthUser } from '../../services/auth.service';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.scss']
})
export class ProfileComponent {
  currentUser: AuthUser | null = null;
  profileName = '';
  profileEmail = '';
  profileMessage = '';

  constructor(private authService: AuthService) {
    this.currentUser = this.authService.getCurrentUser();
    this.resetProfileForm();
  }

  get displayName(): string {
    return this.currentUser?.name?.trim() || 'Taskify User';
  }

  get email(): string {
    return this.emailValue || 'No email on file';
  }

  get role(): string {
    const role = this.currentUser?.role?.trim();
    return role ? this.toTitleCase(role) : 'Member';
  }

  get userInitial(): string {
    return this.displayName.charAt(0).toUpperCase();
  }

  get hasProfileChanges(): boolean {
    return this.profileName.trim() !== this.displayName || this.profileEmail.trim() !== this.emailValue;
  }

  saveProfile(): void {
    const name = this.profileName.trim();
    const email = this.profileEmail.trim();

    if (!name || !email) {
      this.profileMessage = 'Name and email are required.';
      return;
    }

    const updatedUser = this.authService.updateCurrentUser({ name, email });
    if (!updatedUser) {
      this.profileMessage = 'Unable to update profile right now.';
      return;
    }

    this.currentUser = updatedUser;
    this.resetProfileForm();
    this.profileMessage = 'Profile updated.';
  }

  resetProfileForm(): void {
    this.profileName = this.displayName;
    this.profileEmail = this.emailValue;
  }

  private get emailValue(): string {
    return this.currentUser?.email?.trim() || '';
  }

  private toTitleCase(value: string): string {
    return value
      .split(/[\s_-]+/)
      .filter(Boolean)
      .map(part => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
      .join(' ');
  }
}
