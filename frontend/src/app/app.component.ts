import { Component, OnDestroy } from '@angular/core';
import { RouterLink, RouterOutlet, RouterLinkActive, Router } from '@angular/router';
import { AsyncPipe, NgClass, NgForOf, NgIf } from '@angular/common';
import { Observable, Subscription } from 'rxjs';
import { AuthService } from './core/services/auth.service';
import { ApiService } from './core/services/api.service';
import { ToastService, ToastMessage } from './core/services/toast.service';
import { Workspace } from './shared/models';

import { LanguageService, SupportedLanguage } from './core/services/language.service';
import { TranslatePipe } from './shared/pipes/translate.pipe';
import { IconComponent } from './shared/icons/icon.component';

type NavItem = { path: string; labelKey: string; icon: string };

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, NgForOf, NgIf, NgClass, AsyncPipe, TranslatePipe, IconComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent implements OnDestroy {
  isAuthenticated = false;
  workspaces: Workspace[] = [];
  selectedWorkspaceId: string | null = null;
  toasts$: Observable<ToastMessage[]>;
  sidebarOpen = true;
  isDarkMode = false;
  private tokenSub: Subscription | null = null;
  private readonly viewerMessage = 'This workspace is read-only. You can only view data.';
  private readonly roleMessageShown = new Set<string>();

  navItems: NavItem[] = [
    { path: '/dashboard', labelKey: 'nav.dashboard', icon: 'dashboard' },
    { path: '/accounts', labelKey: 'nav.accounts', icon: 'account_balance' },
    { path: '/transactions', labelKey: 'nav.transactions', icon: 'payments' },
    { path: '/loans', labelKey: 'nav.loans', icon: 'request_quote' },
    { path: '/assets', labelKey: 'nav.assets', icon: 'savings' },
    { path: '/properties', labelKey: 'nav.properties', icon: 'home' },
    { path: '/budgets', labelKey: 'nav.budgets', icon: 'pie_chart' },
    { path: '/forecast', labelKey: 'nav.forecast', icon: 'analytics' },
    { path: '/portfolios', labelKey: 'nav.portfolios', icon: 'donut_large' },
    { path: '/audit-logs', labelKey: 'nav.audit', icon: 'history' },
    { path: '/sepay', labelKey: 'nav.sepay', icon: 'inbox' },
    { path: '/automation', labelKey: 'nav.automation', icon: 'auto_awesome' },
    { path: '/assistant', labelKey: 'nav.assistant', icon: 'smart_toy' },
  ];

  constructor(
    private auth: AuthService,
    private api: ApiService,
    private toastService: ToastService,
    private router: Router,
    public langService: LanguageService,
  ) {
    this.toasts$ = this.toastService.toasts$;
    this.tokenSub = this.auth.token$.subscribe((token) => {
      this.isAuthenticated = !!token;
      if (!token) {
        this.workspaces = [];
        this.selectedWorkspaceId = null;
        this.router.navigateByUrl('/login');
        return;
      }

      if (this.auth.workspaceId) {
        this.selectedWorkspaceId = this.auth.workspaceId;
      }
      this.loadWorkspaces();
    });
  }

  ngOnDestroy() {
    this.tokenSub?.unsubscribe();
  }

  toggleSidebar() {
    this.sidebarOpen = !this.sidebarOpen;
  }

  changeLanguage(event: Event) {
    const lang = (event.target as HTMLSelectElement).value as SupportedLanguage;
    this.langService.setLanguage(lang);
  }

  toggleDarkMode() {
    this.isDarkMode = !this.isDarkMode;
    if (this.isDarkMode) {
      document.documentElement.setAttribute('data-theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }
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

  closeToast(id: number) {
    this.toastService.remove(id);
  }

  private loadWorkspaces() {
    this.api.getWorkspaces().subscribe({
      next: (workspaces) => {
        this.workspaces = workspaces;
        this.auth.syncWorkspaceRoles(workspaces);
        if (!this.selectedWorkspaceId && workspaces.length > 0) {
          this.selectedWorkspaceId = workspaces[0].id;
          this.auth.saveWorkspace(workspaces[0].id);
          this.showViewerBannerIfNeeded(workspaces[0].id);
          return;
        }
        if (this.selectedWorkspaceId) {
          const exists = workspaces.some((workspace) => workspace.id === this.selectedWorkspaceId);
          if (!exists && workspaces.length > 0) {
            this.selectedWorkspaceId = workspaces[0].id;
            this.auth.saveWorkspace(workspaces[0].id);
            this.showViewerBannerIfNeeded(workspaces[0].id);
            return;
          }
          if (exists) {
            this.auth.saveWorkspace(this.selectedWorkspaceId);
            this.showViewerBannerIfNeeded(this.selectedWorkspaceId);
          }
        }
      },
      error: () => {
        this.workspaces = [];
      },
    });
  }

  onWorkspaceChange(event: Event) {
    const value = (event.target as HTMLSelectElement).value;
    if (!value) return;
    this.selectedWorkspaceId = value;
    this.auth.saveWorkspace(value);
    this.showViewerBannerIfNeeded(value);
    window.location.reload();
  }

  private showViewerBannerIfNeeded(workspaceId: string) {
    const role = this.auth.workspaceRole;
    if (role === 'viewer' && !this.roleMessageShown.has(workspaceId)) {
      this.roleMessageShown.add(workspaceId);
      this.toastService.error(this.viewerMessage);
    }
  }

  logout() {
    this.auth.clearToken();
    this.router.navigateByUrl('/login');
  }
}
