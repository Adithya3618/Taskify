import { Component } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './signup.component.html',
  styleUrls: ['./signup.component.scss'],
})
/** Signup page: name, email, password, confirm; link to login; redirects to home on submit. */
export class SignupComponent {
  name = '';
  email = '';
  password = '';
  confirmPassword = '';
  loading = false;
  error = '';
  nameError = '';
  emailError = '';
  passwordError = '';
  confirmPasswordError = '';
  showPassword = false;
  showConfirmPassword = false;
  googleLoading = false;
  googleError = '';

  constructor(
    private router: Router,
    private authService: AuthService,
    public themeService: ThemeService
  ) {}

  signInWithGoogle() {
    this.googleLoading = true;
    this.googleError = '';
    this.authService.startGoogleLogin().subscribe({
      next: () => { window.location.href = '/api/auth/google/login'; },
      error: (err: Error) => {
        this.googleLoading = false;
        this.googleError = err.message || 'Google sign-in is not available right now.';
      }
    });
  }

  onSubmit() {
    this.error = '';
    this.nameError = '';
    this.emailError = '';
    this.passwordError = '';
    this.confirmPasswordError = '';
    if (!this.name.trim()) {
      this.nameError = 'Please enter your name.';
      this.error = 'Please fix the highlighted fields.';
      return;
    }
    if (!this.email.trim()) {
      this.emailError = 'Please enter your email.';
      this.error = 'Please fix the highlighted fields.';
      return;
    }
    if (!this.isValidEmail(this.email.trim())) {
      this.emailError = 'Please enter a valid email address.';
      this.error = 'Please fix the highlighted fields.';
      return;
    }
    if (!this.password) {
      this.passwordError = 'Please enter a password.';
      this.error = 'Please fix the highlighted fields.';
      return;
    }
    if (this.password.length < 8) {
      this.passwordError = 'Password must be at least 8 characters.';
      this.error = 'Please fix the highlighted fields.';
      return;
    }
    if (this.password !== this.confirmPassword) {
      this.confirmPasswordError = 'Passwords do not match.';
      this.error = 'Please fix the highlighted fields.';
      return;
    }
    this.loading = true;

    this.authService.register({
      name: this.name,
      email: this.email,
      password: this.password
    }).subscribe({
      next: () => {
        this.loading = false;
        this.router.navigate(['/boards']);
      },
      error: (err) => {
        this.loading = false;
        this.error = err.error?.error || 'Signup failed.';
      }
    });
  }

  private isValidEmail(value: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
  }
}
