import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-google-callback',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './google-callback.component.html',
  styleUrl: './google-callback.component.scss',
})
export class GoogleCallbackComponent implements OnInit {
  error = '';

  constructor(private authService: AuthService, private router: Router) {}

  ngOnInit(): void {
    const params = new URLSearchParams(window.location.search);
    const state = params.get('state');
    const code = params.get('code');
    const errorParam = params.get('error');

    if (errorParam) {
      this.error = errorParam === 'access_denied'
        ? 'Google sign-in was cancelled.'
        : 'Google sign-in failed. Please try again.';
      return;
    }

    if (!state || !code) {
      this.error = 'Invalid callback. Please try signing in again.';
      return;
    }

    this.authService.loginWithGoogleCallback(state, code).subscribe({
      next: () => {
        this.router.navigate(['/boards']);
      },
      error: (err) => {
        const msg = err.error?.error || err.error?.message || '';
        if (msg.toLowerCase().includes('not configured')) {
          this.error = 'Google sign-in is not available right now.';
        } else if (msg.toLowerCase().includes('state') || msg.toLowerCase().includes('oauth')) {
          this.error = 'Sign-in session expired. Please try again.';
        } else {
          this.error = 'Google sign-in failed. Please try again.';
        }
      }
    });
  }

  goToLogin(): void {
    this.router.navigate(['/login']);
  }
}
