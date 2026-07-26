import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { AuthService } from '../../core/services/auth.service';
import { Account, BankConnection, BankFeedTransaction, SePayConnectResponse } from '../../shared/models';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { IconComponent } from '../../shared/icons/icon.component';

type InboxTab = 'pending_review' | 'auto_ready' | 'posted' | 'ignored' | 'all';

@Component({
  selector: 'app-sepay',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, TranslatePipe, IconComponent],
  templateUrl: './sepay.component.html',
})
export class SePayComponent implements OnInit {
  readonly postingTabs: { key: InboxTab; label: string }[] = [
    { key: 'pending_review', label: 'Cần xem' },
    { key: 'auto_ready', label: 'Đã tự ghi (đợi review)' },
    { key: 'posted', label: 'Đã match' },
    { key: 'ignored', label: 'Bỏ qua' },
    { key: 'all', label: 'Tất cả' },
  ];

  connections: BankConnection[] = [];
  accounts: Account[] = [];
  lastConnect?: SePayConnectResponse;
  feed: BankFeedTransaction[] = [];
  selectedFeed?: BankFeedTransaction;
  activeTab: InboxTab = 'pending_review';
  accountIdFilter = '';
  previewResult: unknown = null;
  isBusy = false;
  isActionBusy = false;

  reclassifyForm: FormGroup;

  constructor(
    private api: ApiService,
    private fb: FormBuilder,
    public auth: AuthService,
  ) {
    this.reclassifyForm = this.fb.group({
      type: ['income', Validators.required],
      accountId: [''],
      categoryId: [''],
      reason: ['Phan loai lai tu ngan hang', Validators.required],
    });
  }

  get hasConnected() {
    return this.connections.length > 0;
  }

  get selectedFeedId() {
    return this.selectedFeed?.id || '';
  }

  get activeFeed(): BankFeedTransaction | undefined {
    return this.selectedFeed ? this.feed.find((item) => item.id === this.selectedFeed?.id) : undefined;
  }

  get filteredFeed() {
    if (!this.accountIdFilter) return this.feed;
    return this.feed.filter((item) => item.accountId === this.accountIdFilter);
  }

  get pendingCount() {
    return this.feed.filter((item) => item.postingState === 'pending_review').length;
  }

  get autoReadyCount() {
    return this.feed.filter((item) => item.postingState === 'auto_ready').length;
  }

  get postedCount() {
    return this.feed.filter((item) => item.postingState === 'posted').length;
  }

  get ignoredCount() {
    return this.feed.filter((item) => item.postingState === 'ignored').length;
  }

  get canMutate() {
    return this.auth.canMutate;
  }

  ngOnInit() {
    this.refresh();
    this.api.getAccounts().subscribe((items) => {
      this.accounts = items;
    });
  }

  refresh() {
    const selectedId = this.selectedFeed?.id;
    this.api.listBankConnections().subscribe((items) => {
      this.connections = items;
    });
    const state = this.activeTab === 'all' ? undefined : this.activeTab;
    this.api.listBankFeedTransactions({ state, accountId: this.accountIdFilter || undefined }).subscribe((items) => {
      this.feed = items;
      this.selectedFeed = selectedId ? this.feed.find((item) => item.id === selectedId) : this.selectedFeed;
    });
  }

  setTab(tab: InboxTab) {
    if (this.activeTab === tab) return;
    this.activeTab = tab;
    this.refresh();
  }

  setAccountFilter(accountId: string) {
    this.accountIdFilter = accountId;
    this.refresh();
  }

  clearAccountFilter() {
    this.accountIdFilter = '';
    this.refresh();
  }

  connect() {
    if (!this.canMutate || this.isBusy) return;
    this.isBusy = true;
    this.api.connectSePay().subscribe({
      next: (response) => {
        this.lastConnect = response;
        this.refresh();
      },
      complete: () => {
        this.isBusy = false;
      },
      error: () => {
        this.isBusy = false;
      },
    });
  }

  sync(connectionId: string) {
    if (!this.canMutate || this.isBusy) return;
    this.isBusy = true;
    this.api.syncBankConnection(connectionId).subscribe({
      next: () => this.refresh(),
      complete: () => {
        this.isBusy = false;
      },
      error: () => {
        this.isBusy = false;
      },
    });
  }

  revoke(connectionId: string) {
    if (!this.canMutate) return;
    this.api.revokeBankConnection(connectionId).subscribe({
      next: () => this.refresh(),
    });
  }

  openDetails(item: BankFeedTransaction) {
    this.selectedFeed = item;
  }

  closeDetails() {
    this.selectedFeed = undefined;
  }

  syncLabel(connection: BankConnection) {
    return connection.lastSyncRequestedAt ? new Date(connection.lastSyncRequestedAt).toLocaleString() : 'chưa đồng bộ';
  }

