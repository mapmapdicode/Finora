import { Component, OnInit } from '@angular/core';
import { CommonModule, Location } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { Account } from '../../shared/models';

interface SourceOption {
  id: string;
  name: string;
  subtext: string;
  icon: string;
}

@Component({
  selector: 'app-deposit',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './deposit.component.html',
})
export class DepositComponent implements OnInit {
  amount = 5000000;
  quickChips = [1000000, 5000000, 10000000, 50000000];
  selectedSourceId = 'techcombank';
  isSubmitting = false;
  accounts: Account[] = [];
  selectedAccountId = '';

  sources: SourceOption[] = [
    { id: 'techcombank', name: 'Techcombank', subtext: '**** 8899', icon: 'account_balance' },
    { id: 'mbbank', name: 'MBBank', subtext: '**** 1234', icon: 'account_balance' },
    { id: 'visa', name: 'Thẻ VISA / Mastercard', subtext: '**** 4242', icon: 'credit_card' },
  ];

  get formattedAmount(): string {
    return new Intl.NumberFormat('vi-VN').format(this.amount || 0);
  }

  constructor(
    private location: Location,
    private router: Router,
    private api: ApiService,
    private toast: ToastService
  ) {}

  ngOnInit() {
    this.loadAccounts();
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

  goBack() {
    this.location.back();
  }

  onAmountInput(event: Event) {
    const raw = (event.target as HTMLInputElement).value.replace(/\D/g, '');
    this.amount = Number.parseInt(raw, 10) || 0;
  }

  selectChip(chipAmount: number) {
    this.amount = chipAmount;
  }

  selectSource(id: string) {
    this.selectedSourceId = id;
  }

  formatNumber(val: number): string {
    return new Intl.NumberFormat('vi-VN').format(val);
  }

  submitDeposit() {
    if (this.amount <= 0) {
      this.toast.show('Vui lòng nhập số tiền nạp hợp lệ.', 'error');
      return;
    }

    const selectedSource = this.sources.find((s) => s.id === this.selectedSourceId);
    this.isSubmitting = true;

    const payload = {
      accountId: this.selectedAccountId || (this.accounts[0]?.id ?? ''),
      type: 'income' as const,
      amount: String(this.amount),
      currency: 'VND',
      name: `Nạp tiền từ ${selectedSource?.name || 'Ngân hàng'}`,
      note: `Nạp tiền tự động qua ${selectedSource?.name} (${selectedSource?.subtext})`,
      occurredAt: new Date().toISOString(),
    };

    this.api.createTransaction(payload).subscribe({
      next: () => {
        this.isSubmitting = false;
        this.toast.show(`Nạp thành công ${this.formattedAmount} VND vào tài khoản!`, 'success');
        this.router.navigate(['/dashboard']);
      },
      error: () => {
        this.isSubmitting = false;
        this.toast.show('Thao tác nạp tiền thất bại. Vui lòng thử lại.', 'error');
      },
    });
  }
}
