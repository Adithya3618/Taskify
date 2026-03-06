import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap, catchError, of } from 'rxjs';

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