  callbackLabel(connection: BankConnection) {
    return connection.lastSyncedAt ? new Date(connection.lastSyncedAt).toLocaleString() : 'chưa nhận';
  }

  statusText(item: BankFeedTransaction) {
    if (item.postingState === 'auto_ready') return 'Sẵn sàng tự ghi';
    if (item.postingState === 'posted') return 'Đã match';
    if (item.postingState === 'ignored') return 'Bỏ qua';
    return 'Cần xem';
  }

  statusClass(item: BankFeedTransaction) {
    if (item.postingState === 'auto_ready') return 'bg-emerald-100 text-emerald-800';
    if (item.postingState === 'posted') return 'bg-sky-100 text-sky-800';
    if (item.postingState === 'ignored') return 'bg-rose-100 text-rose-800';
    return 'bg-amber-100 text-amber-800';
  }

  canApprove(item: BankFeedTransaction) {
    return item.postingState === 'pending_review' || item.postingState === 'auto_ready';
  }

  approveLabel(item: BankFeedTransaction) {
    return item.postingState === 'auto_ready' ? 'Đồng ý tự ghi' : 'Đánh dấu đã ghi';
  }

  actionHint(item: BankFeedTransaction) {
    if (item.postingState === 'posted') return 'Đã tự tạo giao dịch';
    if (item.autoClassified && item.postingState === 'pending_review') return 'Đề xuất tự ghi cao';
    if (item.classificationConfidence && item.classificationConfidence >= 70) return 'Nên xem xét & phê duyệt';
    if (item.classificationConfidence) return `Độ tin cậy ${item.classificationConfidence.toFixed(0)}%`;
    return 'Cần xác nhận';
  }

  isTransferLike(item: BankFeedTransaction) {
    return (item.classificationEvidence || '').includes('transfer');
  }

  sourceLabel(item: BankFeedTransaction) {
    const source = ((item as { source?: string }).source) || item.classificationEvidence;
    return source ? `Nguồn: ${source}` : '';
  }

  accountName(accountId: string) {
    const found = this.accounts.find((item) => item.id === accountId);
    return found ? found.name : accountId;
  }

  confidence(item: BankFeedTransaction) {
    const value = item.classificationConfidence;
    if (value === undefined || value === null) return 'chưa xác định';
    return `${value.toFixed(0)}%`;
  }

  approve(item: BankFeedTransaction) {
    if (!this.canMutate || this.isActionBusy) return;
    this.isActionBusy = true;
    this.api.approveBankFeed(item.id).subscribe({
      next: () => {
        this.refresh();
        if (this.selectedFeed?.id === item.id) {
          this.selectedFeed = this.feed.find((row) => row.id === item.id);
        }
      },
      complete: () => {
        this.isActionBusy = false;
      },
      error: () => {
        this.isActionBusy = false;
      },
    });
  }

  ignore(item: BankFeedTransaction) {
    if (!this.canMutate || this.isActionBusy) return;
    this.isActionBusy = true;
    this.api.ignoreBankFeed(item.id).subscribe({
      next: () => this.refresh(),
      complete: () => {
        this.isActionBusy = false;
      },
      error: () => {
        this.isActionBusy = false;
      },
    });
  }

  openReclassify(item: BankFeedTransaction) {
    if (!this.canMutate) return;
    this.selectedFeed = item;
    this.reclassifyForm.patchValue({
      type: item.direction === 'out' ? 'expense' : 'income',
      accountId: item.accountId || '',
      categoryId: '',
      reason: `Phe duyet thu cong: ${item.direction === 'out' ? 'chi' : 'thu'}`,
    });
  }

  closeReclassify() {
    this.selectedFeed = undefined;
    this.reclassifyForm.reset({
      type: 'income',
      accountId: '',
      categoryId: '',
      reason: 'Phan loai lai tu ngan hang',
    });
    this.previewResult = null;
  }

  submitReclassify() {
    if (!this.canMutate || !this.selectedFeed || this.reclassifyForm.invalid) return;
    this.isActionBusy = true;
    this.api
      .reclassifyBankFeed(this.selectedFeed.id, {
        type: this.reclassifyForm.value.type,
        accountId: this.reclassifyForm.value.accountId || '',
        categoryId: this.reclassifyForm.value.categoryId || '',
        reason: this.reclassifyForm.value.reason || '',
      })
      .subscribe({
        next: () => {
          this.closeReclassify();
          this.refresh();
        },
        complete: () => {
          this.isActionBusy = false;
        },
        error: () => {
          this.isActionBusy = false;
        },
      });
  }

  preview() {
    this.api.previewRule({ sample: this.feed.slice(0, 6) }).subscribe((result) => {
      this.previewResult = result;
    });
  }

  trackByFeedId(_index: number, item: BankFeedTransaction) {
    return item.id;
  }
}
