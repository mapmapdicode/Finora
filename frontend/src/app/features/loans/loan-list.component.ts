import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { Account, Customer, Loan, LoanAccruals, LoanPayment, LoanPaymentRequest, LoanPortfolioSummary, LoanScheduleItem } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';
import { IconComponent } from '../../shared/icons/icon.component';
import { normalizeVndAmount } from '../../shared/money-input';
import { RouterLink } from '@angular/router';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { VndMoneyPipe } from '../../shared/pipes/vnd-money.pipe';

@Component({
  selector: 'app-loan-list',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, IconComponent, TranslatePipe, VndMoneyPipe, RouterLink],
  templateUrl: './loan-list.component.html'
})
export class LoanListComponent implements OnInit {
  loans: Loan[] = [];
  customers: Customer[] = [];
  accounts: Account[] = [];
  accrualsMap: Record<string, LoanAccruals> = {};
  paymentRequests: Record<string, LoanPaymentRequest | null> = {};
  summary: LoanPortfolioSummary = { activePrincipal: '0', dailyInterest: '0', accruedInterest: '0', paidInterest: '0' };
  schedule: LoanScheduleItem[] = [];
  paymentHistory: LoanPayment[] = [];
  loanListView: 'open' | 'closed' = 'open';
  creationForm: FormGroup;
  paymentForm: FormGroup;
  requestForm: FormGroup;
  statusMessage = '';
  selectedLoanId = '';
  selectedLoanForPayment = '';
  selectedLoanForHistory: Loan | null = null;
  historyLoading = false;
  showCreateForm = false;

  paymentMode: 'interest_only' | 'principal_only' | 'both' = 'both';
  activePaymentLoan: Loan | null = null;

  get dailyRatePerMillion(): number {
    return Number(normalizeVndAmount(this.creationForm?.value.dailyRatePerMillion)) || 0;
  }

  get annualRateEquivalent(): number {
    return (this.dailyRatePerMillion * 365 * 100) / 1_000_000;
  }

  /** Hợp đồng đã tất toán chỉ nằm trong lịch sử, không lẫn với các khoản còn dư nợ. */
  isOpenLoan(loan: Loan): boolean {
    return !['closed', 'settled', 'cancelled'].includes((loan.status || 'active').toLowerCase());
  }

  get openLoans(): Loan[] {
    return this.loans.filter((loan) => this.isOpenLoan(loan));
  }

  get closedLoans(): Loan[] {
    return this.loans.filter((loan) => !this.isOpenLoan(loan));
  }

  get visibleLoans(): Loan[] {
    return this.loanListView === 'open' ? this.openLoans : this.closedLoans;
  }

  constructor(private api: ApiService, private fb: FormBuilder, public auth: AuthService) {
    this.creationForm = this.fb.group({
      customerId: ['', Validators.required],
      direction: ['receivable', Validators.required],
      principalInitial: ['', Validators.required],
      annualRate: ['0', Validators.required],
	  dailyRatePerMillion: ['', Validators.required],
	  settlementAccountId: ['', Validators.required],
      dayCountBasis: ['ACT_365'],
      portfolioId: [''],
      interestCompounding: [false],
      startAt: [''],
      dueAt: [''],
    });

    this.paymentForm = this.fb.group({
      loanId: [''],
      principalAmount: ['0', Validators.required],
      interestAmount: ['0'],
      feeAmount: ['0'],
      waivedAmount: ['0'],
      occurredAt: [''],
	  accountId: ['', Validators.required],
    });

    this.requestForm = this.fb.group({
      loanId: [''],
      amount: [''],
      currency: ['VND'],
      expiresAt: [''],
    });
  }

  ngOnInit() {
    this.reload();
    this.api.getAccounts().subscribe((items) => (this.accounts = items));
    this.api.getCustomers().subscribe({
      next: (items) => (this.customers = items),
      error: () => (this.statusMessage = 'Không thể tải danh sách khách hàng.'),
    });
  }

