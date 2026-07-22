import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';
import { ToastService } from '../services/toast.service';

export const ErrorInterceptor: HttpInterceptorFn = (req, next) => {
  const toast = inject(ToastService);
  const auth = inject(AuthService);
  const router = inject(Router);

  return next(req).pipe(
    catchError((error: unknown) => {
      if (error instanceof HttpErrorResponse) {
        const message = resolveErrorMessage(error);
        const isAuthEndpoint =
          req.url.includes('/auth/login') || req.url.includes('/auth/register');

        if (error.status === 401 && !isAuthEndpoint) {
          auth.clearToken();
          router.navigateByUrl('/login');
        }
        if (error.status === 403 && !isAuthEndpoint) {
          router.navigateByUrl('/forbidden');
        }

        toast.error(message);
      }

      return throwError(() => error);
    })
  );
};

function resolveErrorMessage(error: HttpErrorResponse): string {
  const payload = error.error;

  if (typeof payload === 'string') {
    try {
      const parsed = JSON.parse(payload);
      if (typeof parsed === 'object' && parsed !== null) {
        const extracted = readMessageFromPayload(parsed as Record<string, unknown>);
        if (extracted) {
          return extracted;
        }
      }
      return payload;
    } catch (_error) {
      return payload;
    }
  }

  if (payload && typeof payload === 'object') {
    const extracted = readMessageFromPayload(payload as Record<string, unknown>);
    if (extracted) {
      return extracted;
    }
  }

  if (error.status === 0) {
    return 'Network error: unable to reach the backend service.';
  }

  const statusText = error.message || 'Request failed';
  return statusText || `Request failed (HTTP ${error.status})`;
}

function readMessageFromPayload(payload: Record<string, unknown>): string {
  const message = typeof payload['message'] === 'string' ? (payload['message'] as string).trim() : '';
  if (message) {
    return message;
  }

  const code = typeof payload['code'] === 'string' ? (payload['code'] as string).trim() : '';
  const details = extractAny(payload);

  if (!details && code) {
    return `Error ${code}`;
  }

  if (details && code) {
    return `${details} (${code})`;
  }

  return details;
}

function extractAny(payload: Record<string, unknown>): string {
  const errorField = payload['error'];
  if (typeof errorField === 'string' && errorField.trim() !== '') {
    return errorField.trim();
  }

  const detailField = payload['details'];
  if (typeof detailField === 'string' && detailField.trim() !== '') {
    return detailField.trim();
  }
  if (Array.isArray(detailField)) {
    const values = detailField.filter((item) => typeof item === 'string' && item.trim() !== '') as string[];
    if (values.length > 0) {
      return values.join(', ');
    }
  }

  return '';
}
