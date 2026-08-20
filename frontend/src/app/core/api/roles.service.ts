import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DataviewService } from './dataview.service';
import { RoleListItem, RoleSave } from './models';

@Injectable({ providedIn: 'root' })
export class RolesService {
  private readonly http = inject(HttpClient);
  private readonly dataviews = inject(DataviewService);

  list(pageNumber: number, pageSize: number): Observable<RoleListItem[]> {
    return this.dataviews.fetch<RoleListItem>('all-roles', pageNumber, pageSize);
  }

  getById(id: number): Observable<RoleListItem | null> {
    return this.dataviews.findById<RoleListItem>('all-roles', id);
  }

  create(role: RoleSave): Observable<{ id: number }> {
    return this.http.post<{ id: number }>('/api/roles/create', role);
  }

  update(role: RoleSave): Observable<void> {
    return this.http.post<void>('/api/roles/update', role);
  }

  delete(ids: number[]): Observable<void> {
    return this.http.post<void>('/api/roles/delete', { ids });
  }
}