  get todayDateString(): string {
    const d = new Date();
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  get lastInterestDate(): Date {
    const loanId = this.selectedLoanForPayment;
    const accrual = this.accrualsMap[loanId];
    const loan = this.activePaymentLoan;
    
    let dateStr = accrual?.lastInterestPaidDate || loan?.startAt;
    if (!dateStr || dateStr.startsWith('0001')) {
      dateStr = loan?.startAt || '';
    }
    if (!dateStr || dateStr.startsWith('0001')) {
      return new Date();
    }
    
    const d = new Date(dateStr);
    if (isNaN(d.getTime()) || d.getFullYear() < 2000) {
      return new Date();
    }
    return d;
  }

  get interestDaysCount(): number {
    const lastDate = this.lastInterestDate;
    const paymentDateStr = this.paymentForm?.value?.occurredAt || this.todayDateString;
    const paymentDate = new Date(paymentDateStr);
    if (isNaN(paymentDate.getTime())) return 0;

    const d1 = new Date(lastDate.getFullYear(), lastDate.getMonth(), lastDate.getDate());
    const d2 = new Date(paymentDate.getFullYear(), paymentDate.getMonth(), paymentDate.getDate());
    const diffMs = d2.getTime() - d1.getTime();
    return Math.max(0, Math.floor(diffMs / (1000 * 60 * 60 * 24)));
  }

  get computedInterestByDays(): number {
    const loanId = this.selectedLoanForPayment;
    if (!loanId) return 0;
    const dailyInterestStr = this.accrualsMap[loanId]?.dailyInterest || '0';
    const dailyInterest = Number(normalizeVndAmount(dailyInterestStr)) || 0;
    const days = this.interestDaysCount;
    return Math.round(days * dailyInterest);
  }

  openPayment(loanId: string) {
    if (!this.auth.canMutate) return;
    this.selectedLoanForPayment = loanId;
    this.activePaymentLoan = this.loans.find(l => l.id === loanId) || null;
    this.paymentMode = 'both';
    this.paymentForm.patchValue({
      loanId,
      principalAmount: '0',
      interestAmount: '0',
      feeAmount: '0',
      waivedAmount: '0',
      occurredAt: this.todayDateString,
      accountId: '',
    });
    this.scrollToElement('loan-payment-form');
  }

  closePayment() {
    this.selectedLoanForPayment = '';
    this.activePaymentLoan = null;
  }

  setPaymentMode(mode: 'interest_only' | 'principal_only' | 'both') {
    this.paymentMode = mode;
    if (mode === 'interest_only') {
      this.paymentForm.patchValue({ principalAmount: '0' }, { emitEvent: false });
    } else if (mode === 'principal_only') {
      this.paymentForm.patchValue({ interestAmount: '0' }, { emitEvent: false });
    }
  }

  setQuickPaymentAmount(type: 'accrued' | 'principal' | 'full') {
    const loanId = this.selectedLoanForPayment;
    const accruedMapInterest = Number(normalizeVndAmount(this.accrualsMap[loanId]?.totalAccruedInterest || '0')) || 0;
    const accruedByDays = this.computedInterestByDays;
    const accrued = accruedByDays > 0 ? accruedByDays : accruedMapInterest;

    const loan = this.activePaymentLoan;
    const principal = Number(normalizeVndAmount(loan?.principalBalance || '0')) || 0;

    if (type === 'accrued') {
      this.paymentMode = 'interest_only';
      this.paymentForm.patchValue({
        interestAmount: new Intl.NumberFormat('vi-VN').format(accrued),
        principalAmount: '0',
      }, { emitEvent: false });
    } else if (type === 'principal') {
      this.paymentMode = 'principal_only';
      this.paymentForm.patchValue({
        interestAmount: '0',
        principalAmount: new Intl.NumberFormat('vi-VN').format(principal),
      }, { emitEvent: false });
    } else if (type === 'full') {
      this.paymentMode = 'both';
      this.paymentForm.patchValue({
        interestAmount: new Intl.NumberFormat('vi-VN').format(accrued),
        principalAmount: new Intl.NumberFormat('vi-VN').format(principal),
      }, { emitEvent: false });
    }
  }

  onPaymentComponentBlur(control: 'interestAmount' | 'principalAmount', event: Event) {
    const input = event.target as HTMLInputElement;
    const formatted = this.formatVndAmount(input.value);
    this.paymentForm.get(control)?.setValue(formatted, { emitEvent: false });
    input.value = formatted;

  }

  submitLoan() {
    if (!this.auth.canMutate) return;
    if (this.creationForm.invalid) {
      this.creationForm.markAllAsTouched();
      this.statusMessage = 'Vui lòng chọn khách hàng, tài khoản và nhập số tiền gốc cùng lãi/ngày hợp lệ.';
      return;
    }
    const payload = {
      customerId: this.creationForm.value.customerId,
      direction: this.creationForm.value.direction,
      principalInitial: normalizeVndAmount(this.creationForm.value.principalInitial),
      annualRate: this.creationForm.value.annualRate,
      dailyRatePerMillion: normalizeVndAmount(this.creationForm.value.dailyRatePerMillion),
      settlementAccountId: this.creationForm.value.settlementAccountId,
      dayCountBasis: this.creationForm.value.dayCountBasis || 'ACT_365',
      portfolioId: this.creationForm.value.portfolioId || '',
      interestCompounding: !!this.creationForm.value.interestCompounding,
      startAt: this.creationForm.value.startAt || undefined,
      dueAt: this.creationForm.value.dueAt || undefined,
    };
    this.api.createLoan(payload).subscribe({
      next: () => {
        this.statusMessage = 'Tạo khoản vay thành công.';
        this.creationForm.reset({
          customerId: '',
          direction: 'receivable',
          principalInitial: '',
          annualRate: '0',
		  dailyRatePerMillion: '',
		  settlementAccountId: '',
          dayCountBasis: 'ACT_365',
          portfolioId: '',
          interestCompounding: false,
          startAt: '',
          dueAt: '',
        });
        this.reload();
      },
      error: (error) => {
        this.statusMessage = error?.error?.message || 'Tạo khoản vay thất bại.';
      },
    });
  }

  reload() {
	this.api.getLoans().subscribe((items: Loan[]) => {
	  this.loans = items;
	  items.forEach((loan) => this.api.getLoanAccruals(loan.id).subscribe((accrual) => this.accrualsMap[loan.id] = accrual));
	});
	this.api.getLoanSummary().subscribe((summary) => this.summary = summary);
	this.api.getLoanSchedule().subscribe((schedule) => this.schedule = schedule);
  }

  loadAccruals(loanId: string) {
    if (this.selectedLoanId === loanId) {
      this.selectedLoanId = '';
      return;
    }
    this.selectedLoanId = loanId;
    this.api.getLoanAccruals(loanId).subscribe({
      next: (acc) => (this.accrualsMap[loanId] = acc),
      error: () => {
        this.statusMessage = 'Không thể lấy chi tiết lãi accrual.';
      },
    });
  }





  openPaymentHistory(loan: Loan) {
    if (this.selectedLoanForHistory?.id === loan.id) {
      this.selectedLoanForHistory = null;
      this.paymentHistory = [];
      return;
    }
    this.selectedLoanForHistory = loan;
    this.paymentHistory = [];
    this.historyLoading = true;
    this.scrollToElement('loan-history-section');
    this.api.getLoanPayments(loan.id).subscribe({
      next: (items) => {
        this.paymentHistory = [...items].sort((a, b) => new Date(b.occurredAt).getTime() - new Date(a.occurredAt).getTime());
        this.historyLoading = false;
        this.scrollToElement('loan-history-section');
      },
      error: () => {
        this.historyLoading = false;
        this.statusMessage = 'Không thể tải lịch sử thanh toán khoản vay.';
      },
    });
  }

  private scrollToElement(elementId: string) {
    setTimeout(() => {
      const el = document.getElementById(elementId);
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }
    }, 80);
  }

