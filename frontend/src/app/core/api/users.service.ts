import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DataviewService } from './dataview.service';
import { UserListItem, UserSave } from './models';

@Injectable({ providedIn: 'root' })
export class UsersService {
  private readonly http = inject(HttpClient);
  private readonly dataviews = inject(DataviewService);

  list(pageNumber: number, pageSize: number): Observable<UserListItem[]> {
    return this.dataviews.fetch<UserListItem>('all-users', pageNumber, pageSize);
  }

  getById(id: number): Observable<UserListItem | null> {
    return this.dataviews.findById<UserListItem>('all-users', id);
  }

  create(user: UserSave): Observable<{ id: number }> {
    return this.http.post<{ id: number }>('/api/users/create', user);
  }

  update(user: UserSave): Observable<{ id: number }> {
    return this.http.post<{ id: number }>('/api/users/update', user);
  }

  delete(ids: number[]): Observable<void> {
    return this.http.post<void>('/api/users/delete', { ids });
  }
}
