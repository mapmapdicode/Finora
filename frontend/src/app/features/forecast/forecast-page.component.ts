import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { ForecastScenario } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-forecast-page',
  standalone: true,
  imports: [ReactiveFormsModule, CommonModule, TranslatePipe, IconComponent],
  templateUrl: './forecast-page.component.html'
})
export class ForecastPageComponent implements OnInit {
  scenarios: ForecastScenario[] = [];
  form: FormGroup;
  lastResult: unknown = null;
  activeScenarioId = '';
  statusMessage = '';
  scenariosLoading = true;
  isCreating = false;
  createInProgress = false;

  constructor(private fb: FormBuilder, private api: ApiService, public auth: AuthService) {
    this.form = this.fb.group({
      name: ['', Validators.required],
      assumptions: ['{}', Validators.required],
    });
  }

  ngOnInit() {
    this.reloadScenarios();
  }

  reloadScenarios() {
    this.scenariosLoading = true;
    this.api.getForecastScenarios().subscribe({
      next: (items) => {
        this.scenarios = items;
        this.scenariosLoading = false;
      },
      error: () => {
        this.scenariosLoading = false;
        this.statusMessage = 'Không lấy được danh sách kịch bản.';
      },
    });
  }

  submit() {
    if (!this.auth.canMutate || this.createInProgress) return;
    if (this.form.invalid) return;
    try {
      JSON.parse(this.form.value.assumptions || '{}');
    } catch (_error) {
      this.statusMessage = 'Giả định cần ở định dạng JSON hợp lệ.';
      return;
    }
    this.createInProgress = true;
    this.api
      .createForecastScenario(this.form.value as { name: string; assumptions: string })
      .subscribe({
        next: (res) => {
          this.lastResult = res;
          this.statusMessage = 'Đã tạo kịch bản. Chọn run để xem kết quả chi tiết.';
          this.createInProgress = false;
          this.isCreating = false;
          this.form.reset({ name: '', assumptions: '{}' });
          this.reloadScenarios();
        },
        error: () => {
          this.createInProgress = false;
          this.statusMessage = 'Không thể tạo kịch bản.';
        },
      });
  }

  openCreate() {
    if (!this.auth.canMutate) return;
    this.isCreating = true;
    this.statusMessage = '';
  }

  closeCreate() {
    if (!this.createInProgress) this.isCreating = false;
  }

  runScenario(item: ForecastScenario) {
    if (!this.auth.canMutate) return;
    this.activeScenarioId = item.id;
    this.api.runForecastScenario(item.id).subscribe({
      next: (res) => {
        this.lastResult = res;
        this.statusMessage = `Đã chạy kịch bản ${item.name}`;
        this.activeScenarioId = '';
        this.reloadScenarios();
      },
      error: () => {
        this.activeScenarioId = '';
        this.statusMessage = `Chạy kịch bản ${item.name} thất bại.`;
      },
    });
  }

  statusLabel(status?: string) {
    if (status === 'done') return 'Hoàn tất';
    if (status === 'running') return 'Đang chạy';
    return 'Chờ chạy';
  }
}
