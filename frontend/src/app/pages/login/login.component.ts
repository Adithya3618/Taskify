import { Component, OnDestroy } from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';

type ForgotStep = 'request' | 'verify' | 'reset' | 'success';

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
  emailError = '';
  passwordError = '';
  showPassword = false;
  googleLoading = false;
  googleError = '';
  showForgotNewPassword = false;
  showForgotConfirmPassword = false;

  showForgotModal = false;
  forgotEmail = '';
  forgotCode = '';
  forgotNewPassword = '';
  forgotConfirmPassword = '';
  forgotStep: ForgotStep = 'request';
  forgotLoading = false;
  forgotError = '';
  forgotInfo = '';

  hasSentCode = false;
  codeCooldown = 0;
  private forgotResetToken = '';
  private codeTimer: ReturnType<typeof setInterval> | null = null;

  constructor(
    private router: Router,
    private authService: AuthService,
    public themeService: ThemeService
  ) {}

  openForgotPassword() {
    this.showForgotModal = true;
    this.forgotEmail = this.email.trim();
    this.resetForgotFlow();
  }

  closeForgotPassword() {
    this.showForgotModal = false;
    this.stopCodeTimer();
    this.resetForgotFlow();
  }

  sendOrResendCode() {
    if (this.forgotLoading || this.codeCooldown > 0) return;

    const email = this.forgotEmail.trim();
    if (!email) {
      this.forgotError = 'Please enter your email.';
      return;
    }
    if (!this.isValidEmail(email)) {
      this.forgotError = 'Please enter a valid email address.';
      return;
    }

    this.forgotLoading = true;
    this.forgotError = '';
    this.forgotInfo = '';

    this.authService.forgotPassword(email).subscribe({
      next: (response) => {
        this.forgotLoading = false;
        this.forgotStep = 'verify';
        this.hasSentCode = true;
        this.forgotInfo = response.message || 'If an account exists, a verification code has been sent.';
        this.startCodeCooldown();
      },
      error: (err) => {
        this.forgotLoading = false;
        this.forgotError = err.error?.error || 'Failed to send code. Please try again.';
      }
    });
  }

  submitForgotPassword() {
    if (this.forgotLoading) return;

    if (this.forgotStep === 'request') {
      this.sendOrResendCode();
      return;
    }

    if (this.forgotStep === 'verify') {
      const email = this.forgotEmail.trim();
      const code = this.forgotCode.trim();

      if (!email) {
        this.forgotError = 'Please enter your email.';
        return;
      }
      if (!code) {
        this.forgotError = 'Please enter the verification code.';
        return;
      }

      this.forgotLoading = true;
      this.forgotError = '';
      this.forgotInfo = '';

      this.authService.verifyResetOtp(email, code).subscribe({
        next: (response) => {
          this.forgotLoading = false;
          this.forgotResetToken = response.reset_token;
          this.forgotStep = 'reset';
          this.forgotCode = '';
        },
        error: (err) => {
          this.forgotLoading = false;
          this.forgotError = err.error?.error || 'Invalid or expired code. Please try again.';
        }
      });
      return;
    }

    if (this.forgotStep === 'reset') {
      if (!this.forgotNewPassword) {
        this.forgotError = 'Please enter a new password.';
        return;
      }
      if (this.forgotNewPassword.length < 8) {
        this.forgotError = 'Password must be at least 8 characters.';
        return;
      }
      if (this.forgotNewPassword !== this.forgotConfirmPassword) {
        this.forgotError = 'Passwords do not match.';
        return;
      }
      if (!this.forgotResetToken) {
        this.forgotError = 'Your reset session has expired. Please request a new code.';
        this.backToRequestStep();
        return;
      }

      this.forgotLoading = true;
      this.forgotError = '';
      this.forgotInfo = '';

      this.authService.resetPassword(this.forgotResetToken, this.forgotNewPassword).subscribe({
        next: (response) => {
          this.forgotLoading = false;
          this.forgotStep = 'success';
          this.forgotInfo = response.message || 'Password has been reset successfully.';
          this.forgotResetToken = '';
          this.forgotNewPassword = '';
          this.forgotConfirmPassword = '';
        },
        error: (err) => {
          this.forgotLoading = false;
          this.forgotError = err.error?.error || 'Failed to reset password. Please try again.';
        }
      });
      return;
    }

    this.closeForgotPassword();
  }

  backToRequestStep() {
    this.forgotStep = 'request';
    this.forgotCode = '';
    this.forgotNewPassword = '';
    this.forgotConfirmPassword = '';
    this.forgotResetToken = '';
    this.forgotError = '';
    this.forgotInfo = '';
    this.stopCodeTimer();
    this.hasSentCode = false;
    this.codeCooldown = 0;
  }

  get codeButtonLabel(): string {
    if (!this.hasSentCode) return 'Send code';
    if (this.codeCooldown > 0) return `Resend code (${this.codeCooldown}s)`;
    return 'Resend code';
  }

  get forgotContinueLabel(): string {
    if (this.forgotStep === 'request') return 'Continue';
    if (this.forgotStep === 'verify') return this.forgotLoading ? 'Verifying...' : 'Verify code';
    if (this.forgotStep === 'reset') return this.forgotLoading ? 'Resetting...' : 'Reset password';
    return 'Done';
  }

  get forgotTitle(): string {
    if (this.forgotStep === 'verify') return 'Verify code';
    if (this.forgotStep === 'reset') return 'Set new password';
    if (this.forgotStep === 'success') return 'Password updated';
    return 'Forgot password';
  }

  get forgotSubtitle(): string {
    if (this.forgotStep === 'verify') return 'Enter the verification code sent to your email.';
    if (this.forgotStep === 'reset') return 'Choose a strong new password for your account.';
    if (this.forgotStep === 'success') return 'You can now sign in with your new password.';
    return 'Enter your registered email to receive a verification code.';
  }

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
    this.emailError = '';
    this.passwordError = '';
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
      this.passwordError = 'Please enter your password.';
      this.error = 'Please fix the highlighted fields.';
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

  private resetForgotFlow() {
    this.forgotStep = 'request';
    this.forgotCode = '';
    this.forgotNewPassword = '';
    this.forgotConfirmPassword = '';
    this.forgotLoading = false;
    this.forgotError = '';
    this.forgotInfo = '';
    this.forgotResetToken = '';
    this.hasSentCode = false;
    this.codeCooldown = 0;
    this.stopCodeTimer();
  }

  private startCodeCooldown() {
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

  private stopCodeTimer() {
    if (this.codeTimer) {
      clearInterval(this.codeTimer);
      this.codeTimer = null;
    }
  }

  private isValidEmail(value: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
  }
}
