import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { Account, Loan, LoanAccruals, LoanPayment, LoanPaymentRequest } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-loan-list',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './loan-list.component.html'
})
export class LoanListComponent implements OnInit {
  loans: Loan[] = [];
  accounts: Account[] = [];
  accrualsMap: Record<string, LoanAccruals> = {};
  paymentRequests: Record<string, LoanPaymentRequest | null> = {};
  creationForm: FormGroup;
  paymentForm: FormGroup;
  requestForm: FormGroup;
  statusMessage = '';
  selectedLoanId = '';
  selectedLoanForPayment = '';

  constructor(private api: ApiService, private fb: FormBuilder, public auth: AuthService) {
    this.creationForm = this.fb.group({
      counterparty: ['', Validators.required],
      direction: ['payable', Validators.required],
      principalInitial: ['', Validators.required],
      annualRate: ['0', Validators.required],
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
    });

    this.requestForm = this.fb.group({
      loanId: [''],
      amount: [''],
      currency: ['VND'],
      expiresAt: [''],
    });
  }

  ngOnInit() {
    this.api.getLoans().subscribe((items: Loan[]) => {
      this.loans = items;
    });
    this.api.getAccounts().subscribe((items) => (this.accounts = items));
  }

  submitLoan() {
    if (!this.auth.canMutate) return;
    if (this.creationForm.invalid) return;
    const payload = {
      counterparty: this.creationForm.value.counterparty,
      direction: this.creationForm.value.direction,
      principalInitial: this.creationForm.value.principalInitial,
      annualRate: this.creationForm.value.annualRate,
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
          counterparty: '',
          direction: 'payable',
          principalInitial: '',
          annualRate: '0',
          dayCountBasis: 'ACT_365',
          portfolioId: '',
          interestCompounding: false,
          startAt: '',
          dueAt: '',
        });
        this.reload();
      },
      error: () => {
        this.statusMessage = 'Tạo khoản vay thất bại.';
      },
    });
  }

  reload() {
    this.api.getLoans().subscribe((items: Loan[]) => (this.loans = items));
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

  openPayment(loanId: string) {
    if (!this.auth.canMutate) return;
    this.selectedLoanForPayment = loanId;
    this.paymentForm.patchValue({
      loanId,
      principalAmount: '',
      interestAmount: '0',
      feeAmount: '0',
      waivedAmount: '0',
      occurredAt: '',
    });
  }

  closePayment() {
    this.selectedLoanForPayment = '';
  }

  submitPayment() {
    if (!this.auth.canMutate) return;
    const loanId = this.paymentForm.value.loanId || this.selectedLoanForPayment;
    if (!loanId || this.paymentForm.invalid) return;
    const payload = {
      principalAmount: this.paymentForm.value.principalAmount || '0',
      interestAmount: this.paymentForm.value.interestAmount || '0',
      feeAmount: this.paymentForm.value.feeAmount || '0',
      waivedAmount: this.paymentForm.value.waivedAmount || '0',
      occurredAt: this.paymentForm.value.occurredAt || undefined,
    };
    this.api.createLoanPayment(loanId, payload).subscribe({
      next: (_result: LoanPayment) => {
        this.statusMessage = 'Đã ghi nhận khoản trả nợ.';
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
        amount: this.requestForm.value.amount || undefined,
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
}
