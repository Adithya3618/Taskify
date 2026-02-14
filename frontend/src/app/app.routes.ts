import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', loadComponent: () => import('./pages/home/home.component').then(m => m.HomeComponent) },
  { path: 'board/:id', loadComponent: () => import('./pages/board/board.component').then(m => m.BoardComponent) },
  { path: '**', redirectTo: '' }
];