  paymentTotal(payment: LoanPayment): number {
    return ['principalAmount', 'interestAmount', 'feeAmount']
      .reduce((sum, field) => sum + (Number.parseFloat(payment[field as keyof LoanPayment] as string) || 0), 0);
  }

  setQuickRate(value: number) {
    this.creationForm.patchValue({
      dailyRatePerMillion: this.formatVndAmount(String(value)),
      annualRate: this.toAnnualRate(value),
    });
  }

  onLoanMoneyBlur(control: 'principalInitial' | 'dailyRatePerMillion', event: Event) {
    const input = event.target as HTMLInputElement;
    const formatted = this.formatVndAmount(input.value);
    this.creationForm.get(control)?.setValue(formatted, { emitEvent: false });
    input.value = formatted;
    if (control === 'dailyRatePerMillion') this.syncAnnualRate();
  }

  syncAnnualRate() {
    this.creationForm.patchValue({ annualRate: this.toAnnualRate(this.dailyRatePerMillion) }, { emitEvent: false });
  }

  private toAnnualRate(value: number) {
    return ((value * 365 * 100) / 1_000_000).toFixed(4);
  }

  private formatVndAmount(value: unknown): string {
    const normalized = normalizeVndAmount(value);
    return normalized ? new Intl.NumberFormat('vi-VN').format(Number(normalized)) : '';
  }

