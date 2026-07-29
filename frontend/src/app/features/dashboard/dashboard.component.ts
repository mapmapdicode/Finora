import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiService } from '../../core/services/api.service';
import { Account, NetWorthSummary, Portfolio, PortfolioSnapshotPage, Transaction } from '../../shared/models';

import { AuthService } from '../../core/services/auth.service';
import { RouterLink } from '@angular/router';
import { IconComponent } from '../../shared/icons/icon.component';

import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, IconComponent, RouterLink],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css',
})
export class DashboardComponent implements OnInit {
  loading = signal(true);
  hideBalance = signal(false);
  summary = signal('Chưa kết nối dữ liệu');
  dashboardError = '';

  toggleBalanceVisibility() {
    this.hideBalance.update((v) => !v);
  }
  summaryByPortfolio: string[] = [];
  accounts: Account[] = [];
  transactions: Transaction[] = [];
  portfolios: Portfolio[] = [];

  selectedPortfolioId = '';
  snapshots: NetWorthSummary[] = [];
  snapshotCursor = '';
  snapshotLoading = false;

  trendSvgPath = '';
  trendAreaPath = '';
  trendDotPositions: Array<{ x: number; y: number; value: number; asOf: string }> = [];
  trendMin = 0;
  trendMax = 0;
  readonly trendHeight = 150;
  readonly trendWidth = 400;
  readonly trendPadding = 15;

  asOfPreviewInput = '';
  asOfSnapshotAt = '';
  asOfError = '';

  netWorth: NetWorthSummary = {
    asOfAt: '',
    baseCurrency: 'VND',
    netWorth: '0.00',
    cash: '0.00',
    liabilities: '0.00',
    netWorthChange: '0.00',
    snapshotVersion: 0,
    assets: {
      cash: '0.00',
      receivables: '0.00',
      property: '0.00',
      otherAssets: '0.00',
      accruedInterest: '0.00',
    },
    dataQuality: {
      reconciledAccounts: 0,
      staleValuations: 0,
      asOfSource: 'ledger',
    },
    attribution: {
      externalCashFlow: '0.00',
      accruedInterest: '0.00',
      valuationChange: '0.00',
      accruedFee: '0.00',
    },
  };
  selectedSnapshotNetWorth: NetWorthSummary | null = null;

  get assetBreakdown() {
    const assets = this.drillDownSnapshot().assets;
    if (!assets) return [];
    const entries = [
      { label: 'Tiền mặt', amount: assets.cash, color: '#5e2d91' },
      { label: 'Khoản phải thu', amount: assets.receivables, color: '#078847' },
      { label: 'Bất động sản', amount: assets.property, color: '#fc7728' },
      { label: 'Tài sản khác', amount: assets.otherAssets, color: '#8d0052' },
      { label: 'Lãi tích lũy', amount: assets.accruedInterest, color: '#5c6b85' },
    ].map((item) => ({ ...item, numericAmount: Number.parseFloat(item.amount) || 0 }));
    const total = entries.reduce((sum, item) => sum + Math.max(0, item.numericAmount), 0);
    if (!total) return [];
    return entries
      .filter((item) => item.numericAmount > 0)
      .map(({ numericAmount, ...item }) => ({ ...item, percentage: Math.round((numericAmount / total) * 100) }));
  }

  constructor(private api: ApiService, public auth: AuthService) {}

  ngOnInit() {
    this.refresh();
  }

  public refresh() {
    this.loading.set(true);
    this.dashboardError = '';
    forkJoin({
      portfolios: this.api.getPortfolios().pipe(catchError(() => of([]))),
      accounts: this.api.getAccounts().pipe(catchError(() => of([]))),
      transactionsPage: this.api.getTransactions({ limit: 5 }).pipe(catchError(() => of({ items: [] }))),
    }).subscribe({
      next: ({ portfolios, accounts, transactionsPage }) => {
        this.accounts = accounts;
        const map: Record<string, number> = {};
        for (const account of accounts) {
          const key = account.portfolioId || 'default';
          map[key] = (map[key] || 0) + 1;
        }
        this.summaryByPortfolio = Object.entries(map).map(
          ([portfolioId, count]) => `${portfolioId}: ${count} TK`
        );

        this.transactions = (transactionsPage.items || []).slice(0, 5);

        this.portfolios = portfolios;
        if (!portfolios.length) {
          this.loading.set(false);
          this.summary.set('Chưa có danh mục.');
          return;
        }

        const targetId =
          this.selectedPortfolioId && portfolios.some((item) => item.id === this.selectedPortfolioId)
            ? this.selectedPortfolioId
            : portfolios[0].id;

        this.selectPortfolio(targetId);
      },
      error: () => {
        this.loading.set(false);
        this.dashboardError = 'Không thể tải dữ liệu tổng quan. Kiểm tra kết nối rồi thử lại.';
        this.summary.set('Không lấy được dữ liệu tổng quan.');
      },
    });
  }

