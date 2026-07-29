import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { ApiService } from '../../core/services/api.service';
import { Asset, AssetValuation } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';
import { IconComponent } from '../../shared/icons/icon.component';
import { normalizeVndAmount } from '../../shared/money-input';

import { TranslatePipe } from '../../shared/pipes/translate.pipe';

@Component({
  selector: 'app-asset-list',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, IconComponent, TranslatePipe],
  templateUrl: './asset-list.component.html'
})
export class AssetListComponent implements OnInit {
  assets: Asset[] = [];
  valuations: Record<string, AssetValuation[]> = {};
  selectedAssetId: string | null = null;
  statusMessage = '';
  assetsLoading = true;
  isCreating = false;
  assetSaving = false;
  valuationSaving = false;
  assetForm: FormGroup;
  valuationForm: FormGroup;

  constructor(private api: ApiService, private fb: FormBuilder, public auth: AuthService) {
    this.assetForm = this.fb.group({
      name: ['', Validators.required],
      assetType: ['investment', Validators.required],
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
    this.assetsLoading = true;
    this.statusMessage = '';
    this.api.getAssets().subscribe({
      next: (items) => {
        this.assets = items;
        this.assetsLoading = false;
      },
      error: () => {
        this.assetsLoading = false;
        this.statusMessage = 'Không lấy được danh sách tài sản.';
      },
    });
  }

  submitAsset() {
    if (!this.auth.canMutate) return;
    if (this.assetForm.invalid) return;
    this.assetSaving = true;
    this.api.createAsset({
      name: this.assetForm.value.name,
      assetType: this.assetForm.value.assetType,
      portfolioId: this.assetForm.value.portfolioId || '',
    }).subscribe({
      next: () => {
        this.statusMessage = 'Đã tạo tài sản.';
        this.assetForm.reset({ name: '', assetType: 'investment', portfolioId: '' });
        this.assetSaving = false;
        this.isCreating = false;
        this.reload();
      },
      error: () => {
        this.assetSaving = false;
        this.statusMessage = 'Không thể tạo tài sản.';
      },
    });
  }

  openCreate() {
    if (!this.auth.canMutate) return;
    this.isCreating = true;
    this.statusMessage = '';
  }

  closeCreate() {
    if (!this.assetSaving) this.isCreating = false;
  }

  openValuation(assetId: string) {
    if (!this.auth.canMutate) return;
    this.selectedAssetId = assetId;
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
    if (!this.selectedAssetId || this.valuationForm.invalid) return;
    this.valuationSaving = true;
    this.api
      .addAssetValuation(this.selectedAssetId, {
        valuationAmount: normalizeVndAmount(this.valuationForm.value.valuationAmount),
        currency: this.valuationForm.value.currency || 'VND',
        source: this.valuationForm.value.source || 'self_declared',
        effectiveAt: this.valuationForm.value.effectiveAt || undefined,
      })
      .subscribe({
        next: () => {
          this.statusMessage = 'Đã thêm định giá tài sản.';
          this.selectedAssetId = null;
          this.valuationSaving = false;
        },
        error: () => {
          this.valuationSaving = false;
          this.statusMessage = 'Không thể thêm định giá.';
        },
      });
  }

  closeValuation() {
    this.selectedAssetId = null;
    this.statusMessage = '';
  }

  assetTypeLabel(assetType: string | undefined) {
    const labels: Record<string, string> = {
      investment: 'Đầu tư',
      vehicle: 'Phương tiện',
      collectible: 'Đồ sưu tầm',
      other: 'Tài sản khác',
    };
    return labels[assetType || ''] || assetType || 'Chưa xác định';
  }
}
