import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { BudgetPeriod } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { IconComponent } from '../../shared/icons/icon.component';
import { normalizeVndAmount } from '../../shared/money-input';

@Component({
  selector: 'app-budget-page',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, TranslatePipe, IconComponent],
  templateUrl: './budget-page.component.html',
})
export class BudgetPageComponent implements OnInit {
  form!: FormGroup;
  summary: BudgetPeriod | null = null;
  loading = true;
  saving = false;
  isCreating = false;
  statusMessage = '';
  loadError = '';

  constructor(
    private fb: FormBuilder,
    private api: ApiService,
    private route: ActivatedRoute,
    private router: Router,
    public auth: AuthService,
  ) {
    this.form = this.fb.group({
      period: [this.currentMonth()],
      categoryId: [''],
      limit: ['', [Validators.required]],
    });
  }

  ngOnInit() {
    const fromRoute = this.route.snapshot.paramMap.get('period');
    if (fromRoute) {
      this.form.patchValue({ period: fromRoute });
    }
    this.loadBudget();
  }

  currentMonth() {
    return new Date().toISOString().slice(0, 7);
  }

  loadBudget() {
    const period = (this.form.value.period || '').trim();
    if (!/^\d{4}-\d{2}$/.test(period)) {
      this.loadError = 'Chọn kỳ ngân sách hợp lệ theo tháng.';
      return;
    }
    this.loading = true;
    this.loadError = '';
    this.api.getBudget(period).subscribe({
      next: (res) => {
        this.summary = res;
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.summary = { period, workspaceId: '', asOfAt: '', rows: [] };
        this.loadError = 'Không thể tải ngân sách của kỳ này.';
      },
    });
  }

  submit() {
    if (!this.auth.canMutate || this.saving) return;
    if (this.form.invalid) return;
    const period = (this.form.value.period || '').trim();
    const payload = {
      period,
      categoryId: this.form.value.categoryId || 'uncategorized',
      limit: normalizeVndAmount(this.form.value.limit) || '0',
    };

    this.saving = true;
    this.api.upsertBudget(period, payload).subscribe({
      next: () => {
        this.statusMessage = 'Đã lưu giới hạn chi tiêu.';
        this.saving = false;
        this.isCreating = false;
        this.router.navigateByUrl(`/budgets/${encodeURIComponent(period)}`).then(() => this.loadBudget());
      },
      error: () => {
        this.saving = false;
        this.statusMessage = 'Không thể lưu giới hạn chi tiêu.';
      },
    });
  }

  openCreate() {
    if (!this.auth.canMutate) return;
    this.isCreating = true;
    this.statusMessage = '';
  }

  closeCreate() {
    if (!this.saving) this.isCreating = false;
  }
}
