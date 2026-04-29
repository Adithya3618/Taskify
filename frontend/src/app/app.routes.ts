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

  // Profile
  {
    path: 'profile',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/profile/profile.component').then((m) => m.ProfileComponent),
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
  {
    path: 'auth/google/callback',
    loadComponent: () =>
      import('./pages/google-callback/google-callback.component').then((m) => m.GoogleCallbackComponent),
  },

  // Board — calendar / planner (must be registered before `board/:id` so the segment `planner` is not parsed as an id)
  {
    path: 'board/:id/planner',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/planner-board/planner-board.component').then((m) => m.PlannerBoardComponent),
  },

  // Kanban board view
  {
    path: 'board/:id/:section',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/board/board.component').then((m) => m.BoardComponent),
  },

  // Backward-compatible board route
  {
    path: 'board/:id',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./pages/board/board.component').then((m) => m.BoardComponent),
  },

  // Fallback
  { path: '**', redirectTo: '' },
];
