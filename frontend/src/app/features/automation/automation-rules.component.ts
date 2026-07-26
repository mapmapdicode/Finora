import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { AuthService } from '../../core/services/auth.service';
import { BankAutomationRule } from '../../shared/models';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { IconComponent } from '../../shared/icons/icon.component';

type Direction = 'in' | 'out';

@Component({
  selector: 'app-automation-rules',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, TranslatePipe, IconComponent],
  templateUrl: './automation-rules.component.html',
})
export class AutomationRulesComponent implements OnInit {
  rules: BankAutomationRule[] = [];
  editingId: string | null = null;
  statusMessage = '';
  previewResult: unknown = null;
  form: FormGroup;

  constructor(
    private api: ApiService,
    private fb: FormBuilder,
    public auth: AuthService,
  ) {
    this.form = this.fb.group({
      name: ['', Validators.required],
      accountId: [''],
      direction: ['out', Validators.required],
      actionType: ['categorize', Validators.required],
      type: ['expense', Validators.required],
      priority: [10, [Validators.required, Validators.min(0)]],
      enabled: [true],
      categoryId: [''],
      contentPattern: [''],
      referencePattern: [''],
      minAmount: [''],
      maxAmount: [''],
      predicate: [''],
    });
  }

  ngOnInit() {
    this.reload();
  }

  reload() {
    this.api.listAutomationRules().subscribe({
      next: (items) => {
        this.rules = items;
      },
      error: () => {
        this.statusMessage = 'Không thể tải danh sách quy tắc.';
      },
    });
  }

  submitRule() {
    if (!this.auth.canMutate) return;
    if (this.form.invalid) return;
    const payload = this.form.value;
    if (!payload.contentPattern && !payload.referencePattern && !payload.predicate) {
      this.statusMessage = 'Cần ít nhất một điều kiện khớp nội dung hoặc tham chiếu.';
      return;
    }
    const body = {
      accountId: payload.accountId || '',
      name: payload.name,
      direction: payload.direction,
      actionType: payload.actionType,
      type: payload.type,
      predicate: payload.predicate || '',
      categoryId: payload.categoryId || '',
      priority: Number(payload.priority || 0),
      enabled: !!payload.enabled,
      contentPattern: payload.contentPattern || '',
      referencePattern: payload.referencePattern || '',
      minAmount: payload.minAmount || '',
      maxAmount: payload.maxAmount || '',
    };

    if (this.editingId) {
      this.api.updateAutomationRule(this.editingId, body).subscribe({
        next: () => {
          this.statusMessage = 'Đã cập nhật quy tắc.';
          this.editingId = null;
          this.form.reset({
            name: '',
            accountId: '',
            direction: 'out',
            actionType: 'categorize',
            type: 'expense',
            priority: 10,
            enabled: true,
            categoryId: '',
            contentPattern: '',
            referencePattern: '',
            minAmount: '',
            maxAmount: '',
            predicate: '',
          });
          this.reload();
        },
        error: () => {
          this.statusMessage = 'Cập nhật quy tắc thất bại.';
        },
      });
      return;
    }

    this.api.createAutomationRule(body).subscribe({
      next: () => {
        this.statusMessage = 'Đã tạo quy tắc mới.';
        this.form.reset({
          name: '',
          accountId: '',
          direction: 'out',
          actionType: 'categorize',
          type: 'expense',
          priority: 10,
          enabled: true,
          categoryId: '',
          contentPattern: '',
          referencePattern: '',
          minAmount: '',
          maxAmount: '',
          predicate: '',
        });
        this.reload();
      },
      error: () => {
        this.statusMessage = 'Tạo quy tắc thất bại.';
      },
    });
  }

  beginEdit(rule: BankAutomationRule) {
    if (!this.auth.canMutate) return;
    this.editingId = rule.id;
    this.previewResult = null;
    this.form.patchValue({
      name: rule.name,
      accountId: rule.accountId || '',
      direction: (rule.direction || 'out') as Direction,
      actionType: rule.actionType || 'categorize',
      type: rule.type || 'expense',
      priority: rule.priority || 10,
      enabled: rule.enabled,
      categoryId: rule.categoryId || '',
      contentPattern: rule.contentPattern || '',
      referencePattern: rule.referencePattern || '',
      minAmount: rule.minAmount || '',
      maxAmount: rule.maxAmount || '',
      predicate: rule.predicate || '',
    });
  }

  cancelEdit() {
    this.editingId = null;
    this.form.reset({
      name: '',
      accountId: '',
      direction: 'out',
      actionType: 'categorize',
      type: 'expense',
      priority: 10,
      enabled: true,
      categoryId: '',
      contentPattern: '',
      referencePattern: '',
      minAmount: '',
      maxAmount: '',
      predicate: '',
    });
  }

  toggleEnabled(rule: BankAutomationRule) {
    if (!this.auth.canMutate) return;
    this.api.updateAutomationRule(rule.id, { enabled: !rule.enabled }).subscribe(() => this.reload());
  }

  remove(ruleId: string) {
    if (!this.auth.canMutate) return;
    this.api.deleteAutomationRule(ruleId).subscribe({
      next: () => {
        this.statusMessage = 'Đã xoá quy tắc.';
        this.reload();
      },
      error: () => {
        this.statusMessage = 'Không thể xoá quy tắc.';
      },
    });
  }

  preview() {
    this.api.previewRule({ sample: 20 }).subscribe({
      next: (result) => {
        this.previewResult = result;
      },
      error: () => {
        this.statusMessage = 'Không thể preview rule.';
      },
    });
  }

  directionLabel(direction: string) {
    return direction === 'in' ? 'Tiền vào' : 'Tiền ra';
  }
}
