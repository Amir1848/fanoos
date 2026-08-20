import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import { AuthService } from './auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);

  const token = auth.token;
  const isLoginRequest = req.url.endsWith('/api/login');

  const authorizedReq =
    token && !isLoginRequest
      ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
      : req;

  return next(authorizedReq).pipe(
    catchError((err: HttpErrorResponse) => {
      if (err.status === 401 && !isLoginRequest) {
        auth.logout();
      }
      return throwError(() => err);
    }),
  );
};
