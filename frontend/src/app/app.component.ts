import { Component, OnDestroy } from '@angular/core';
import { RouterLink, RouterOutlet, RouterLinkActive, Router } from '@angular/router';
import { AsyncPipe, NgClass, NgForOf, NgIf } from '@angular/common';
import { Observable, Subscription } from 'rxjs';
import { AuthService } from './core/services/auth.service';
import { ApiService } from './core/services/api.service';
import { ToastService, ToastMessage } from './core/services/toast.service';
import { Workspace } from './shared/models';

type NavItem = { path: string; label: string };

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, NgForOf, NgIf, NgClass, AsyncPipe],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent implements OnDestroy {
  isAuthenticated = false;
  workspaces: Workspace[] = [];
  selectedWorkspaceId: string | null = null;
  toasts$: Observable<ToastMessage[]>;
  private tokenSub: Subscription | null = null;
  private readonly viewerMessage = 'This workspace is read-only. You can only view data.';
  private readonly roleMessageShown = new Set<string>();

  navItems: NavItem[] = [
    { path: '/dashboard', label: 'Dashboard' },
    { path: '/accounts', label: 'Accounts' },
    { path: '/transactions', label: 'Transactions' },
    { path: '/loans', label: 'Loans' },
    { path: '/assets', label: 'Assets' },
    { path: '/properties', label: 'Properties' },
    { path: '/budgets', label: 'Budgets' },
    { path: '/forecast', label: 'Forecast' },
    { path: '/portfolios', label: 'Portfolios' },
    { path: '/audit-logs', label: 'Audit Logs' },
    { path: '/sepay', label: 'Inbox SePay' },
    { path: '/automation', label: 'Automation Rules' },
    { path: '/assistant', label: 'Assistant' },
  ];

  constructor(
    private auth: AuthService,
    private api: ApiService,
    private toastService: ToastService,
    private router: Router,
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

  toastClass(type: 'info' | 'success' | 'error') {
    if (type === 'success') {
      return 'bg-emerald-50 border-emerald-300 text-emerald-900';
    }
    if (type === 'error') {
      return 'bg-rose-50 border-rose-300 text-rose-900';
    }
    return 'bg-sky-50 border-sky-300 text-sky-900';
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
