import { CommonModule } from '@angular/common';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { debounceTime, distinctUntilChanged, Subject, takeUntil } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Account, Transaction, TransactionListPage } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';
import { IconComponent } from '../../shared/icons/icon.component';
import { normalizeVndAmount } from '../../shared/money-input';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';

type TransactionFilters = {
  accountId: string;
  type: string;
  status: string;
  categoryId: string;
  search: string;
  from: string;
  to: string;
  limit: number;
};

@Component({
  selector: 'app-transaction-list',
  standalone: true,
  imports: [ReactiveFormsModule, CommonModule, IconComponent, TranslatePipe, RouterLink],
  templateUrl: './transaction-list.component.html'
})
export class TransactionListComponent implements OnInit, OnDestroy {
  form!: FormGroup;
  filterForm!: FormGroup;
  items: Transaction[] = [];
  accounts: Account[] = [];
  accountsLoading = true;
  mode: 'transaction' | 'transfer' = 'transaction';
  statusMessage = '';
  nextCursor = '';
  loading = false;
  pageLimit = 25;
  totalFiltered = 0;
  submitInFlight = false;
  entryOpen = false;
  selectedTransactionIds = new Set<string>();
  accountNameById = new Map<string, string>();

  get totalIncome(): number {
    return this.items
      .filter((x) => x.type === 'income')
      .reduce((sum, x) => sum + (Number.parseFloat(String(x.amount)) || 0), 0);
  }

  get totalExpense(): number {
    return this.items
      .filter((x) => x.type === 'expense')
      .reduce((sum, x) => sum + (Number.parseFloat(String(x.amount)) || 0), 0);
  }

  get netCashFlow(): number {
    return this.totalIncome - this.totalExpense;
  }

  get currencies(): string[] {
    return Array.from(new Set(this.items.map((item) => item.currency || 'VND')));
  }

  get hasMultipleCurrencies(): boolean {
    return this.currencies.length > 1;
  }

  get incomeTotalText(): string {
    return this.totalTextFor('income', '+');
  }

  get expenseTotalText(): string {
    return this.totalTextFor('expense', '-');
  }

  get netCashFlowText(): string {
    if (this.hasMultipleCurrencies) return 'Đa tiền tệ';
    const currency = this.currencies[0] || 'VND';
    return `${this.netCashFlow >= 0 ? '+' : ''}${this.netCashFlow.toLocaleString('vi-VN', { maximumFractionDigits: 0 })} ${currency}`;
  }

  private readonly destroy$ = new Subject<void>();

  constructor(
    private fb: FormBuilder,
    private api: ApiService,
    public auth: AuthService,
    private route: ActivatedRoute,
    private router: Router,
  ) {
    this.form = this.fb.group({
      mode: ['transaction'],
      accountId: [''],
      portfolioId: [''],
      type: ['expense', Validators.required],
      status: ['posted'],
      categoryId: [''],
      amount: ['', Validators.required],
      currency: ['VND', Validators.required],
      occurredAt: [''],
      note: [''],
      fromAccountId: [''],
      toAccountId: [''],
      fromAmount: [''],
    });

    this.filterForm = this.fb.group({
      accountId: [''],
      type: [''],
      status: [''],
      categoryId: [''],
      search: [''],
      from: [''],
      to: [''],
      limit: [this.pageLimit],
    });
  }

