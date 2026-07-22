import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { ApiService } from '../../core/services/api.service';
import { Asset, AssetValuation } from '../../shared/models';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-asset-list',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './asset-list.component.html'
})
export class AssetListComponent implements OnInit {
  assets: Asset[] = [];
  valuations: Record<string, AssetValuation[]> = {};
  selectedAssetId: string | null = null;
  statusMessage = '';
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
    this.api.getAssets().subscribe({
      next: (items) => (this.assets = items),
      error: () => {
        this.statusMessage = 'Không lấy được danh sách tài sản.';
      },
    });
  }

  submitAsset() {
    if (!this.auth.canMutate) return;
    if (this.assetForm.invalid) return;
    this.api.createAsset({
      name: this.assetForm.value.name,
      assetType: this.assetForm.value.assetType,
      portfolioId: this.assetForm.value.portfolioId || '',
    }).subscribe({
      next: () => {
        this.statusMessage = 'Đã tạo tài sản.';
        this.assetForm.reset({ name: '', assetType: 'investment', portfolioId: '' });
        this.reload();
      },
      error: () => {
        this.statusMessage = 'Không thể tạo tài sản.';
      },
    });
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
    this.api
      .addAssetValuation(this.selectedAssetId, {
        valuationAmount: this.valuationForm.value.valuationAmount || '',
        currency: this.valuationForm.value.currency || 'VND',
        source: this.valuationForm.value.source || 'self_declared',
        effectiveAt: this.valuationForm.value.effectiveAt || undefined,
      })
      .subscribe({
        next: () => {
          this.statusMessage = 'Đã thêm định giá tài sản.';
          this.selectedAssetId = null;
        },
        error: () => {
          this.statusMessage = 'Không thể thêm định giá.';
        },
      });
  }

  closeValuation() {
    this.selectedAssetId = null;
    this.statusMessage = '';
  }
}
