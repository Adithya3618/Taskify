import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap, of, throwError } from 'rxjs';
import { catchError, map } from 'rxjs/operators';

export interface AuthUser {
  id?: string;
  name: string;
  email: string;
  role?: string;
}

export interface LoginResponse {
  user: AuthUser;
  token: string;
}

export interface ForgotPasswordResponse {
  message: string;
}

export interface VerifyOtpResponse {
  reset_token: string;
}

export interface ResetPasswordResponse {
  message: string;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly tokenKey = 'taskify.auth.token';
  private readonly sessionKey = 'taskify.auth.session';
  private readonly apiUrl = 'api/auth';

  constructor(private http: HttpClient) {}

  register(input: { name: string; email: string; password: string }): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${this.apiUrl}/register`, input).pipe(
      tap(response => {
        this.setToken(response.token);
        this.setSession(response.user);
      })
    );
  }

  login(email: string, password: string): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${this.apiUrl}/login`, { email, password }).pipe(
      tap(response => {
        this.setToken(response.token);
        this.setSession(response.user);
      })
    );
  }

  forgotPassword(email: string): Observable<ForgotPasswordResponse> {
    return this.http.post<ForgotPasswordResponse>(`${this.apiUrl}/forgot-password`, { email });
  }

  verifyResetOtp(email: string, code: string): Observable<VerifyOtpResponse> {
    return this.http.post<VerifyOtpResponse>(`${this.apiUrl}/verify-otp`, { email, code });
  }

  resetPassword(resetToken: string, newPassword: string): Observable<ResetPasswordResponse> {
    return this.http.post<ResetPasswordResponse>(`${this.apiUrl}/reset-password`, {
      reset_token: resetToken,
      new_password: newPassword
    });
  }

  handleGoogleToken(params: { token: string; name: string; email: string; id: string }): void {
    this.setToken(params.token);
    this.setSession({ id: params.id, name: params.name, email: params.email });
  }

  /** Check if Google OAuth is configured, then redirect. Returns error observable if not available. */
  startGoogleLogin(): Observable<boolean> {
    return this.http.get('/api/auth/google/login', { responseType: 'text', observe: 'response' }).pipe(
      map(() => true),
      catchError(err => {
        if (err.status === 503 || err.status === 500) {
          return throwError(() => new Error('Google sign-in is not available right now.'));
        }
        // status 0 = CORS/network error = backend redirected to Google = it IS configured
        return of(true);
      })
    );
  }

  loginWithGoogleCallback(state: string, code: string): Observable<LoginResponse> {
    return this.http.get<LoginResponse>(`${this.apiUrl}/google/callback`, {
      params: { state, code }
    }).pipe(
      tap(response => {
        this.setToken(response.token);
        this.setSession(response.user);
      })
    );
  }

  logout(): void {
    localStorage.removeItem(this.tokenKey);
    localStorage.removeItem(this.sessionKey);
  }

  getToken(): string | null {
    return localStorage.getItem(this.tokenKey);
  }

  private setToken(token: string): void {
    localStorage.setItem(this.tokenKey, token);
  }

  getCurrentUser(): AuthUser | null {
    const raw = localStorage.getItem(this.sessionKey);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as AuthUser;
    } catch {
      return null;
    }
  }

  private setSession(user: AuthUser): void {
    localStorage.setItem(this.sessionKey, JSON.stringify(user));
  }

  updateCurrentUser(patch: Partial<AuthUser>): AuthUser | null {
    const current = this.getCurrentUser();
    if (!current) return null;
    const updated: AuthUser = { ...current, ...patch };
    this.setSession(updated);
    return updated;
  }

  isAuthenticated(): boolean {
    return !!this.getToken();
  }
}