  onPortfolioChange(portfolioId: string) {
    this.selectPortfolio(portfolioId);
  }

  selectPortfolio(portfolioId: string) {
    if (!portfolioId) {
      return;
    }
    this.selectedPortfolioId = portfolioId;
    this.asOfSnapshotAt = '';
    this.asOfPreviewInput = '';
    this.asOfError = '';
    this.selectedSnapshotNetWorth = null;
    this.loadNetWorth(portfolioId);
    this.loadSnapshots(portfolioId, undefined, true);
  }

  loadNetWorth(portfolioId: string, asOfAt?: string) {
    this.loading.set(true);
    const shouldUseAsOf = !!(asOfAt && asOfAt.trim());
    const requestedMode = shouldUseAsOf ? 'asOf' : 'current';

    this.api.getNetWorth(portfolioId, asOfAt?.trim()).subscribe({
      next: (nw) => {
        this.netWorth = nw;
        this.loading.set(false);
        this.asOfError = '';
        this.dashboardError = '';
        this.selectedSnapshotNetWorth = shouldUseAsOf ? nw : null;
        this.summary.set(this.buildSummary(nw, requestedMode, asOfAt || nw.asOfAt));
      },
      error: () => {
        this.loading.set(false);
        this.netWorth = this.defaultEmptyNetWorth();
        this.selectedSnapshotNetWorth = null;
        this.dashboardError = shouldUseAsOf
          ? 'Không thể tải giá trị ròng tại thời điểm đã chọn.'
          : 'Không thể tải giá trị ròng hiện tại.';
        this.summary.set(
          shouldUseAsOf
            ? 'Không lấy được dữ liệu tài sản ròng cho thời điểm đã chọn.'
            : 'Không lấy được dữ liệu net worth.'
        );
      },
    });
  }

  loadSnapshots(portfolioId: string, cursor: string | undefined = undefined, replace = true) {
    if (!portfolioId) return;
    this.snapshotLoading = true;
    this.api.getPortfolioSnapshots(portfolioId, 12, cursor).subscribe({
      next: (page: PortfolioSnapshotPage) => {
        if (replace) {
          this.snapshots = page.items;
        } else {
          this.snapshots = [...this.snapshots, ...page.items];
        }
        this.snapshotCursor = page.nextCursor || '';
        this.buildTrendFromSnapshots();
        this.snapshotLoading = false;
      },
      error: () => {
        this.snapshotLoading = false;
        this.snapshotCursor = '';
        this.trendSvgPath = '';
        this.trendAreaPath = '';
      },
    });
  }

  loadMoreSnapshots() {
    if (!this.snapshotCursor || !this.selectedPortfolioId) return;
    this.loadSnapshots(this.selectedPortfolioId, this.snapshotCursor, false);
  }

  previewAsOfNow(event: Event) {
    event.preventDefault();
    const iso = this.toIsoDate(this.asOfPreviewInput);
    if (!iso) {
      this.asOfError = 'Vui lòng nhập thời điểm hợp lệ.';
      return;
    }
    if (!this.selectedPortfolioId) return;
    this.asOfSnapshotAt = '';
    this.asOfError = '';
    this.selectedSnapshotNetWorth = null;
    this.loadNetWorth(this.selectedPortfolioId, iso);
  }

  selectSnapshot(item: NetWorthSummary) {
    if (!this.selectedPortfolioId || !item?.asOfAt) return;
    this.asOfPreviewInput = item.asOfAt;
    this.asOfSnapshotAt = item.asOfAt;
    this.selectedSnapshotNetWorth = item;
    this.asOfError = '';
    this.loadNetWorth(this.selectedPortfolioId, item.asOfAt);
  }

  onTrendPointClick(pointAsOf: string, event: Event) {
    event.preventDefault();
    const candidate = this.snapshotByAsOf(pointAsOf);
    if (!candidate) return;
    this.selectSnapshot(candidate);
  }

  isActiveTrendPoint(pointAsOf: string) {
    return !!this.asOfSnapshotAt && this.asOfSnapshotAt === pointAsOf;
  }

  clearAsOfPreview() {
    this.asOfSnapshotAt = '';
    this.asOfPreviewInput = '';
    this.asOfError = '';
    this.selectedSnapshotNetWorth = null;
    if (!this.selectedPortfolioId) return;
    this.loadNetWorth(this.selectedPortfolioId);
  }

