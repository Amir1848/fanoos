import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DataviewService } from './dataview.service';
import { DeviceLinkListItem, DeviceLinkSave } from './models';

/** A single home's control wiring — one page comfortably covers all of them. */
const ALL_LINKS_PAGE_SIZE = 500;

@Injectable({ providedIn: 'root' })
export class DeviceLinksService {
  private readonly http = inject(HttpClient);
  private readonly dataviews = inject(DataviewService);

  listForEnvironment(environmentId: number): Observable<DeviceLinkListItem[]> {
    return this.dataviews.fetch<DeviceLinkListItem>('all-device-links', 1, ALL_LINKS_PAGE_SIZE, {
      environmentId,
    });
  }

  create(link: DeviceLinkSave): Observable<{ id: number }> {
    return this.http.post<{ id: number }>('/api/device-links/create', link);
  }

  delete(ids: number[]): Observable<void> {
    return this.http.post<void>('/api/device-links/delete', { ids });
  }
}
