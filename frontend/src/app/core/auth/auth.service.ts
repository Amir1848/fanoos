import { Injectable, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, map, of, tap } from 'rxjs';

const TOKEN_KEY = 'fanoos.token';

interface JwtPayload {
  username?: string;
  exp?: number;
}

function decodeJwt(token: string): JwtPayload | null {
  try {
    const payload = token.split('.')[1];
    const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(json) as JwtPayload;
  } catch {
    return null;
  }
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly router = inject(Router);

  private readonly tokenSignal = signal<string | null>(localStorage.getItem(TOKEN_KEY));

  private readonly resourcesSignal = signal<ReadonlySet<string>>(new Set());
  private resourcesLoaded = false;

  readonly username = computed(() => {
    const token = this.tokenSignal();
    return token ? (decodeJwt(token)?.username ?? '') : '';
  });

  get token(): string | null {
    return this.tokenSignal();
  }

  isAuthenticated(): boolean {
    const token = this.tokenSignal();
    if (!token) {
      return false;
    }

    const payload = decodeJwt(token);
    if (!payload?.exp) {
      return false;
    }

    return payload.exp * 1000 > Date.now();
  }

  /** Resource keys of the logged-in user, fetched once per session from the backend. */
  get resources(): ReadonlySet<string> {
    return this.resourcesSignal();
  }

  hasResource(resourceKey: string): boolean {
    return this.resourcesSignal().has(resourceKey);
  }

  ensureResourcesLoaded(): Observable<ReadonlySet<string>> {
    if (this.resourcesLoaded && this.isAuthenticated()) {
      return of(this.resourcesSignal());
    }

    return this.http.get<string[]>('/api/auth/resources').pipe(
      map((keys) => new Set(keys) as ReadonlySet<string>),
      tap((resources) => {
        this.resourcesSignal.set(resources);
        this.resourcesLoaded = true;
      }),
    );
  }

  login(username: string, password: string): Observable<{ token: string }> {
    return this.http.post<{ token: string }>('/api/login', { username, password }).pipe(
      tap((res) => {
        localStorage.setItem(TOKEN_KEY, res.token);
        this.tokenSignal.set(res.token);
      }),
    );
  }

  logout(): void {
    localStorage.removeItem(TOKEN_KEY);
    this.tokenSignal.set(null);
    this.resourcesSignal.set(new Set());
    this.resourcesLoaded = false;
    // Stay inside the area the user came from (mobile PWA vs admin panel).
    const mobile = this.router.url.startsWith('/m');
    this.router.navigate([mobile ? '/m/login' : '/login']);
  }
}