  isSelectedSnapshot(item: NetWorthSummary) {
    return !!this.asOfSnapshotAt && this.asOfSnapshotAt === item.asOfAt;
  }

  drillDownSnapshot(): NetWorthSummary {
    return this.selectedSnapshotNetWorth || this.netWorth;
  }

  drillDownMode() {
    return this.selectedSnapshotNetWorth ? 'snapshot' : 'current';
  }

  snapshotByAsOf(asOfAt: string): NetWorthSummary | null {
    return this.snapshots.find((item) => item.asOfAt === asOfAt) || null;
  }

  drillDownAsOfText(): string {
    const data = this.drillDownSnapshot();
    if (!data.asOfAt) return '';
    return new Date(data.asOfAt).toLocaleString();
  }

  formatAmount(value: string | number | undefined): string {
    const amount = Number.parseFloat(String(value ?? 0));
    return new Intl.NumberFormat('vi-VN', { maximumFractionDigits: 0 }).format(Number.isFinite(amount) ? amount : 0);
  }

  formatDate(value: string | undefined): string {
    if (!value) return '';
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? '' : new Intl.DateTimeFormat('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric' }).format(date);
  }

  transactionLabel(transaction: Transaction): string {
    return transaction.name || transaction.note || (transaction.type === 'income' ? 'Khoản thu' : transaction.type === 'transfer' ? 'Chuyển tiền' : 'Khoản chi');
  }

  private toIsoDate(raw: string): string {
    if (!raw || !raw.trim()) {
      return '';
    }
    const parsed = new Date(raw.trim());
    if (Number.isNaN(parsed.getTime())) return '';
    return parsed.toISOString();
  }

  private buildSummary(nw: NetWorthSummary, mode: 'current' | 'asOf', asOfAt: string) {
    const asOfText = asOfAt ? ` (thời điểm: ${new Date(asOfAt).toLocaleString()})` : '';
    const modeText = mode === 'asOf' ? 'Tài sản ròng theo thời điểm' : 'Tài sản ròng hiện tại';
    return `${modeText}: ${nw.netWorth} ${nw.baseCurrency}${asOfText}`;
  }

  private buildTrendFromSnapshots() {
    const ordered = [...this.snapshots].sort((a, b) =>
      new Date(a.asOfAt).getTime() - new Date(b.asOfAt).getTime()
    );
    const values = ordered
      .map((item) => parseFloat(item.netWorth))
      .filter((value) => Number.isFinite(value));
    if (values.length < 2) {
      this.trendSvgPath = '';
      this.trendAreaPath = '';
      this.trendDotPositions = [];
      return;
    }

    this.trendMin = Math.min(...values);
    this.trendMax = Math.max(...values);
    const range = this.trendMax === this.trendMin ? 1 : this.trendMax - this.trendMin;

    const usableHeight = this.trendHeight - this.trendPadding * 2;
    const usableWidth = this.trendWidth - this.trendPadding * 2;
    const widthStep = ordered.length > 1 ? usableWidth / (ordered.length - 1) : 0;

    const dots = ordered.map((item, index) => {
      const value = parseFloat(item.netWorth);
      const finiteValue = Number.isFinite(value) ? value : this.trendMin;
      const x = this.trendPadding + widthStep * index;
      const y = this.trendPadding + (this.trendMax - finiteValue) * (usableHeight / range);
      return {
        x,
        y,
        value: finiteValue,
        asOf: item.asOfAt,
      };
    });
    this.trendDotPositions = dots;

    const d = dots
      .map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x.toFixed(2)} ${point.y.toFixed(2)}`)
      .join(' ');
    this.trendSvgPath = d;
    const baseline = this.trendHeight - this.trendPadding;
    this.trendAreaPath = `${d} L ${dots[dots.length - 1].x.toFixed(2)} ${baseline} L ${dots[0].x.toFixed(2)} ${baseline} Z`;
  }

  private defaultEmptyNetWorth(): NetWorthSummary {
    return {
      asOfAt: '',
      baseCurrency: 'VND',
      netWorth: '0.00',
      cash: '0.00',
      liabilities: '0.00',
      netWorthChange: '0.00',
      snapshotVersion: 0,
      assets: {
        cash: '0.00',
        receivables: '0.00',
        property: '0.00',
        otherAssets: '0.00',
        accruedInterest: '0.00',
      },
      dataQuality: {
        reconciledAccounts: 0,
        staleValuations: 0,
        asOfSource: 'ledger',
      },
      attribution: {
        externalCashFlow: '0.00',
        accruedInterest: '0.00',
        valuationChange: '0.00',
        accruedFee: '0.00',
      },
    };
  }
}
