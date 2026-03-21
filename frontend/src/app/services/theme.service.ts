import { Injectable } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private _isDark = true;

  get isDark(): boolean { return this._isDark; }

  constructor() {
    const saved = localStorage.getItem('taskify-theme');
    this._isDark = saved !== 'light';
    this._apply();
  }

  toggle(): void {
    this._isDark = !this._isDark;
    localStorage.setItem('taskify-theme', this._isDark ? 'dark' : 'light');
    this._apply();
  }

  private _apply(): void {
    document.documentElement.setAttribute('data-theme', this._isDark ? 'dark' : 'light');
  }
}
