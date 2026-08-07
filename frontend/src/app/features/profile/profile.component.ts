import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { AuthService } from '../../core/services/auth.service';
import { MySePaySummary } from '../../shared/models';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, RouterLink, IconComponent],
  templateUrl: './profile.component.html',
})
export class ProfileComponent implements OnInit {
  sepay: MySePaySummary | null = null;
  loading = true;
  error = '';
  notice = '';

  constructor(private api: ApiService, private auth: AuthService, private router: Router) {}

  ngOnInit() {
    this.refresh();
  }

  get linkedAccounts() {
    return this.sepay?.bankAccounts || [];
  }

  get mappedCount() {
    return this.linkedAccounts.filter((account) => !!account.mapping?.accountId).length;
  }

  refresh() {
    this.loading = true;
    this.error = '';
    this.api.getMySePaySummary().subscribe({
      next: (summary) => {
        this.sepay = summary;
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.error = 'Chưa thể tải trạng thái liên kết ngân hàng.';
      },
    });
  }

  relativeSync() {
    const raw = this.sepay?.profile?.lastSyncedAt;
    if (!raw) return 'Chưa đồng bộ';
    const date = new Date(raw);
    if (Number.isNaN(date.getTime())) return 'Chưa đồng bộ';
    const minutes = Math.floor((Date.now() - date.getTime()) / 60000);
    if (minutes < 1) return 'Đồng bộ vừa xong';
    if (minutes < 60) return `Đồng bộ ${minutes} phút trước`;
    if (minutes < 1440) return `Đồng bộ ${Math.floor(minutes / 60)} giờ trước`;
    return `Đồng bộ ${Math.floor(minutes / 1440)} ngày trước`;
  }

  logout() {
    this.auth.clearToken();
    this.router.navigateByUrl('/login');
  }
}
