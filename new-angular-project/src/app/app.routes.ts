import { Routes } from '@angular/router';
import { authGuard, loginGuard } from './core/auth/auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    canActivate: [loginGuard],
    loadComponent: () => import('./features/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () => import('./features/shell/shell.component').then((m) => m.ShellComponent),
    children: [
      { path: '', pathMatch: 'full', redirectTo: 'dashboard' },
      {
        path: 'dashboard',
        loadComponent: () =>
          import('./features/dashboard/dashboard.component').then((m) => m.DashboardComponent),
      },
      {
        path: 'users',
        loadComponent: () =>
          import('./features/users/users-list.component').then((m) => m.UsersListComponent),
      },
      {
        path: 'users/new',
        loadComponent: () =>
          import('./features/users/user-form.component').then((m) => m.UserFormComponent),
      },
      {
        path: 'users/:id',
        loadComponent: () =>
          import('./features/users/user-form.component').then((m) => m.UserFormComponent),
      },
      {
        path: 'roles',
        loadComponent: () =>
          import('./features/roles/roles-list.component').then((m) => m.RolesListComponent),
      },
      {
        path: 'roles/new',
        loadComponent: () =>
          import('./features/roles/role-form.component').then((m) => m.RoleFormComponent),
      },
      {
        path: 'roles/:id',
        loadComponent: () =>
          import('./features/roles/role-form.component').then((m) => m.RoleFormComponent),
      },
      {
        path: 'roles/:id/access',
        loadComponent: () =>
          import('./features/roles/role-access.component').then((m) => m.RoleAccessComponent),
      },
    ],
  },
  { path: '**', redirectTo: '' },
];