  remove(item: Loan) {
    if (!this.auth.canMutate || !window.confirm(`Xóa khoản vay với “${item.counterparty}”?`)) return;
    this.api.deleteLoan(item.id).subscribe({
      next: () => {
        this.statusMessage = 'Đã xóa khoản vay.';
        this.reload();
      },
      error: () => (this.statusMessage = 'Không thể xóa khoản vay có dữ liệu thanh toán liên kết.'),
    });
  }

  submitPayment() {
    if (!this.auth.canMutate) return;
    const loanId = this.paymentForm.value.loanId || this.selectedLoanForPayment;
    if (!loanId || this.paymentForm.invalid) return;
    const principal = normalizeVndAmount(this.paymentForm.value.principalAmount) || '0';
    const interest = normalizeVndAmount(this.paymentForm.value.interestAmount) || '0';

    if (Number(principal) <= 0 && Number(interest) <= 0) {
      this.statusMessage = 'Vui lòng nhập số tiền ghi nhận (lãi hoặc gốc).';
      return;
    }

    const payload = {
      principalAmount: principal,
      interestAmount: interest,
      feeAmount: normalizeVndAmount(this.paymentForm.value.feeAmount) || '0',
      waivedAmount: normalizeVndAmount(this.paymentForm.value.waivedAmount) || '0',
      occurredAt: this.paymentForm.value.occurredAt || undefined,
      accountId: this.paymentForm.value.accountId,
    };
    this.api.createLoanPayment(loanId, payload).subscribe({
      next: (_result: LoanPayment) => {
        this.statusMessage = 'Đã ghi nhận khoản thu / trả nợ thành công.';
        this.reload();
        this.closePayment();
      },
      error: () => {
        this.statusMessage = 'Tạo khoản trả nợ thất bại.';
      },
    });
  }

  openRequest(loanId: string) {
    if (!this.auth.canMutate) return;
    this.requestForm.patchValue({ loanId, amount: '', currency: 'VND', expiresAt: '' });
    this.paymentRequests[loanId] = null;
  }

  requestCode(loanId: string) {
    if (!this.auth.canMutate) return;
    this.api
      .createLoanPaymentRequest(loanId, {
        amount: normalizeVndAmount(this.requestForm.value.amount) || undefined,
        currency: this.requestForm.value.currency || 'VND',
        expiresAt: this.requestForm.value.expiresAt || undefined,
        note: 'Yêu cầu thanh toán QR',
      })
      .subscribe({
        next: (res: LoanPaymentRequest) => {
          this.paymentRequests[loanId] = res;
          this.statusMessage = `Đã tạo yêu cầu thanh toán (QR). Mã: ${res.paymentCode || res.id || loanId}`;
        },
        error: () => {
          this.statusMessage = 'Tạo yêu cầu thanh toán thất bại.';
        },
      });
  }

  statusText(item: Loan) {
    return item.direction === 'receivable' ? 'Phải thu' : 'Phải trả';
  }

  loanStatusLabel(status: string | undefined) {
    const labels: Record<string, string> = {
      active: 'Đang hiệu lực',
      overdue: 'Quá hạn',
      settled: 'Đã tất toán',
      closed: 'Đã tất toán',
      cancelled: 'Đã huỷ',
    };
    return labels[status || ''] || status || 'Chưa xác định';
  }
}