  ngOnInit() {
    this.api.getAccounts().subscribe({
      next: (items) => {
        this.accounts = items;
        this.accountNameById = new Map(items.map((item) => [item.id, item.name]));

        if (!this.filterForm.value.accountId && items.length) {
          const defaultAccountId = items[0]?.id || '';
          this.form.patchValue({ accountId: defaultAccountId, fromAccountId: defaultAccountId });
        }
        this.accountsLoading = false;
      },
      error: () => {
        this.accounts = [];
        this.accountsLoading = false;
      },
    });

    this.route.queryParamMap.pipe(takeUntil(this.destroy$)).subscribe((params) => {
      this.entryOpen = params.get('entry') === '1';
      const state = this.readFiltersFromQuery(params);
      this.filterForm.patchValue(state, { emitEvent: false });
      this.reload();
    });

    this.filterForm
      .get('search')
      ?.valueChanges.pipe(debounceTime(300), distinctUntilChanged(), takeUntil(this.destroy$))
      .subscribe(() => {
        this.applyFilters();
      });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private readFiltersFromQuery(params: { get: (name: string) => string | null }): TransactionFilters {
    const rawLimit = params.get('limit') ?? '';
    const parsedLimit = Number.parseInt(rawLimit, 10);
    const limit = Number.isFinite(parsedLimit) && parsedLimit > 0 ? parsedLimit : this.pageLimit;

    return {
      accountId: params.get('accountId')?.trim() ?? '',
      type: params.get('type')?.trim() ?? '',
      status: params.get('status')?.trim() ?? '',
      categoryId: params.get('categoryId')?.trim() ?? '',
      search: params.get('search')?.trim() ?? '',
      from: params.get('from')?.trim() ?? '',
      to: params.get('to')?.trim() ?? '',
      limit,
    };
  }

  private syncQueryWithFilters(payload: ReturnType<TransactionListComponent['buildListPayload']>) {
    const queryParams: Record<string, string | null> = {
      accountId: payload.accountId || null,
      type: payload.type || null,
      status: payload.status || null,
      categoryId: payload.categoryId || null,
      search: payload.search || null,
      from: payload.from || null,
      to: payload.to || null,
      limit: payload.limit ? String(payload.limit) : null,
    };

    this.router.navigate([], {
      relativeTo: this.route,
      queryParams,
      replaceUrl: true,
    });
  }

  reload() {
    if (!this.filterForm.value.accountId && this.accounts[0]?.id) {
      this.filterForm.patchValue({ accountId: this.accounts[0].id }, { emitEvent: false });
    }

    const payload = this.buildListPayload();
    this.selectedTransactionIds.clear();
    this.loading = true;
    this.nextCursor = '';
    this.totalFiltered = 0;
    this.api.getTransactions(payload).subscribe({
      next: (page: TransactionListPage) => {
        this.items = page.items ?? [];
        this.nextCursor = page.nextCursor || '';
        this.totalFiltered = this.items.length + (this.nextCursor ? 1 : 0);
        this.loading = false;
      },
      error: () => {
        this.statusMessage = 'Không thể tải giao dịch.';
        this.loading = false;
      },
    });
  }

  loadMore() {
    if (!this.nextCursor) return;
    const payload = this.buildListPayload();
    this.loading = true;
    this.api.getTransactions({ ...payload, cursor: this.nextCursor }).subscribe({
      next: (page: TransactionListPage) => {
        const nextItems = page.items ?? [];
        const existingIds = new Set(this.items.map((item) => item.id));
        for (const item of nextItems) {
          if (!existingIds.has(item.id)) {
            this.items.push(item);
          }
        }
        this.nextCursor = page.nextCursor || '';
        this.totalFiltered = this.items.length + (this.nextCursor ? 1 : 0);
        this.loading = false;
      },
      error: () => {
        this.statusMessage = 'Không thể tải thêm giao dịch.';
        this.loading = false;
      },
    });
  }

  applyFilters() {
    const payload = this.buildListPayload();
    this.syncQueryWithFilters(payload);
  }

  clearFilters() {
    this.filterForm.reset({
      accountId: '',
      type: '',
      status: '',
      categoryId: '',
      search: '',
      from: '',
      to: '',
      limit: this.pageLimit,
    });
    this.applyFilters();
  }

  submit() {
    if (!this.auth.canMutate || this.submitInFlight) return;
    const mode = this.form.value.mode === 'transfer' ? 'transfer' : 'transaction';
    this.mode = mode;
    if (this.form.invalid) return;

    this.submitInFlight = true;
    if (mode === 'transfer') {
      const payload = {
        fromAccountId: this.form.value.fromAccountId || '',
        toAccountId: this.form.value.toAccountId || '',
        amount: normalizeVndAmount(this.form.value.amount),
        currency: this.form.value.currency || 'VND',
        note: this.form.value.note || '',
        occurredAt: this.form.value.occurredAt || undefined,
      };
      this.api.createTransfer(payload).subscribe({
      next: () => {
          this.statusMessage = 'Đã tạo yêu cầu chuyển tiền.';
          this.resetForm();
          this.entryOpen = false;
          this.applyFilters();
          this.submitInFlight = false;
        },
        error: () => {
          this.statusMessage = 'Không thể tạo yêu cầu chuyển tiền.';
          this.submitInFlight = false;
        },
      });
      return;
    }

    const payload = {
      accountId: this.form.value.accountId || '',
      portfolioId: this.form.value.portfolioId || '',
      categoryId: this.form.value.categoryId || '',
      type: this.form.value.type || 'expense',
      amount: normalizeVndAmount(this.form.value.amount),
      currency: this.form.value.currency || 'VND',
      status: this.form.value.status || 'posted',
      note: this.form.value.note || '',
      occurredAt: this.form.value.occurredAt || undefined,
    };

    this.api.createTransaction(payload).subscribe({
      next: () => {
        this.statusMessage = 'Đã ghi nhận giao dịch.';
        this.resetForm();
        this.entryOpen = false;
        this.applyFilters();
        this.submitInFlight = false;
      },
      error: () => {
        this.statusMessage = 'Không thể ghi nhận giao dịch.';
        this.submitInFlight = false;
      },
    });
  }

  setMode(mode: 'transaction' | 'transfer') {
    this.mode = mode;
    this.form.patchValue({ mode });
  }

  openEntry(mode: 'transaction' | 'transfer' = 'transaction') {
    if (!this.auth.canMutate) return;
    if (!this.accounts.length) {
      this.statusMessage = 'Hãy thêm ít nhất một tài khoản trước khi ghi giao dịch.';
      return;
    }
    this.entryOpen = true;
    this.setMode(mode);
    this.statusMessage = '';
  }

  closeEntry() {
    if (!this.submitInFlight) {
      this.entryOpen = false;
    }
  }

  isSelected(txId: string) {
    return this.selectedTransactionIds.has(txId);
  }

  toggleSelection(txId: string) {
    if (this.selectedTransactionIds.has(txId)) {
      this.selectedTransactionIds.delete(txId);
      return;
    }
    this.selectedTransactionIds.add(txId);
  }

  get selectedCount() {
    return this.selectedTransactionIds.size;
  }

  get selectedAmountText() {
    const totals = new Map<string, number>();

    for (const item of this.items) {
      if (!this.selectedTransactionIds.has(item.id)) continue;
      const value = Number.parseFloat(item.amount);
      if (!Number.isFinite(value)) continue;
      const currency = item.currency || 'VND';
      totals.set(currency, (totals.get(currency) || 0) + value);
    }

    if (!totals.size) {
      return '0';
    }

    return Array.from(totals.entries())
      .map(([currency, value]) => `${value.toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 2 })} ${currency}`)
      .join(' + ');
  }

  statusClass(status: string | undefined) {
    switch ((status || 'posted').toLowerCase()) {
      case 'posted':
        return 'bg-emerald-100 text-emerald-800 border-emerald-200';
      case 'pending':
        return 'bg-amber-100 text-amber-900 border-amber-300';
      case 'reconciled':
        return 'bg-sky-100 text-sky-800 border-sky-200';
      case 'voided':
        return 'bg-rose-100 text-rose-800 border-rose-200';
      default:
        return 'bg-slate-100 text-slate-700 border-slate-200';
    }
  }

  accountLabel(accountId: string) {
    return this.accountNameById.get(accountId) || accountId;
  }

  private resetForm() {
    const accountFallback = this.accounts[0]?.id || '';
    this.form.patchValue({
      accountId: accountFallback,
      portfolioId: '',
      categoryId: '',
      type: 'expense',
      status: 'posted',
      amount: '',
      note: '',
      occurredAt: '',
      fromAccountId: accountFallback,
      toAccountId: '',
      fromAmount: '',
    });
  }

  private buildListPayload() {
    const accountId = this.filterForm.value.accountId || '';
    const limitValue = this.filterForm.value.limit;
    const parsedLimit = Number.parseInt(limitValue, 10);
    const limit = Number.isFinite(parsedLimit) && parsedLimit > 0 ? parsedLimit : this.pageLimit;
    this.pageLimit = limit;
    return {
      accountId: accountId || undefined,
      type: this.filterForm.value.type || undefined,
      status: this.filterForm.value.status || undefined,
      categoryId: this.filterForm.value.categoryId || undefined,
      search: this.filterForm.value.search || undefined,
      from: this.filterForm.value.from || undefined,
      to: this.filterForm.value.to || undefined,
      limit,
    };
  }

  private totalTextFor(type: 'income' | 'expense', sign: '+' | '-'): string {
    const totals = new Map<string, number>();
    for (const item of this.items) {
      if (item.type !== type) continue;
      const amount = Number.parseFloat(item.amount);
      if (!Number.isFinite(amount)) continue;
      const currency = item.currency || 'VND';
      totals.set(currency, (totals.get(currency) || 0) + amount);
    }
    if (!totals.size) return `0 ${this.currencies[0] || 'VND'}`;
    return Array.from(totals.entries())
      .map(([currency, amount]) => `${sign}${amount.toLocaleString('vi-VN', { maximumFractionDigits: 0 })} ${currency}`)
      .join(' · ');
  }

  typeLabel(type: string | undefined) {
    const labels: Record<string, string> = {
      income: 'Thu nhập',
      expense: 'Chi phí',
      transfer: 'Chuyển khoản',
      valuation_adjustment: 'Điều chỉnh định giá',
    };
    return labels[type || ''] || type || 'Khác';
  }

  statusLabel(status: string | undefined) {
    const labels: Record<string, string> = {
      posted: 'Đã ghi nhận',
      pending: 'Đang chờ',
      voided: 'Đã huỷ',
      rejected: 'Đã từ chối',
    };
    return labels[status || ''] || status || 'Chưa xác định';
  }
}
