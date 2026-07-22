import { Component, OnInit } from '@angular/core';
import { CommonModule, NgForOf, NgIf } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { AuthService } from '../../core/services/auth.service';
import { Portfolio, NetWorthSummary, PortfolioSnapshotPage } from '../../shared/models';

@Component({
  selector: 'app-portfolio-list',
  standalone: true,
  imports: [CommonModule, NgForOf, NgIf, ReactiveFormsModule],
  templateUrl: './portfolio-list.component.html',
})
export class PortfolioListComponent implements OnInit {
  form!: FormGroup;
  portfolios: Portfolio[] = [];
  selectedPortfolioId = '';
  snapshots: NetWorthSummary[] = [];
  nextCursor = '';
  snapshotLoading = false;
  createInProgress = false;
  error = '';
  baseCurrencies = ['VND', 'USD', 'EUR'];

  constructor(
    private fb: FormBuilder,
    private api: ApiService,
    public auth: AuthService,
  ) {
    this.form = this.fb.group({
      name: ['', Validators.required],
      baseCurrency: ['VND', Validators.required],
    });
  }

  ngOnInit() {
    this.reloadPortfolios();
  }

  reloadPortfolios() {
    this.api.getPortfolios().subscribe({
      next: (items) => {
        this.portfolios = items;
        if (!this.selectedPortfolioId && items.length > 0) {
          this.selectPortfolio(items[0].id);
        } else if (this.selectedPortfolioId && !items.some((p) => p.id === this.selectedPortfolioId)) {
          this.selectPortfolio(items[0]?.id || '');
        } else if (!this.selectedPortfolioId) {
          this.snapshots = [];
          this.nextCursor = '';
        }
      },
      error: () => {
        this.error = 'Không thể tải danh mục.';
      },
    });
  }

  selectPortfolio(portfolioId: string) {
    this.selectedPortfolioId = portfolioId;
    if (!portfolioId) {
      this.snapshots = [];
      this.nextCursor = '';
      return;
    }
    this.loadSnapshots(portfolioId, undefined, true);
  }

  createPortfolio() {
    if (!this.auth.canMutate || this.createInProgress || this.form.invalid) return;
    this.createInProgress = true;
    this.error = '';
    const payload = {
      name: this.form.value.name || '',
      baseCurrency: this.form.value.baseCurrency || 'VND',
    };
    this.api.createPortfolio(payload).subscribe({
      next: () => {
        this.form.reset({ name: '', baseCurrency: 'VND' });
        this.reloadPortfolios();
        this.createInProgress = false;
      },
      error: () => {
        this.error = 'Không thể tạo danh mục.';
        this.createInProgress = false;
      },
    });
  }

  loadSnapshots(portfolioId: string, cursor = '', replace = true) {
    if (!portfolioId) return;
    this.snapshotLoading = true;
    this.api.getPortfolioSnapshots(portfolioId, 8, cursor)
      .subscribe({
        next: (page: PortfolioSnapshotPage) => {
          if (replace) {
            this.snapshots = page.items;
          } else {
            this.snapshots = [...this.snapshots, ...page.items];
          }
          this.nextCursor = page.nextCursor || '';
          this.snapshotLoading = false;
        },
        error: () => {
          this.snapshotLoading = false;
          this.error = 'Không thể tải lịch sử tài sản.';
        },
      });
  }

  loadMoreSnapshots() {
    if (!this.nextCursor || this.snapshotLoading || !this.selectedPortfolioId) return;
    this.snapshotLoading = true;
    this.api.getPortfolioSnapshots(this.selectedPortfolioId, 8, this.nextCursor)
      .subscribe({
        next: (page: PortfolioSnapshotPage) => {
          this.snapshots = [...this.snapshots, ...page.items];
          this.nextCursor = page.nextCursor || '';
          this.snapshotLoading = false;
        },
        error: () => {
          this.snapshotLoading = false;
          this.error = 'Không thể tải tiếp lịch sử tài sản.';
        },
      });
  }

  latestNetWorth(): NetWorthSummary | undefined {
    return this.snapshots[0];
  }

  getSelectedPortfolioName(): string {
    const found = this.portfolios.find((item) => item.id === this.selectedPortfolioId);
    return found ? found.name : 'Selected Portfolio';
  }

  formatChange(value?: string) {
    if (!value) return '0.00';
    const amount = Number.parseFloat(value);
    if (Number.isNaN(amount)) return value;
    return amount.toFixed(2);
  }
}
