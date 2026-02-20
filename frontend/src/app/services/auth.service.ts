import { Injectable } from '@angular/core';

export interface AuthUser {
  name: string;
  email: string;
}

interface StoredUser extends AuthUser {
  password: string;
}

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly usersKey = 'taskify.auth.users';
  private readonly sessionKey = 'taskify.auth.session';

  register(input: { name: string; email: string; password: string }): { ok: boolean; error?: string } {
    const users = this.getUsers();
    const email = input.email.trim().toLowerCase();
    if (users.some(u => u.email === email)) {
      return { ok: false, error: 'An account with this email already exists.' };
    }

    users.push({
      name: input.name.trim(),
      email,
      password: input.password
    });
    this.saveUsers(users);
    this.setSession({ name: input.name.trim(), email });
    return { ok: true };
  }

  login(email: string, password: string): { ok: boolean; error?: string } {
    const normalizedEmail = email.trim().toLowerCase();
    const user = this.getUsers().find(u => u.email === normalizedEmail && u.password === password);
    if (!user) {
      return { ok: false, error: 'Invalid email or password.' };
    }

    this.setSession({ name: user.name, email: user.email });
    return { ok: true };
  }

  logout(): void {
    localStorage.removeItem(this.sessionKey);
  }

  getCurrentUser(): AuthUser | null {
    const raw = localStorage.getItem(this.sessionKey);
    if (!raw) return null;
    try {
      const parsed = JSON.parse(raw) as AuthUser;
      if (!parsed?.email || !parsed?.name) return null;
      return parsed;
    } catch {
      return null;
    }
  }

  isAuthenticated(): boolean {
    return !!this.getCurrentUser();
  }

  private getUsers(): StoredUser[] {
    const raw = localStorage.getItem(this.usersKey);
    if (!raw) return [];
    try {
      const parsed = JSON.parse(raw) as StoredUser[];
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }

  private saveUsers(users: StoredUser[]): void {
    localStorage.setItem(this.usersKey, JSON.stringify(users));
  }

  private setSession(user: AuthUser): void {
    localStorage.setItem(this.sessionKey, JSON.stringify(user));
  }
}
