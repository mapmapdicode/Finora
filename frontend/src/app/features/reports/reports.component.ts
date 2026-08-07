import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { ApiService } from '../../core/services/api.service';
import { Transaction, TransactionListPage } from '../../shared/models';
import { VndMoneyPipe } from '../../shared/pipes/vnd-money.pipe';

type ReportPeriod = 'week' | 'month';
type CategoryTotal = { name: string; amount: number; percentage: number };

@Component({
  selector: 'app-reports',
  standalone: true,
  imports: [CommonModule, VndMoneyPipe],
  templateUrl: './reports.component.html',
})
export class ReportsComponent implements OnInit {
  period: ReportPeriod = 'week';
  transactions: Transaction[] = [];
  loading = true;
  error = '';
  rangeLabel = '';

  constructor(private api: ApiService) {}

  ngOnInit() { this.load(); }

  get income(): number { return this.totalFor('income'); }
  get expense(): number { return this.totalFor('expense'); }
  get balance(): number { return this.income - this.expense; }
  get savingsRate(): number { return this.income > 0 ? Math.round((this.balance / this.income) * 100) : 0; }

  get expenseCategories(): CategoryTotal[] {
    const totals = new Map<string, number>();
    for (const item of this.transactions.filter((transaction) => transaction.type === 'expense')) {
      const key = item.categoryId || item.name || 'Chưa phân loại';
      totals.set(key, (totals.get(key) || 0) + this.amountOf(item));
    }
    return [...totals.entries()]
      .map(([name, amount]) => ({ name, amount, percentage: this.expense ? Math.round((amount / this.expense) * 100) : 0 }))
      .sort((a, b) => b.amount - a.amount)
      .slice(0, 5);
  }

  setPeriod(period: ReportPeriod) {
    if (this.period === period) return;
    this.period = period;
    this.load();
  }

  load() {
    const { from, to } = this.dateRange();
    this.loading = true;
    this.error = '';
    this.loadTransactionPage({ from, to, limit: 100 }, []);
  }

  private totalFor(type: 'income' | 'expense') { return this.transactions.filter((item) => item.type === type).reduce((sum, item) => sum + this.amountOf(item), 0); }
  private amountOf(item: Transaction) { return Number.parseFloat(item.amount) || 0; }

  private loadTransactionPage(options: { from: string; to: string; limit: number; cursor?: string }, collected: Transaction[]) {
    this.api.getTransactions(options).subscribe({
      next: (page: TransactionListPage) => {
        const next = [...collected, ...(page.items || [])];
        // Fetch every page in the selected period (up to 1,000 entries) so a
        // monthly report does not silently omit an active user's transactions.
        if (page.nextCursor && next.length < 1_000) {
          this.loadTransactionPage({ ...options, cursor: page.nextCursor }, next);
          return;
        }
        this.transactions = next;
        this.loading = false;
      },
      error: () => { this.transactions = []; this.loading = false; this.error = 'Không thể tải báo cáo. Vui lòng thử lại.'; },
    });
  }

  private dateRange() {
    const now = new Date();
    const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59);
    const start = new Date(end);
    if (this.period === 'week') start.setDate(end.getDate() - 6);
    else start.setDate(1);
    const format = (date: Date) => date.toISOString().slice(0, 10);
    this.rangeLabel = `${start.toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit' })} – ${end.toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric' })}`;
    return { from: format(start), to: format(end) };
  }
}
