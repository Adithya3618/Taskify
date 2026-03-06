import { Component, OnDestroy } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss'],
})
/** Login page: email/password form, link to signup, redirects to home on submit. */
export class LoginComponent implements OnDestroy {
  email = '';
  password = '';
  loading = false;
  error = '';
  showForgotModal = false;
  forgotEmail = '';
  forgotCode = '';
  hasSentCode = false;
  codeCooldown = 0;
  private codeTimer: ReturnType<typeof setInterval> | null = null;

  constructor(
    private router: Router,
    private authService: AuthService
  ) {}

  openForgotPassword() {
    this.showForgotModal = true;
    this.forgotEmail = this.email.trim();
    this.forgotCode = '';
    this.resetCodeTimer();
  }

  closeForgotPassword() {
    this.showForgotModal = false;
    this.stopCodeTimer();
  }

  submitForgotPassword() {
    // UI-only flow: no validation or backend call.
    this.closeForgotPassword();
  }

  sendOrResendCode() {
    if (this.codeCooldown > 0) return;
    this.hasSentCode = true;
    this.codeCooldown = 30;
    this.stopCodeTimer();
    this.codeTimer = setInterval(() => {
      if (this.codeCooldown > 0) {
        this.codeCooldown -= 1;
      }
      if (this.codeCooldown <= 0) {
        this.stopCodeTimer();
      }
    }, 1000);
  }

  get codeButtonLabel(): string {
    if (!this.hasSentCode) return 'Send code';
    if (this.codeCooldown > 0) return `Resend code (${this.codeCooldown}s)`;
    return 'Resend code';
  }

  onSubmit() {
    this.error = '';
    if (!this.email.trim()) {
      this.error = 'Please enter your email.';
      return;
    }
    if (!this.password) {
      this.error = 'Please enter your password.';
      return;
    }
    this.loading = true;

    this.authService.login(this.email, this.password).subscribe({
      next: () => {
        this.loading = false;
        this.router.navigate(['/boards']);
      },
      error: (err) => {
        this.loading = false;
        this.error = err.error?.error || 'Login failed.';
      }
    });
  }

  ngOnDestroy() {
    this.stopCodeTimer();
  }

  private resetCodeTimer() {
    this.hasSentCode = false;
    this.codeCooldown = 0;
    this.stopCodeTimer();
  }

  private stopCodeTimer() {
    if (this.codeTimer) {
      clearInterval(this.codeTimer);
      this.codeTimer = null;
    }
  }
}
