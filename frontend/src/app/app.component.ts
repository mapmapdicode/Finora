import { Component, HostListener, OnDestroy } from '@angular/core';
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
type CommandItem = NavItem & { description: string; queryParams?: Record<string, string> };

export function toastRole(type: ToastMessage['type']): 'alert' | 'status' {
  return type === 'error' ? 'alert' : 'status';
}

export function shouldCloseSidebarOnEscape(sidebarOpen: boolean, viewportWidth: number): boolean {
  return sidebarOpen && viewportWidth < 768;
}

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, NgForOf, NgIf, NgClass, AsyncPipe, TranslatePipe],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent implements OnDestroy {
  isAuthenticated = false;
  workspaces: Workspace[] = [];
  selectedWorkspaceId: string | null = null;
  toasts$: Observable<ToastMessage[]>;
  // Desktop benefits from a persistent rail; on phones an open drawer must be an explicit action.
  sidebarOpen = typeof window === 'undefined' || window.innerWidth >= 768;
  isDarkMode = false;
  commandPaletteOpen = false;
  commandQuery = '';
  private tokenSub: Subscription | null = null;
  private readonly viewerMessage = 'Không gian làm việc này chỉ cho phép xem dữ liệu.';
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
    { path: '/profile', labelKey: 'Cá nhân', icon: 'person' },
  ];

  constructor(
    private auth: AuthService,
    private api: ApiService,
    private toastService: ToastService,
    private router: Router,
    public langService: LanguageService,
  ) {
    this.toasts$ = this.toastService.toasts$;
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
        this.workspaces = [];
        this.selectedWorkspaceId = null;
        this.router.navigateByUrl('/login');
        return;
      }

      this.api.getUserSettings().subscribe({
        next: (settings) => {
          this.amountDisplayMode = settings.amountDisplayMode;
          localStorage.setItem('finora.amountDisplayMode', settings.amountDisplayMode);
        },
      });

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

  closeSidebarOnMobile() {
    if (typeof window !== 'undefined' && window.innerWidth < 768) {
      this.sidebarOpen = false;
    }
  }

  @HostListener('window:resize')
  onViewportResize() {
    // A viewport change (rotation, split-screen, desktop resize) must not leave
    // an invisible mobile drawer with its backdrop still intercepting touch.
    this.sidebarOpen = window.innerWidth >= 768;
  }

  @HostListener('window:keydown.escape')
  onEscapeKey() {
    if (this.commandPaletteOpen) {
      this.closeCommandPalette();
      return;
    }
    if (typeof window !== 'undefined' && shouldCloseSidebarOnEscape(this.sidebarOpen, window.innerWidth)) {
      this.sidebarOpen = false;
    }
  }

  @HostListener('window:keydown.control.k', ['$event'])
  @HostListener('window:keydown.meta.k', ['$event'])
  onCommandShortcut(event: Event) {
    if (!this.isAuthenticated) return;
    event.preventDefault();
    this.openCommandPalette();
  }

  get commandItems(): CommandItem[] {
    return [
      { path: '/transactions', labelKey: 'Ghi giao dịch mới', description: 'Thêm khoản thu, chi hoặc chuyển tiền', icon: 'add', queryParams: { entry: '1' } },
      ...this.navItems.map((item) => ({ ...item, description: this.commandDescription(item.path) })),
    ];
  }

  get filteredCommands(): CommandItem[] {
    const query = this.commandQuery.trim().toLocaleLowerCase('vi-VN');
    if (!query) return this.commandItems;
    return this.commandItems.filter((item) =>
      `${item.labelKey} ${item.description}`.toLocaleLowerCase('vi-VN').includes(query)
    );
  }

  openCommandPalette() {
    this.commandQuery = '';
    this.commandPaletteOpen = true;
  }

  closeCommandPalette() {
    this.commandPaletteOpen = false;
    this.commandQuery = '';
  }

  updateCommandQuery(event: Event) {
    this.commandQuery = (event.target as HTMLInputElement).value;
  }

  runCommand(item: CommandItem) {
    this.closeCommandPalette();
    this.router.navigate([item.path], { queryParams: item.queryParams });
    this.closeSidebarOnMobile();
  }

  changeLanguage(event: Event) {
    const lang = (event.target as HTMLSelectElement).value as SupportedLanguage;
    this.langService.setLanguage(lang);
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

  private commandDescription(path: string): string {
    const descriptions: Record<string, string> = {
      '/dashboard': 'Xem sức khỏe tài chính và hoạt động gần đây',
      '/accounts': 'Quản lý các tài khoản tiền và ngân hàng',
      '/transactions': 'Tìm kiếm và theo dõi giao dịch',
      '/loans': 'Theo dõi khoản vay và nghĩa vụ nợ',
      '/assets': 'Quản lý tài sản đầu tư',
      '/properties': 'Quản lý bất động sản',
      '/budgets': 'Lập ngân sách theo tháng',
      '/forecast': 'Dự báo tài chính',
      '/portfolios': 'Quản lý danh mục',
      '/audit-logs': 'Xem lịch sử hoạt động',
      '/sepay': 'Kết nối giao dịch ngân hàng',
      '/automation': 'Thiết lập quy tắc tự động',
      '/assistant': 'Nhận hỗ trợ từ trợ lý AI',
    };
    return descriptions[path] || '';
  }
}
