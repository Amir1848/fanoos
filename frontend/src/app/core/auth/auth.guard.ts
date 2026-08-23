import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { map } from 'rxjs';
import { AuthService } from './auth.service';

/**
 * Auth state lives in an HttpOnly session cookie, so guards must ask the
 * server whether the session is valid (GET /api/auth/resources) instead of
 * inspecting a token locally.
 */
export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  return auth.ensureSessionLoaded().pipe(
    map((ok) => (ok ? true : router.createUrlTree(['/login']))),
  );
};

export const loginGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  return auth.ensureSessionLoaded().pipe(
    map((ok) => (ok ? router.createUrlTree(['/']) : true)),
  );
};

/** Guards for the mobile PWA area (/m) — redirect within the area. */
export const mobileAuthGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  return auth.ensureSessionLoaded().pipe(
    map((ok) => (ok ? true : router.createUrlTree(['/m/login']))),
  );
};

export const mobileLoginGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  return auth.ensureSessionLoaded().pipe(
    map((ok) => (ok ? router.createUrlTree(['/m']) : true)),
  );
};
