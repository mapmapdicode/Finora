import { Component } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

import { CommonModule } from '@angular/common';
import { IconComponent } from '../../shared/icons/icon.component';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink, IconComponent],
  templateUrl: './register.component.html',
})
export class RegisterComponent {
  form!: FormGroup;
  message = '';

  constructor(
    private fb: FormBuilder,
    private auth: AuthService,
    private router: Router
  ) {
    this.form = this.fb.group({
      email: ['thanhoangz', [Validators.required]],
      password: ['HoangThanZ6^', Validators.required],
      name: ['Than Hoang Z', Validators.required],
    });
  }

  submit() {
    if (this.form.invalid) return;
    const { email, password, name } = this.form.value;
    this.auth.register({ email, password, name }).subscribe({
      next: (res) => {
        this.auth.persistSession(res);
        this.router.navigateByUrl('/dashboard');
      },
      error: (error) => {
        this.message = error?.error?.message || 'Đăng ký không thành công';
      },
    });
  }
}
