import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map, of, expand, reduce, takeWhile } from 'rxjs';

export interface DataviewRequest {
  parameters?: unknown;
  pageSize: number;
  pageNumber: number;
}

const LOOKUP_PAGE_SIZE = 200;
const LOOKUP_MAX_PAGES = 25;

@Injectable({ providedIn: 'root' })
export class DataviewService {
  private readonly http = inject(HttpClient);

  fetch<T>(key: string, pageNumber: number, pageSize: number, parameters: unknown = null): Observable<T[]> {
    const body: DataviewRequest = { parameters, pageSize, pageNumber };
    return this.http.post<T[]>(`/api/dataviews/${key}`, body);
  }

  /**
   * The backend exposes no get-by-id endpoint, so a record opened by direct
   * URL is located by scanning the dataview pages until it is found.
   */
  findById<T extends { id: number }>(key: string, id: number): Observable<T | null> {
    let page = 1;
    return this.fetch<T>(key, page, LOOKUP_PAGE_SIZE).pipe(
      expand((rows) =>
        rows.length === LOOKUP_PAGE_SIZE && !rows.some((r) => r.id === id) && page < LOOKUP_MAX_PAGES
          ? this.fetch<T>(key, ++page, LOOKUP_PAGE_SIZE)
          : of<T[]>([]),
      ),
      takeWhile((rows) => rows.length > 0, true),
      reduce<T[], T | null>((found, rows) => found ?? rows.find((r) => r.id === id) ?? null, null),
      map((found) => found ?? null),
    );
  }
}
