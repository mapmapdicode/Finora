import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { Account, Customer } from '../../shared/models';
import { normalizeVndAmount } from '../../shared/money-input';

@Component({
  selector: 'app-loan-create',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './loan-create.component.html',
})
export class LoanCreateComponent implements OnInit {
  customers: Customer[] = [];
  selectedCustomerId = '';
  newBorrowerName = '';
  newBorrowerPhone = '';
  isAddingNewBorrower = false;
  customersLoading = false;
  creatingCustomer = false;
  customerLoadError = '';

  amountString = '';
  interestRate = 2000; // 2k, 3k, 4k, 5k
  interestPeriod: 'monthly' | 'flexible' = 'monthly';
  loanDate = new Date().toISOString().substring(0, 10);
  dueDate = new Date(Date.now() + 30 * 86400000).toISOString().substring(0, 10);
  notes = '';
  selectedAccountId = '';
  accounts: Account[] = [];
  isSubmitting = false;

  get numericAmount(): number {
    return Number(normalizeVndAmount(this.amountString)) || 0;
  }

  get formattedAmount(): string {
    return new Intl.NumberFormat('vi-VN').format(this.numericAmount);
  }

  get annualRateEquivalent(): number {
    return (this.interestRate * 365 * 100) / 1000000;
  }

  constructor(
    private router: Router,
    private api: ApiService,
    private toast: ToastService
  ) {}

  ngOnInit() {
    this.loadAccounts();
    this.loadCustomers();
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

  loadCustomers() {
    this.customersLoading = true;
    this.customerLoadError = '';
    this.api.getCustomers().subscribe({
      next: (customers) => {
        this.customers = customers;
        this.customersLoading = false;
      },
      error: () => {
        this.customersLoading = false;
        this.customerLoadError = 'Không thể tải danh sách khách hàng.';
      },
    });
  }

  onBorrowerSelectionChange(value: string) {
    if (value === '__add_new__') {
      this.openAddNewBorrower();
    } else {
      this.selectedCustomerId = value;
      this.isAddingNewBorrower = false;
    }
  }

  toggleAddNewBorrower() {
    if (this.isAddingNewBorrower) {
      this.isAddingNewBorrower = false;
      return;
    }
    this.openAddNewBorrower();
  }

  private openAddNewBorrower() {
    this.selectedCustomerId = '';
    this.isAddingNewBorrower = true;
    this.newBorrowerName = '';
    this.newBorrowerPhone = '';
  }

  addQuickBorrower() {
    const name = this.newBorrowerName.trim();
    if (!name) {
      this.toast.show('Vui lòng nhập tên người vay mới.', 'error');
      return;
    }

    this.creatingCustomer = true;
    this.api.createCustomer({ name, phone: this.newBorrowerPhone.trim() || undefined }).subscribe({
      next: (customer) => {
        this.creatingCustomer = false;
        const existingIndex = this.customers.findIndex((item) => item.id === customer.id);
        if (existingIndex >= 0) {
          this.customers[existingIndex] = customer;
        } else {
          this.customers = [customer, ...this.customers];
        }
        this.selectedCustomerId = customer.id;
        this.isAddingNewBorrower = false;
        this.newBorrowerName = '';
        this.newBorrowerPhone = '';
        this.toast.show(`Đã lưu và chọn khách hàng ${customer.name}.`, 'success');
      },
      error: (error) => {
        this.creatingCustomer = false;
        this.toast.show(error?.error?.message || 'Không thể lưu khách hàng. Vui lòng thử lại.', 'error');
      },
    });
  }

  onAmountInput(event: Event) {
    this.amountString = (event.target as HTMLInputElement).value;
  }

  onAmountBlur(event: Event) {
    const formatted = this.numericAmount > 0 ? new Intl.NumberFormat('vi-VN').format(this.numericAmount) : '';
    this.amountString = formatted;
    (event.target as HTMLInputElement).value = formatted;
  }

  setInterestRate(rate: number) {
    this.interestRate = rate;
  }

  setInterestPeriod(period: 'monthly' | 'flexible') {
    this.interestPeriod = period;
  }

  submitLoan() {
    if (this.isAddingNewBorrower) {
      this.toast.show('Hãy lưu khách hàng mới trước khi tạo khoản vay.', 'error');
      return;
    }

    const customer = this.customers.find((item) => item.id === this.selectedCustomerId);
    if (!customer) {
      this.toast.show('Vui lòng chọn khách hàng.', 'error');
      return;
    }

    if (this.numericAmount <= 0) {
      this.toast.show('Vui lòng nhập số tiền cho vay hợp lệ.', 'error');
      return;
    }

    this.isSubmitting = true;

    const payload = {
      customerId: customer.id,
      counterparty: customer.name,
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
        this.toast.show(`Đã tạo hồ sơ vay thành công cho ${customer.name}!`, 'success');
        this.router.navigate(['/loans']);
      },
      error: () => {
        this.isSubmitting = false;
        this.toast.show('Không thể tạo khoản vay. Vui lòng thử lại.', 'error');
      },
    });
  }
}
