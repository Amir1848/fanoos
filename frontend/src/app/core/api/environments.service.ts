import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DataviewService } from './dataview.service';
import { EnvironmentListItem, EnvironmentSave } from './models';

@Injectable({ providedIn: 'root' })
export class EnvironmentsService {
  private readonly http = inject(HttpClient);
  private readonly dataviews = inject(DataviewService);

  list(pageNumber: number, pageSize: number): Observable<EnvironmentListItem[]> {
    return this.dataviews.fetch<EnvironmentListItem>('all-environments', pageNumber, pageSize);
  }

  getById(id: number): Observable<EnvironmentListItem | null> {
    return this.dataviews.findById<EnvironmentListItem>('all-environments', id);
  }

  create(environment: EnvironmentSave): Observable<{ id: number }> {
    return this.http.post<{ id: number }>('/api/environments/create', environment);
  }

  update(environment: EnvironmentSave): Observable<void> {
    return this.http.post<void>('/api/environments/update', environment);
  }

  delete(ids: number[]): Observable<void> {
    return this.http.post<void>('/api/environments/delete', { ids });
  }
}
