import { Injectable, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap } from 'rxjs';

const TOKEN_KEY = 'marketino.token';

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
    this.router.navigate(['/login']);
  }
}
