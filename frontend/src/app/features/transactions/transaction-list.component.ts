import { CommonModule } from '@angular/common';
import { Component, OnDestroy, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { debounceTime, distinctUntilChanged, Subject, takeUntil } from 'rxjs';
import { ApiService } from '../../core/services/api.service';
import { Account, Transaction, TransactionListPage } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';

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
  imports: [ReactiveFormsModule, CommonModule],
  templateUrl: './transaction-list.component.html'
})
export class TransactionListComponent implements OnInit, OnDestroy {
  form!: FormGroup;
  filterForm!: FormGroup;
  items: Transaction[] = [];
  accounts: Account[] = [];
  mode: 'transaction' | 'transfer' = 'transaction';
  statusMessage = '';
  nextCursor = '';
  loading = false;
  pageLimit = 25;
  totalFiltered = 0;
  submitInFlight = false;
  selectedTransactionIds = new Set<string>();
  accountNameById = new Map<string, string>();

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
    this.api.getAccounts().subscribe((items) => {
      this.accounts = items;
      this.accountNameById = new Map(items.map((item) => [item.id, item.name]));

      if (!this.filterForm.value.accountId && items.length) {
        const defaultAccountId = items[0]?.id || '';
        this.form.patchValue({ accountId: defaultAccountId, fromAccountId: defaultAccountId });
      }
    });

    this.route.queryParamMap.pipe(takeUntil(this.destroy$)).subscribe((params) => {
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
        this.statusMessage = 'Unable to load transactions.';
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
        this.statusMessage = 'Unable to load next page.';
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
        amount: this.form.value.amount || '',
        currency: this.form.value.currency || 'VND',
        note: this.form.value.note || '',
        occurredAt: this.form.value.occurredAt || undefined,
      };
      this.api.createTransfer(payload).subscribe({
        next: () => {
          this.statusMessage = 'Transfer request created.';
          this.resetForm();
          this.applyFilters();
          this.submitInFlight = false;
        },
        error: () => {
          this.statusMessage = 'Unable to create transfer.';
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
      amount: this.form.value.amount || '',
      currency: this.form.value.currency || 'VND',
      status: this.form.value.status || 'posted',
      note: this.form.value.note || '',
      occurredAt: this.form.value.occurredAt || undefined,
    };

    this.api.createTransaction(payload).subscribe({
      next: () => {
        this.statusMessage = 'Transaction added.';
        this.resetForm();
        this.applyFilters();
        this.submitInFlight = false;
      },
      error: () => {
        this.statusMessage = 'Unable to add transaction.';
        this.submitInFlight = false;
      },
    });
  }

  setMode(mode: 'transaction' | 'transfer') {
    this.mode = mode;
    this.form.patchValue({ mode });
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
}

