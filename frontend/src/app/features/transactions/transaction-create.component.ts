import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { Account } from '../../shared/models';
import { SelectMenuComponent, SelectMenuOption } from '../../shared/components/select-menu.component';

interface CategoryOption {
  id: string;
  label: string;
  icon: string;
}

@Component({
  selector: 'app-transaction-create',
  standalone: true,
  imports: [CommonModule, FormsModule, SelectMenuComponent],
  templateUrl: './transaction-create.component.html',
})
export class TransactionCreateComponent implements OnInit {
  type: 'expense' | 'income' = 'expense';
  amountString = '0';
  selectedCategory = 'Ăn uống';
  note = '';
  taxDeductible = false;
  selectedAccountId = '';
  accounts: Account[] = [];
  isSubmitting = false;
  occurredAt = new Date().toISOString().substring(0, 10);

  categories: CategoryOption[] = [
    { id: 'an-uong', label: 'Ăn uống', icon: 'restaurant' },
    { id: 'di-chuyen', label: 'Di chuyển', icon: 'directions_car' },
    { id: 'mua-sam', label: 'Mua sắm', icon: 'shopping_bag' },
    { id: 'giai-tri', label: 'Giải trí', icon: 'sports_esports' },
    { id: 'nha-cua', label: 'Nhà cửa', icon: 'home' },
    { id: 'luong', label: 'Thu nhập / Lương', icon: 'payments' },
  ];

  get numericAmount(): number {
    return Number.parseInt(this.amountString, 10) || 0;
  }

  get formattedAmount(): string {
    return new Intl.NumberFormat('vi-VN').format(this.numericAmount);
  }

  get accountOptions(): SelectMenuOption[] {
    return this.accounts.map((account) => ({ value: account.id, label: `${account.name} (${account.currency})` }));
  }

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private api: ApiService,
    private toast: ToastService
  ) {}

  ngOnInit() {
    this.loadAccounts();
    this.route.queryParamMap.subscribe((params) => {
      const requestedType = params.get('type');
      if (requestedType === 'income' || requestedType === 'expense') {
        this.setType(requestedType);
      }
    });
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

  setType(newType: 'expense' | 'income') {
    this.type = newType;
    if (newType === 'income' && this.selectedCategory === 'Ăn uống') {
      this.selectedCategory = 'Thu nhập / Lương';
    }
  }

  selectCategory(catLabel: string) {
    this.selectedCategory = catLabel;
  }

  onAmountInput(event: Event) {
    const input = event.target as HTMLInputElement;
    const digits = input.value.replace(/\D/g, '').slice(0, 12).replace(/^0+(?=\d)/, '');
    this.amountString = digits || '0';
    input.value = this.formattedAmount;
  }

  submitTransaction() {
    if (this.numericAmount <= 0) {
      this.toast.show('Vui lòng nhập số tiền hợp lệ.', 'error');
      return;
    }
    if (!this.selectedAccountId && !this.accounts.length) {
      this.toast.show('Hãy tạo hoặc chọn một tài khoản trước khi ghi giao dịch.', 'error');
      return;
    }

    this.isSubmitting = true;
    const payload = {
      accountId: this.selectedAccountId || (this.accounts[0]?.id ?? ''),
      type: this.type as 'income' | 'expense',
      amount: String(this.numericAmount),
      currency: 'VND',
      name: `${this.selectedCategory} - ${this.type === 'expense' ? 'Chi phí' : 'Thu nhập'}`,
      note: `${this.note}${this.taxDeductible ? ' [Khấu trừ thuế]' : ''}`,
      occurredAt: new Date(`${this.occurredAt}T12:00:00`).toISOString(),
    };

    this.api.createTransaction(payload).subscribe({
      next: () => {
        this.isSubmitting = false;
        this.toast.show(
          `Đã lưu thành công ${this.type === 'expense' ? 'Chi phí' : 'Thu nhập'} ${this.formattedAmount} VND!`,
          'success'
        );
        this.router.navigate(['/dashboard']);
      },
      error: () => {
        this.isSubmitting = false;
        this.toast.show('Không thể lưu giao dịch. Vui lòng thử lại.', 'error');
      },
    });
  }
}
