import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { BudgetPeriod } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-budget-page',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './budget-page.component.html',
})
export class BudgetPageComponent implements OnInit {
  form!: FormGroup;
  summary: BudgetPeriod | null = null;
  loading = false;
  statusMessage = '';

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
    if (!period) return;
    this.loading = true;
    this.statusMessage = 'Đang tải budget...';
    this.api.getBudget(period).subscribe({
      next: (res) => {
        this.summary = res;
        this.loading = false;
        this.statusMessage = '';
      },
      error: () => {
        this.loading = false;
        this.summary = { period, workspaceId: '', asOfAt: '', rows: [] };
        this.statusMessage = 'Không lấy được budget của kỳ này.';
      },
    });
  }

  submit() {
    if (!this.auth.canMutate) return;
    if (this.form.invalid) return;
    const period = (this.form.value.period || '').trim();
    const payload = {
      period,
      categoryId: this.form.value.categoryId || 'uncategorized',
      limit: this.form.value.limit || '0',
    };

    this.api.upsertBudget(period, payload).subscribe({
      next: () => {
        this.statusMessage = 'Đã lưu giới hạn chi tiêu.';
        this.router.navigateByUrl(`/budgets/${encodeURIComponent(period)}`).then(() => this.loadBudget());
      },
      error: () => {
        this.statusMessage = 'Lưu budget thất bại.';
      },
    });
  }
}
