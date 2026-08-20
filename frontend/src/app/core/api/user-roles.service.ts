import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { RoleListItem } from './models';

@Injectable({ providedIn: 'root' })
export class UserRolesService {
  private readonly http = inject(HttpClient);

  listAllRoles(): Observable<RoleListItem[]> {
    return this.http.post<RoleListItem[]>('/api/user-roles/roles', {});
  }

  listForUser(userId: number): Observable<RoleListItem[]> {
    return this.http.post<RoleListItem[]>('/api/user-roles/list', { userId });
  }

  assign(userId: number, roleIds: number[]): Observable<void> {
    if (roleIds.length === 0) {
      return of(undefined);
    }
    return this.http.post<void>('/api/user-roles/assign', { userId, roleIds });
  }

  remove(userId: number, roleIds: number[]): Observable<void> {
    if (roleIds.length === 0) {
      return of(undefined);
    }
    return this.http.post<void>('/api/user-roles/remove', { userId, roleIds });
  }
}
