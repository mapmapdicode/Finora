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
  verificationRequired = false;
  submitting = false;

  constructor(
    private fb: FormBuilder,
    private auth: AuthService,
    private router: Router
  ) {
    this.form = this.fb.group({
      email: ['thanhoangz', [Validators.required]],
      password: ['HoangThanZ6^', Validators.required],
      confirmPassword: ['HoangThanZ6^', Validators.required],
      name: ['Than Hoang Z', Validators.required],
      code: [''],
    });
  }

  submit() {
    if (this.form.invalid || this.submitting) return;
    const { email, password, confirmPassword, name } = this.form.value;
    if (password !== confirmPassword) {
      this.message = 'Mật khẩu xác nhận chưa khớp.';
      return;
    }
    this.submitting = true;
    this.auth.register({ email, password, confirmPassword, name }).subscribe({
      next: () => {
        this.submitting = false;
        this.verificationRequired = true;
        this.message = 'Tài khoản đã được tạo. Nhập mã 6 số được gửi tới email để hoàn tất.';
      },
      error: (error) => {
        this.submitting = false;
        this.message = error?.error?.message || 'Đăng ký không thành công';
      },
    });
  }

  verifyEmail() {
    const email = this.form.value.email || '';
    const code = this.form.value.code || '';
    if (!/^\d{6}$/.test(code)) {
      this.message = 'Vui lòng nhập mã xác thực gồm 6 chữ số.';
      return;
    }
    this.submitting = true;
    this.auth.verifyEmail(email, code).subscribe({
      next: (res) => {
        this.submitting = false;
        this.auth.persistSession(res);
        this.router.navigateByUrl('/dashboard');
      },
      error: (error) => {
        this.submitting = false;
        this.message = error?.error?.message || 'Mã xác thực không hợp lệ hoặc đã hết hạn.';
      },
    });
  }

  resendCode() {
    this.auth.resendVerificationEmail(this.form.value.email || '').subscribe({
      next: () => (this.message = 'Nếu tài khoản cần xác thực, mã mới đã được gửi.'),
      error: () => (this.message = 'Không thể gửi lại mã ngay bây giờ.'),
    });
  }
}
