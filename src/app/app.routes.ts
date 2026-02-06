import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', loadComponent: () => import('./components/board/board.component').then(m => m.BoardComponent) },
  { path: '**', redirectTo: '' }
];
