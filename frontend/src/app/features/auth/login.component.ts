import { Component } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

import { CommonModule } from '@angular/common';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, IconComponent],
  templateUrl: './login.component.html',
})
export class LoginComponent {
  form!: FormGroup;
  message = '';

  constructor(
    private fb: FormBuilder,
    private auth: AuthService,
    private router: Router
  ) {
    this.form = this.fb.group({
      email: ['demo@wealthos.vn', [Validators.required, Validators.email]],
      password: ['demo-pass', Validators.required],
    });
  }

  submit() {
    if (this.form.invalid) return;
    this.auth.login(this.form.value.email || '', this.form.value.password || '').subscribe({
      next: (res) => {
        this.auth.persistSession(res);
        this.router.navigateByUrl('/dashboard');
      },
      error: () => {
        this.message = 'Đăng nhập không thành công';
      },
    });
  }
}
