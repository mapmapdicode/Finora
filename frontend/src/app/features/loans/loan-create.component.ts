import { Component, OnInit } from '@angular/core';
import { CommonModule, Location } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { Account, Loan } from '../../shared/models';

@Component({
  selector: 'app-loan-create',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './loan-create.component.html',
})
export class LoanCreateComponent implements OnInit {
  borrowerName = '';
  newBorrowerName = '';
  isAddingNewBorrower = false;
  borrowers: string[] = ['Nguyễn Văn A', 'Trần Thị B', 'Lê Văn C', 'Công ty TNHH Hưng Thịnh'];

  amountString = '0';
  interestRate = 2000; // 2k, 3k, 4k, 5k
  interestPeriod: 'monthly' | 'flexible' = 'monthly';
  loanDate = new Date().toISOString().substring(0, 10);
  dueDate = new Date(Date.now() + 30 * 86400000).toISOString().substring(0, 10);
  notes = '';
  selectedAccountId = '';
  accounts: Account[] = [];
  isSubmitting = false;

  get numericAmount(): number {
    return Number.parseInt(this.amountString.replace(/\D/g, ''), 10) || 0;
  }

  get formattedAmount(): string {
    return new Intl.NumberFormat('vi-VN').format(this.numericAmount);
  }

  get annualRateEquivalent(): number {
    return (this.interestRate * 365 * 100) / 1000000;
  }

  constructor(
    private location: Location,
    private router: Router,
    private api: ApiService,
    private toast: ToastService
  ) {}

  ngOnInit() {
    this.loadAccounts();
    this.loadExistingBorrowers();
  }

  loadAccounts() {
    this.api.getAccounts().subscribe({
      next: (accs) => {
        this.accounts = accs;
        if (accs.length) {
          this.selectedAccountId = accs[0].id;
        }
      },
    });
  }

  loadExistingBorrowers() {
    this.api.getLoans().subscribe({
      next: (loans: Loan[]) => {
        const set = new Set(this.borrowers);
        loans.forEach((loan) => {
          if (loan.counterparty && loan.counterparty.trim()) {
            set.add(loan.counterparty.trim());
          }
        });
        this.borrowers = Array.from(set);
        if (this.borrowers.length && !this.borrowerName) {
          this.borrowerName = this.borrowers[0];
        }
      },
      error: () => {
        if (this.borrowers.length && !this.borrowerName) {
          this.borrowerName = this.borrowers[0];
        }
      },
    });
  }

  goBack() {
    this.location.back();
  }

  onBorrowerSelectChange(event: Event) {
    const val = (event.target as HTMLSelectElement).value;
    if (val === '__add_new__') {
      this.isAddingNewBorrower = true;
      this.newBorrowerName = '';
    } else {
      this.borrowerName = val;
      this.isAddingNewBorrower = false;
    }
  }

  toggleAddNewBorrower() {
    this.isAddingNewBorrower = !this.isAddingNewBorrower;
    if (this.isAddingNewBorrower) {
      this.newBorrowerName = '';
    }
  }

  addQuickBorrower() {
    const name = this.newBorrowerName.trim();
    if (!name) {
      this.toast.show('Vui lòng nhập tên người vay mới.', 'error');
      return;
    }

    if (!this.borrowers.includes(name)) {
      this.borrowers.unshift(name);
    }
    this.borrowerName = name;
    this.isAddingNewBorrower = false;
    this.newBorrowerName = '';
    this.toast.show(`Đã thêm người vay: ${name}`, 'success');
  }

  onAmountInput(event: Event) {
    const raw = (event.target as HTMLInputElement).value.replace(/\D/g, '');
    const num = Number.parseInt(raw, 10) || 0;
    this.amountString = new Intl.NumberFormat('vi-VN').format(num);
  }

  setInterestRate(rate: number) {
    this.interestRate = rate;
  }

  setInterestPeriod(period: 'monthly' | 'flexible') {
    this.interestPeriod = period;
  }

  submitLoan() {
    const targetBorrower = this.isAddingNewBorrower ? this.newBorrowerName.trim() : this.borrowerName.trim();

    if (!targetBorrower) {
      this.toast.show('Vui lòng chọn hoặc nhập tên người vay.', 'error');
      return;
    }

    if (this.numericAmount <= 0) {
      this.toast.show('Vui lòng nhập số tiền cho vay hợp lệ.', 'error');
      return;
    }

    this.isSubmitting = true;

    const payload = {
      counterparty: targetBorrower,
      direction: 'receivable' as const,
      principalInitial: String(this.numericAmount),
      annualRate: String(this.annualRateEquivalent),
      dailyRatePerMillion: String(this.interestRate),
      settlementAccountId: this.selectedAccountId || (this.accounts[0]?.id ?? ''),
      dayCountBasis: 'ACT_365' as const,
      startAt: new Date(this.loanDate).toISOString(),
      dueAt: this.dueDate ? new Date(this.dueDate).toISOString() : undefined,
      note: `${this.notes} [Lãi: ${this.interestRate / 1000}k/triệu/ngày - Thu lãi: ${this.interestPeriod === 'monthly' ? 'Hàng tháng' : 'Linh hoạt'}]`,
    };

    this.api.createLoan(payload).subscribe({
      next: () => {
        this.isSubmitting = false;
        this.toast.show(`Đã tạo hồ sơ vay thành công cho ${targetBorrower}!`, 'success');
        this.router.navigate(['/loans']);
      },
      error: () => {
        this.isSubmitting = false;
        this.toast.show('Không thể tạo khoản vay. Vui lòng thử lại.', 'error');
      },
    });
  }
}
