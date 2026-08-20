import { HttpErrorResponse } from '@angular/common/http';

/** Extracts a human-readable message from a backend error response. */
export function apiErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof HttpErrorResponse) {
    const serverMessage = (err.error as { error?: string } | null)?.error;
    if (serverMessage) {
      return serverMessage;
    }
    if (err.status === 0) {
      return 'ارتباط با سرور برقرار نشد';
    }
  }
  return fallback;
}
