import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { DataviewService } from './dataview.service';
import { ResourceItem } from './models';

/** Resources are a small, fixed permission catalog — one page comfortably covers all of them. */
const ALL_RESOURCES_PAGE_SIZE = 500;

@Injectable({ providedIn: 'root' })
export class RoleResourcesService {
  private readonly http = inject(HttpClient);
  private readonly dataviews = inject(DataviewService);

  listAllResources(): Observable<ResourceItem[]> {
    return this.dataviews.fetch<ResourceItem>('all-resources', 1, ALL_RESOURCES_PAGE_SIZE);
  }

  listForRole(roleId: number): Observable<ResourceItem[]> {
    return this.http.post<ResourceItem[]>('/api/role-resources/list', { roleId });
  }

  assign(roleId: number, resourceIds: number[]): Observable<void> {
    if (resourceIds.length === 0) {
      return of(undefined);
    }
    return this.http.post<void>('/api/role-resources/assign', { roleId, resourceIds });
  }

  remove(roleId: number, resourceIds: number[]): Observable<void> {
    if (resourceIds.length === 0) {
      return of(undefined);
    }
    return this.http.post<void>('/api/role-resources/remove', { roleId, resourceIds });
  }
}
