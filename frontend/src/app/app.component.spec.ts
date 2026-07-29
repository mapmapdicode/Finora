import { ICON_PATHS } from './shared/icons/icon.registry';
import { appRoutes } from './app.routes';
import { shouldCloseSidebarOnEscape, toastRole } from './app.component';

describe('Test bootstrap', () => {
  it('should keep the test runner green', () => {
    expect(true).toBeTrue();
  });

  it('keeps named UI icon aliases mapped to real paths', () => {
    expect(ICON_PATHS['progress_activity']?.length).toBeGreaterThan(0);
    expect(ICON_PATHS['cloud_off']?.length).toBeGreaterThan(0);
    expect(ICON_PATHS['smart_toy']).toEqual(ICON_PATHS['robot']);
    expect(ICON_PATHS['arrow_forward']?.length).toBeGreaterThan(0);
  });

  it('assigns non-blocking and assertive toast roles by severity', () => {
    expect(toastRole('error')).toBe('alert');
    expect(toastRole('success')).toBe('status');
    expect(toastRole('info')).toBe('status');
  });

  it('only closes an open mobile drawer on Escape', () => {
    expect(shouldCloseSidebarOnEscape(true, 390)).toBeTrue();
    expect(shouldCloseSidebarOnEscape(false, 390)).toBeFalse();
    expect(shouldCloseSidebarOnEscape(true, 1024)).toBeFalse();
  });

  it('keeps the audited route inventory explicit', () => {
    const paths = appRoutes.map((route) => route.path);
    expect(paths).toEqual(
      jasmine.arrayContaining([
        'login',
        'register',
        'dashboard',
        'accounts',
        'transactions',
        'loans',
        'assets',
        'properties',
        'forecast',
        'budgets',
        'sepay',
        'automation',
        'assistant',
        'audit-logs',
        'forbidden',
        'portfolios',
        '**',
      ]),
    );
  });
});
