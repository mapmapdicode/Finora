import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { Property } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-property-list',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, TranslatePipe, IconComponent],
  templateUrl: './property-list.component.html',
})
export class PropertyListComponent implements OnInit {
  properties: Property[] = [];
  statusMessage = '';
  propertyForm: FormGroup;
  valuationForm: FormGroup;
  selectedPropertyId: string | null = null;

  constructor(private api: ApiService, private fb: FormBuilder, public auth: AuthService) {
    this.propertyForm = this.fb.group({
      name: ['', Validators.required],
      address: ['', Validators.required],
      areaM2: ['0', Validators.required],
      portfolioId: [''],
    });
    this.valuationForm = this.fb.group({
      valuationAmount: ['', Validators.required],
      currency: ['VND', Validators.required],
      source: ['self_declared', Validators.required],
      effectiveAt: [''],
    });
  }

  ngOnInit() {
    this.reload();
  }

  reload() {
    this.api.listProperties().subscribe({
      next: (items) => {
        this.properties = items;
      },
      error: () => {
        this.statusMessage = 'Không lấy được danh sách tài sản bất động sản.';
      },
    });
  }

  submitProperty() {
    if (!this.auth.canMutate) return;
    if (this.propertyForm.invalid) return;
    const payload = {
      name: this.propertyForm.value.name,
      address: this.propertyForm.value.address,
      areaM2: this.propertyForm.value.areaM2,
      portfolioId: this.propertyForm.value.portfolioId || '',
    };
    this.api.createProperty(payload).subscribe({
      next: () => {
        this.propertyForm.reset({ name: '', address: '', areaM2: '0', portfolioId: '' });
        this.reload();
      },
      error: () => {
        this.statusMessage = 'Không thể tạo tài sản bất động sản.';
      },
    });
  }

  openValuation(propertyId: string) {
    if (!this.auth.canMutate) return;
    this.selectedPropertyId = propertyId;
    this.valuationForm.reset({
      valuationAmount: '',
      currency: 'VND',
      source: 'self_declared',
      effectiveAt: '',
    });
    this.statusMessage = '';
  }

  submitValuation() {
    if (!this.auth.canMutate) return;
    if (!this.selectedPropertyId || this.valuationForm.invalid) return;
    this.api
      .addPropertyValuation(this.selectedPropertyId, {
        valuationAmount: this.valuationForm.value.valuationAmount || '',
        currency: this.valuationForm.value.currency || 'VND',
        source: this.valuationForm.value.source || 'self_declared',
        effectiveAt: this.valuationForm.value.effectiveAt || undefined,
      })
      .subscribe({
        next: () => {
          this.statusMessage = 'Đã thêm định giá tài sản.';
          this.selectedPropertyId = null;
        },
        error: () => {
          this.statusMessage = 'Không thể thêm định giá.';
        },
      });
  }

  closeValuation() {
    this.selectedPropertyId = null;
  }
}
