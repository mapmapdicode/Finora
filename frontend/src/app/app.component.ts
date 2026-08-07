import { Component, OnDestroy } from '@angular/core';
import { NavigationEnd, RouterOutlet, Router } from '@angular/router';
import { AsyncPipe, NgClass, NgForOf, NgIf } from '@angular/common';
import { Observable, Subscription } from 'rxjs';
import { AuthService } from './core/services/auth.service';
import { ApiService } from './core/services/api.service';
import { ToastService, ToastMessage } from './core/services/toast.service';

import { LanguageService } from './core/services/language.service';

import { IconComponent } from './shared/icons/icon.component';
import { TranslatePipe } from './shared/pipes/translate.pipe';

export function toastRole(type: ToastMessage['type']): 'alert' | 'status' {
  return type === 'error' ? 'alert' : 'status';
}

export function shouldCloseSidebarOnEscape(sidebarOpen: boolean, viewportWidth: number): boolean {
  return sidebarOpen && viewportWidth < 768;
}

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, NgForOf, NgIf, NgClass, AsyncPipe, IconComponent, TranslatePipe],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent implements OnDestroy {
  isAuthenticated = false;
  toasts$: Observable<ToastMessage[]>;
  isDarkMode = false;
  private tokenSub: Subscription | null = null;
  private navigationSub: Subscription;
  private routeHistory: string[] = [];
  private isNavigatingBack = false;
  currentUrl = '';

  constructor(
    private auth: AuthService,
    private api: ApiService,
    private toastService: ToastService,
    private router: Router,
    public langService: LanguageService,
  ) {
    this.toasts$ = this.toastService.toasts$;
    this.navigationSub = this.router.events.subscribe((event) => {
      if (!(event instanceof NavigationEnd)) return;

      this.currentUrl = event.urlAfterRedirects;
      if (this.isNavigatingBack) {
        this.isNavigatingBack = false;
        return;
      }
      if (this.routeHistory.at(-1) !== this.currentUrl) {
        this.routeHistory.push(this.currentUrl);
      }
    });
    const savedAmountMode = localStorage.getItem('finora.amountDisplayMode');
    if (savedAmountMode === 'compact' || savedAmountMode === 'full') {
      this.amountDisplayMode = savedAmountMode;
    }
    if (localStorage.getItem('finora.theme') === 'dark') {
      this.isDarkMode = true;
      document.documentElement.setAttribute('data-theme', 'dark');
    }
    this.tokenSub = this.auth.token$.subscribe((token) => {
      this.isAuthenticated = !!token;
      if (!token) {
        this.router.navigateByUrl('/login');
        return;
      }

      this.api.getUserSettings().subscribe({
        next: (settings) => {
          this.amountDisplayMode = settings.amountDisplayMode;
          localStorage.setItem('finora.amountDisplayMode', settings.amountDisplayMode);
        },
      });

    });
  }

  ngOnDestroy() {
    this.tokenSub?.unsubscribe();
    this.navigationSub.unsubscribe();
  }

  amountDisplayMode: 'full' | 'compact' = 'full';

  toggleDarkMode() {
    this.isDarkMode = !this.isDarkMode;
    if (this.isDarkMode) {
      document.documentElement.setAttribute('data-theme', 'dark');
      localStorage.setItem('finora.theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
      localStorage.removeItem('finora.theme');
    }
  }

  toggleAmountDisplayMode() {
    const next = this.amountDisplayMode === 'full' ? 'compact' : 'full';
    this.api.updateUserSettings({ amountDisplayMode: next }).subscribe({
      next: (settings) => {
        this.amountDisplayMode = settings.amountDisplayMode;
        localStorage.setItem('finora.amountDisplayMode', settings.amountDisplayMode);
        this.toastService.success(`Đã chuyển chế độ hiển thị số tiền sang: ${settings.amountDisplayMode === 'compact' ? 'Viết tắt' : 'Đầy đủ'}`);
      },
      error: () => this.toastService.error('Không thể lưu lựa chọn hiển thị số tiền.'),
    });
  }

  toastClass(type: 'info' | 'success' | 'error') {
    if (type === 'success') {
      return 'bg-emerald-50 text-emerald-900 border-emerald-300';
    }
    if (type === 'error') {
      return 'bg-rose-50 text-rose-900 border-rose-300';
    }
    return 'bg-sky-50 text-sky-900 border-sky-300';
  }

  toastRole(type: ToastMessage['type']) {
    return toastRole(type);
  }

  closeToast(id: number) {
    this.toastService.remove(id);
  }

  get showBackButton(): boolean {
    return this.currentUrl !== '/dashboard';
  }

  get isPublicRoute(): boolean {
    return this.currentUrl === '/login' || this.currentUrl === '/register';
  }

  goHome() {
    this.router.navigateByUrl('/dashboard');
  }

  goBack() {
    const previousUrl = this.routeHistory.length > 1 ? this.routeHistory[this.routeHistory.length - 2] : '/dashboard';
    if (this.routeHistory.length > 1) {
      this.routeHistory.pop();
    }
    this.isNavigatingBack = true;
    this.router.navigateByUrl(previousUrl);
  }

  logout() {
    this.auth.clearToken();
    this.router.navigateByUrl('/login');
  }
}
