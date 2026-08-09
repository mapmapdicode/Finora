import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { Account, Loan, Transaction, TransactionListPage } from '../../shared/models';
import { VndMoneyPipe } from '../../shared/pipes/vnd-money.pipe';

type ReportPeriod = 'week' | 'month';
type Breakdown = { name: string; amount: number; percentage: number };
type DailyFlow = { day: string; label: string; in: number; out: number; net: number; width: number };

@Component({
  selector: 'app-reports', standalone: true, imports: [CommonModule, FormsModule, VndMoneyPipe],
  templateUrl: './reports.component.html',
})
export class ReportsComponent implements OnInit {
  period: ReportPeriod = 'month';
  selectedMonth = new Date().toISOString().slice(0, 7);
  allTransactions: Transaction[] = [];
  accounts: Account[] = [];
  loans: Loan[] = [];
  loading = true;
  error = '';
  rangeLabel = '';

  constructor(private api: ApiService) {}

  ngOnInit() {
    this.loadAccountsAndLoans();
    this.loadTransactionPage({ limit: 250 }, []);
  }

  get transactions(): Transaction[] {
    const inMonth = this.allTransactions.filter((item) => this.dateKey(item).startsWith(this.selectedMonth));
    if (this.period === 'month') return inMonth;
    const anchor = inMonth.map((item) => this.dateKey(item)).sort().at(-1);
    if (!anchor) return [];
    const end = new Date(`${anchor}T00:00:00`);
    const start = new Date(end); start.setDate(end.getDate() - 6);
    const startKey = this.isoDate(start);
    return inMonth.filter((item) => this.dateKey(item) >= startKey && this.dateKey(item) <= anchor);
  }

  get income(): number { return this.totalFor('income'); }
  get expense(): number { return this.totalFor('expense'); }
  get loanCollections(): number { return this.totalFor('loan_payment'); }
  get loanDisbursements(): number { return this.totalFor('loan_disbursement'); }
  get cashIn(): number { return this.income + this.loanCollections; }
  get cashOut(): number { return this.expense + this.loanDisbursements; }
  get cashChange(): number { return this.cashIn - this.cashOut; }
  get lifestyleBalance(): number { return this.income - this.expense; }
  get savingsRate(): number { return this.income > 0 ? Math.round((this.lifestyleBalance / this.income) * 100) : 0; }
  get activeLoanBalance(): number { return this.loans.filter((loan) => loan.direction === 'receivable' && loan.status !== 'closed').reduce((sum, loan) => sum + this.amount(loan.principalBalance || loan.principalInitial), 0); }
  get activeLoanCount(): number { return this.loans.filter((loan) => loan.direction === 'receivable' && loan.status !== 'closed').length; }

  get expenseCategories(): Breakdown[] { return this.breakdown('expense'); }
  get incomeSources(): Breakdown[] { return this.breakdown('income'); }

  get dailyFlows(): DailyFlow[] {
    const days = new Map<string, { in: number; out: number }>();
    for (const transaction of this.transactions) {
      const key = this.dateKey(transaction);
      const item = days.get(key) || { in: 0, out: 0 };
      if (transaction.type === 'income' || transaction.type === 'loan_payment') item.in += this.amountOf(transaction);
      if (transaction.type === 'expense' || transaction.type === 'loan_disbursement') item.out += this.amountOf(transaction);
      days.set(key, item);
    }
    const max = Math.max(1, ...[...days.values()].map((item) => Math.max(item.in, item.out)));
    return [...days.entries()].sort(([a], [b]) => a.localeCompare(b)).slice(-10).map(([day, item]) => ({
      day, label: new Date(`${day}T00:00:00`).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit' }),
      ...item, net: item.in - item.out, width: Math.max(item.in, item.out) / max * 100,
    }));
  }

  get accountBalances(): Array<{ name: string; type: string; amount: number }> {
    return this.accounts.map((account) => {
      const amount = this.allTransactions.filter((item) => item.accountId === account.id).reduce((sum, item) => {
        const value = this.amountOf(item);
        if (item.type === 'income' || item.type === 'loan_payment') return sum + value;
        if (item.type === 'expense' || item.type === 'loan_disbursement') return sum - value;
        return sum;
      }, 0);
      return { name: account.name, type: account.type, amount };
    }).sort((a, b) => b.amount - a.amount);
  }

  setPeriod(period: ReportPeriod) { this.period = period; this.updateRangeLabel(); }
  onMonthChange() { this.updateRangeLabel(); }

  private loadAccountsAndLoans() {
    this.api.getAccounts().subscribe({ next: (accounts) => { this.accounts = accounts || []; }, error: () => undefined });
    this.api.getLoans().subscribe({ next: (loans) => { this.loans = loans || []; }, error: () => undefined });
  }

  private loadTransactionPage(options: { limit: number; cursor?: string }, collected: Transaction[]) {
    this.api.getTransactions(options).subscribe({
      next: (page: TransactionListPage) => {
        const all = [...collected, ...(page.items || [])];
        if (page.nextCursor && all.length < 2_000) { this.loadTransactionPage({ ...options, cursor: page.nextCursor }, all); return; }
        this.allTransactions = all;
        const latestMonth = all.map((item) => this.dateKey(item).slice(0, 7)).sort().at(-1);
        if (latestMonth) this.selectedMonth = latestMonth;
        this.updateRangeLabel();
        this.loading = false;
      },
      error: () => { this.loading = false; this.error = 'Không thể tải báo cáo. Vui lòng thử lại.'; },
    });
  }

  private updateRangeLabel() {
    const monthTransactions = this.allTransactions.filter((item) => this.dateKey(item).startsWith(this.selectedMonth));
    if (this.period === 'week') {
      const end = monthTransactions.map((item) => this.dateKey(item)).sort().at(-1);
      if (end) {
        const startDate = new Date(`${end}T00:00:00`); startDate.setDate(startDate.getDate() - 6);
        this.rangeLabel = `${startDate.toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit' })} – ${new Date(`${end}T00:00:00`).toLocaleDateString('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric' })}`;
        return;
      }
    }
    const [year, month] = this.selectedMonth.split('-').map(Number);
    this.rangeLabel = Number.isFinite(year) && Number.isFinite(month) ? `Tháng ${month}/${year}` : 'Chưa chọn kỳ báo cáo';
  }

  private breakdown(type: 'income' | 'expense'): Breakdown[] {
    const totals = new Map<string, number>();
    for (const item of this.transactions.filter((transaction) => transaction.type === type)) {
      const name = item.name?.trim() || 'Chưa phân loại';
      totals.set(name, (totals.get(name) || 0) + this.amountOf(item));
    }
    const total = type === 'income' ? this.income : this.expense;
    return [...totals.entries()].map(([name, amount]) => ({ name, amount, percentage: total ? Math.round(amount / total * 100) : 0 })).sort((a, b) => b.amount - a.amount).slice(0, 5);
  }
  private totalFor(type: Transaction['type']) { return this.transactions.filter((item) => item.type === type).reduce((sum, item) => sum + this.amountOf(item), 0); }
  private amountOf(item: Transaction) { return this.amount(item.amount); }
  private amount(value?: string) { return Number.parseFloat(value || '0') || 0; }
  private dateKey(item: Transaction) { return (item.occurredAt || '').slice(0, 10); }
  private isoDate(date: Date) { return [date.getFullYear(), String(date.getMonth() + 1).padStart(2, '0'), String(date.getDate()).padStart(2, '0')].join('-'); }
}
