import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';

const IDEMPOTENCY_SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS']);

export const AuthInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const requestMethod = req.method.toUpperCase();
  const isMutation = !IDEMPOTENCY_SAFE_METHODS.has(requestMethod);
  const isAuthEndpoint =
    req.url.includes('/auth/login') ||
    req.url.includes('/auth/register') ||
    req.url.includes('/auth/login/') ||
    req.url.includes('/auth/register/');

  if (isMutation && auth.isAuthenticated() && auth.isViewerRole() && !isAuthEndpoint) {
    const error = new HttpErrorResponse({
      status: 403,
      statusText: 'Forbidden',
      error: {
        code: 'FORBIDDEN',
        message: 'This workspace is read-only. Switch to an editable workspace to perform this action.',
      },
      url: req.url,
    });
    return throwError(() => error);
  }

  const token = localStorage.getItem('wealthos.token');
  const workspaceId = localStorage.getItem('wealthos.workspace');

  let headers = req.headers;

  if (token) {
    headers = headers.set('Authorization', `Bearer ${token}`);
  }

  if (workspaceId) {
    headers = headers.set('x-workspace-id', workspaceId);
  }

  if (!headers.has('X-Request-ID')) {
    headers = headers.set('X-Request-ID', generateRequestId());
  }

  const method = req.method.toUpperCase();
  if (!IDEMPOTENCY_SAFE_METHODS.has(method) && !headers.has('Idempotency-Key')) {
    headers = headers.set('Idempotency-Key', generateIdempotencyKey());
  }

  return next(req.clone({ headers }));
};

function generateRequestId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `req-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

function generateIdempotencyKey(): string {
  const now = Date.now().toString(36);
  const random = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID()
    : Math.random().toString(36).slice(2);

  return `idemp-${now}-${random}`;
}
