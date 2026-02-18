import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';

export const routes: Routes = [
  // Home
  {
    path: '',
    loadComponent: () =>
      import('./pages/welcome/welcome.component').then((m) => m.WelcomeComponent),
  },

  // Boards
  {
    path: 'boards',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/home/home.component').then((m) => m.HomeComponent),
  },

  // Features
  {
    path: 'features',
    loadComponent: () =>
      import('./pages/features/features.component').then((m) => m.FeaturesComponent),
  },

  // About
  {
    path: 'about',
    loadComponent: () =>
      import('./pages/about/about.component').then((m) => m.AboutComponent),
  },

  // Auth
  {
    path: 'login',
    loadComponent: () =>
      import('./pages/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'signup',
    loadComponent: () =>
      import('./pages/signup/signup.component').then((m) => m.SignupComponent),
  },

  // Kanban board view
  {
    path: 'board/:id',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/board/board.component').then((m) => m.BoardComponent),
  },

  // Fallback
  { path: '**', redirectTo: '' },
];
