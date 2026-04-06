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
    const errorParam = params.get('error');

    if (errorParam) {
      this.error = errorParam === 'access_denied'
        ? 'Google sign-in was cancelled.'
        : 'Google sign-in failed. Please try again.';
      return;
    }

    // Backend redirect flow: token passed directly as query param
    const token = params.get('token');
    if (token) {
      this.authService.handleGoogleToken({
        token,
        name: params.get('name') || '',
        email: params.get('email') || '',
        id: params.get('id') || ''
      });
      this.router.navigate(['/boards']);
      return;
    }

    // Fallback: frontend-redirect flow (state + code)
    const state = params.get('state');
    const code = params.get('code');
    if (!state || !code) {
      this.error = 'Invalid callback. Please try signing in again.';
      return;
    }

    this.authService.loginWithGoogleCallback(state, code).subscribe({
      next: () => this.router.navigate(['/boards']),
      error: () => { this.error = 'Google sign-in failed. Please try again.'; }
    });
  }

  goToLogin(): void {
    this.router.navigate(['/login']);
  }
}
