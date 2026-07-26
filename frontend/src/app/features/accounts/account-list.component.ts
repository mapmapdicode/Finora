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

  constructor(private fb: FormBuilder, private api: ApiService, public auth: AuthService) {}

  ngOnInit() {
    this.form = this.fb.group({
      name: ['', Validators.required],
      type: ['cash', Validators.required],
      currency: ['VND', Validators.required],
      portfolioId: [''],
    });
    this.api.getPortfolios().subscribe((items) => {
      this.portfolios = items;
      const defaultPortfolio = items[0]?.id || '';
      this.form.patchValue({ portfolioId: defaultPortfolio });
    });
    this.reload();
  }

  private reload() {
    this.api.getAccounts().subscribe((items) => this.accounts = items);
  }

  submit() {
    if (!this.auth.canMutate) return;
    if (this.form.invalid) return;
    this.api.createAccount(this.form.value as any).subscribe(() => {
      const defaultPortfolio = this.portfolios[0]?.id || '';
      this.form.reset({ type: 'cash', currency: 'VND', portfolioId: defaultPortfolio });
      this.reload();
    });
  }
}
