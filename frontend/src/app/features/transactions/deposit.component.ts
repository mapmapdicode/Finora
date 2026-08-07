import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { Account } from '../../shared/models';

@Component({
  selector: 'app-deposit',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './deposit.component.html',
})
export class DepositComponent implements OnInit {
  amount = 0;
  quickChips = [1000000, 5000000, 10000000, 50000000];
  isSubmitting = false;
  accounts: Account[] = [];
  selectedAccountId = '';

  get formattedAmount(): string {
    return new Intl.NumberFormat('vi-VN').format(this.amount || 0);
  }

  constructor(
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

  onAmountInput(event: Event) {
    const raw = (event.target as HTMLInputElement).value.replace(/\D/g, '');
    this.amount = Number.parseInt(raw, 10) || 0;
  }

  selectChip(chipAmount: number) {
    this.amount = chipAmount;
  }

  formatNumber(val: number): string {
    return new Intl.NumberFormat('vi-VN').format(val);
  }

  submitDeposit() {
    if (this.amount <= 0) {
      this.toast.show('Vui lòng nhập số tiền nạp hợp lệ.', 'error');
      return;
    }

    this.isSubmitting = true;

    const payload = {
      accountId: this.selectedAccountId || (this.accounts[0]?.id ?? ''),
      type: 'income' as const,
      amount: String(this.amount),
      currency: 'VND',
      name: 'Nạp tiền vào ví',
      note: 'Ghi nhận nạp tiền thủ công',
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
