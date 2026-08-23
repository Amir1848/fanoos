import { Injectable, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, of, tap, catchError, map, switchMap } from 'rxjs';

/**
 * Session-based auth: the server owns the session in an HttpOnly cookie
 * (set by POST /api/login, cleared by POST /api/auth/logout). The browser
 * never stores a token; this service only caches what the server tells us
 * about the current session (username + resource keys) so guards and
 * templates can react synchronously. Every page load re-verifies the
 * session via GET /api/auth/resources.
 */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly router = inject(Router);

  private readonly authedSignal = signal(false);
  private readonly usernameSignal = signal('');
  private readonly resourcesSignal = signal<ReadonlySet<string>>(new Set());
  private sessionLoaded = false;

  readonly username = computed(() => this.usernameSignal());

  get isAuthenticated(): boolean {
    return this.authedSignal();
  }

  /** Resource keys of the logged-in user, fetched once per session from the backend. */
  get resources(): ReadonlySet<string> {
    return this.resourcesSignal();
  }

  hasResource(resourceKey: string): boolean {
    return this.resourcesSignal().has(resourceKey);
  }

  /**
   * Verifies the session cookie against the server and refreshes the
   * cached username/resources. Emits false when there is no valid session.
   */
  ensureSessionLoaded(): Observable<boolean> {
    if (this.sessionLoaded && this.authedSignal()) {
      return of(true);
    }

    return this.fetchSession().pipe(
      map(() => true),
      catchError(() => of(false)),
    );
  }

  login(username: string, password: string): Observable<SessionInfo> {
    return this.http.post<{ username: string }>('/api/login', { username, password }).pipe(
      tap((res) => {
        this.authedSignal.set(true);
        this.usernameSignal.set(res.username);
        this.sessionLoaded = true;
      }),
      // The login response carries no resource keys — load them now so
      // menus gated by *appAuthorize render right after login (guards skip
      // the fetch once the session is already marked loaded).
      switchMap(() => this.fetchSession()),
    );
  }

  private fetchSession(): Observable<SessionInfo> {
    return this.http.get<SessionInfo>('/api/auth/resources').pipe(
      tap((info) => {
        this.authedSignal.set(true);
        this.usernameSignal.set(info.username);
        this.resourcesSignal.set(new Set(info.resources));
        this.sessionLoaded = true;
      }),
    );
  }

  logout(): void {
    // Clear the server-side session (cookie is HttpOnly, so only the
    // server can clear it). Fire-and-forget: local state resets either way.
    this.http.post('/api/auth/logout', {}).subscribe({ error: () => undefined });

    this.authedSignal.set(false);
    this.usernameSignal.set('');
    this.resourcesSignal.set(new Set());
    this.sessionLoaded = false;
    // Stay inside the area the user came from (mobile PWA vs admin panel).
    const mobile = this.router.url.startsWith('/m');
    this.router.navigate([mobile ? '/m/login' : '/login']);
  }
}

interface SessionInfo {
  username: string;
  resources: string[];
}
