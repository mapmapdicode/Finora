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
    return 'Không thể kết nối máy chủ. Kiểm tra mạng rồi thử lại.';
  }

  const statusMessages: Record<number, string> = {
    400: 'Yêu cầu không hợp lệ.',
    401: 'Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.',
    403: 'Bạn không có quyền thực hiện thao tác này.',
    404: 'Không tìm thấy dữ liệu yêu cầu.',
    409: 'Dữ liệu đang xung đột. Vui lòng kiểm tra lại.',
    422: 'Dữ liệu chưa hợp lệ. Vui lòng kiểm tra lại.',
    500: 'Hệ thống đang gặp sự cố. Vui lòng thử lại sau.',
  };
  const rawMessage = error.message?.trim();
  if (rawMessage && !/^Http failure response/i.test(rawMessage)) {
    return rawMessage;
  }
  return statusMessages[error.status] || `Yêu cầu không thành công (HTTP ${error.status}).`;
}

function readMessageFromPayload(payload: Record<string, unknown>): string {
  const message = typeof payload['message'] === 'string' ? (payload['message'] as string).trim() : '';
  if (message) {
    return message;
  }

  const code = typeof payload['code'] === 'string' ? (payload['code'] as string).trim() : '';
  const details = extractAny(payload);

  if (!details && code) {
    return `Lỗi hệ thống (mã ${code}).`;
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
