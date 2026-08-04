import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { CommonModule, NgForOf, NgIf, NgClass } from '@angular/common';
import { ApiService } from '../../core/services/api.service';
import { AuthService } from '../../core/services/auth.service';
import { Portfolio } from '../../shared/models';
import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-account-list',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, NgForOf, NgIf, NgClass, TranslatePipe, IconComponent],
  templateUrl: './account-list.component.html',
  styleUrl: './account-list.component.css'
})
export class AccountListComponent implements OnInit {
  form!: FormGroup;
  accounts: any[] = [];
  portfolios: Portfolio[] = [];
  isCreating = false;
  saving = false;
  loading = true;
  accountLoadError = '';
  statusMessage = '';

  constructor(private fb: FormBuilder, private api: ApiService, public auth: AuthService) {}

  ngOnInit() {
    this.form = this.fb.group({
      name: ['', Validators.required],
      type: ['cash', Validators.required],
      currency: ['VND', Validators.required],
      portfolioId: [''],
    });
    this.api.getPortfolios().subscribe({
      next: (items) => {
        this.portfolios = items;
        const defaultPortfolio = items[0]?.id || '';
        this.form.patchValue({ portfolioId: defaultPortfolio });
      },
      error: () => {
        this.portfolios = [];
      },
    });
    this.reload();
  }

  reload() {
    this.loading = true;
    this.accountLoadError = '';
    this.api.getAccounts().subscribe({
      next: (items) => {
        this.accounts = items;
        this.loading = false;
      },
      error: () => {
        this.accounts = [];
        this.loading = false;
        this.accountLoadError = 'Không thể tải danh sách tài khoản.';
      },
    });
  }

  openCreate() {
    if (!this.auth.canMutate) return;
    this.isCreating = true;
    this.statusMessage = '';
  }

  closeCreate() {
    if (this.saving) return;
    this.isCreating = false;
  }

  submit() {
    if (!this.auth.canMutate) return;
    if (this.form.invalid) return;
    this.saving = true;
    this.statusMessage = '';
    this.api.createAccount(this.form.value as any).subscribe({
      next: () => {
        const defaultPortfolio = this.portfolios[0]?.id || '';
        this.form.reset({ type: 'cash', currency: 'VND', portfolioId: defaultPortfolio });
        this.saving = false;
        this.isCreating = false;
        this.statusMessage = 'Đã tạo tài khoản.';
        this.reload();
      },
      error: () => {
        this.saving = false;
        this.statusMessage = 'Không thể tạo tài khoản. Kiểm tra lại thông tin rồi thử lại.';
      },
    });
  }

  remove(item: { id: string; name: string }) {
    if (!this.auth.canMutate || !item.id) return;
    if (!window.confirm(`Xóa tài khoản “${item.name}”? Chỉ có thể xóa tài khoản không có giao dịch liên kết.`)) return;
    this.statusMessage = '';
    this.api.deleteAccount(item.id).subscribe({
      next: () => {
        this.statusMessage = `Đã xóa tài khoản “${item.name}”.`;
        this.reload();
      },
      error: (error) => {
        const code = error?.error?.code;
        this.statusMessage = code === 'ACCOUNT_HAS_TRANSACTIONS'
          ? 'Không thể xóa tài khoản vì đang có giao dịch hoặc dòng tiền liên kết.'
          : 'Không thể xóa tài khoản.';
      },
    });
  }

  typeLabel(type: string | undefined) {
    const labels: Record<string, string> = {
      bank: 'Ngân hàng',
      cash: 'Tiền mặt',
      card: 'Thẻ',
    };
    return labels[type || ''] || type || 'Khác';
  }
}
