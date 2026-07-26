import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { ApiService } from '../../core/services/api.service';
import { AuditLog } from '../../shared/models';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-audit-logs',
  standalone: true,
  imports: [CommonModule, TranslatePipe, IconComponent],
  templateUrl: './audit-logs.component.html',
})
export class AuditLogsComponent implements OnInit {
  logs: AuditLog[] = [];
  loading = false;
  error = '';
  actionFilter = '';
  targetFilter = '';

  constructor(private api: ApiService) {}

  ngOnInit() {
    this.refresh();
  }

  refresh() {
    this.loading = true;
    this.error = '';
    this.api.listAuditLogs().subscribe({
      next: (items) => {
        this.logs = items;
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.error = 'Không thể tải nhật ký hoạt động.';
      },
    });
  }

  setActionFilter(value: string) {
    this.actionFilter = value.trim().toLowerCase();
  }

  setTargetFilter(value: string) {
    this.targetFilter = value.trim().toLowerCase();
  }

  onActionFilterChange(event: Event) {
    const raw = (event.target as HTMLSelectElement | null)?.value || '';
    this.setActionFilter(raw);
  }

  onTargetFilterChange(event: Event) {
    const raw = (event.target as HTMLSelectElement | null)?.value || '';
    this.setTargetFilter(raw);
  }

  get filteredLogs() {
    return this.logs.filter((item) => {
      if (this.actionFilter && item.action.toLowerCase() !== this.actionFilter) {
        return false;
      }
      if (this.targetFilter && item.targetType.toLowerCase() !== this.targetFilter) {
        return false;
      }
      return true;
    });
  }

  uniqueActions(): string[] {
    const set = new Set<string>();
    for (const item of this.logs) {
      if (item.action) {
        set.add(item.action);
      }
    }
    return Array.from(set).sort();
  }

  uniqueTargets(): string[] {
    const set = new Set<string>();
    for (const item of this.logs) {
      if (item.targetType) {
        set.add(item.targetType);
      }
    }
    return Array.from(set).sort();
  }

  formatAsDate(value: string) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString();
  }

  prettyJson(value: string | undefined) {
    if (!value) return '-';
    const trimmed = value.trim();
    if (!trimmed) return '-';
    try {
      return JSON.stringify(JSON.parse(trimmed), null, 2);
    } catch (_error) {
      return trimmed;
    }
  }
}
