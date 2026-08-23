import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import { AuthService } from './auth.service';

/**
 * Requests that must never trigger the automatic 401 logout redirect:
 * the session probe (guards own the not-logged-in redirect — reacting here
 * too races their navigation and causes redirect loops), login, and logout.
 */
const noLogoutUrls = ['/api/auth/resources', '/api/login', '/api/auth/logout'];

/**
 * Credentials ride in the HttpOnly session cookie set by the server — no
 * Authorization header to attach. This interceptor only watches for 401s
 * (expired/invalid session) on regular API calls and routes the user back
 * to their area's login page.
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);

  const skipLogout = noLogoutUrls.some((url) => req.url.endsWith(url));

  return next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      if (err.status === 401 && !skipLogout) {
        auth.logout();
      }
      return throwError(() => err);
    }),
  );
};
