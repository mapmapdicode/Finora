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

  constructor(private fb: FormBuilder, private api: ApiService, public auth: AuthService) {
    this.form = this.fb.group({
      name: ['Kế hoạch cá nhân', Validators.required],
      assumptions: ['{}', Validators.required],
    });
  }

  ngOnInit() {
    this.reloadScenarios();
  }

  reloadScenarios() {
    this.api.getForecastScenarios().subscribe({
      next: (items) => {
        this.scenarios = items;
      },
      error: () => {
        this.statusMessage = 'Không lấy được danh sách kịch bản.';
      },
    });
  }

  submit() {
    if (!this.auth.canMutate) return;
    if (this.form.invalid) return;
    this.api
      .createForecastScenario(this.form.value as { name: string; assumptions: string })
      .subscribe({
        next: (res) => {
          this.lastResult = res;
          this.statusMessage = 'Đã tạo kịch bản. Chọn run để xem kết quả chi tiết.';
          this.reloadScenarios();
        },
      });
  }

  runScenario(item: ForecastScenario) {
    if (!this.auth.canMutate) return;
    this.activeScenarioId = item.id;
    this.api.runForecastScenario(item.id).subscribe({
      next: (res) => {
        this.lastResult = res;
        this.statusMessage = `Đã chạy kịch bản ${item.name}`;
        this.reloadScenarios();
      },
      error: () => {
        this.statusMessage = `Chạy kịch bản ${item.name} thất bại.`;
      },
    });
  }
}
