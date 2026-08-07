import { Component, OnDestroy, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';
import { CommonModule } from '@angular/common';

import { LanguageService, SupportedLanguage } from '../../core/services/language.service';

export interface SavedUser {
  name: string;
  email: string;
  avatar?: string;
}

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './login.component.html',
})
export class LoginComponent implements OnInit, OnDestroy {
  form!: FormGroup;
  message = '';
  verificationRequired = false;
  submitting = false;

  hasSavedUser = false;
  lastUser: SavedUser | null = null;

  // UI state management for glassmorphism layout & fast auth
  authMode: 'faceid' | 'password' = 'password';
  showPassword = false;
  isScanningFace = false;
  
  // Quick Action Modals
  showOtpModal = false;
  showQrModal = false;
  showSupportModal = false;
  
  // OTP state
  otpCode = '849 204';
  otpSeconds = 28;
  otpCopied = false;
  private timer: any;

  constructor(
    private fb: FormBuilder,
    private auth: AuthService,
    private router: Router,
    public langService: LanguageService
  ) {
    this.form = this.fb.group({
      email: ['', [Validators.required]],
      password: ['', Validators.required],
      code: [''],
    });
  }

  ngOnInit(): void {
    const rawSaved = localStorage.getItem('finora.last_user');
    if (rawSaved) {
      try {
        const parsed = JSON.parse(rawSaved);
        if (parsed && (parsed.email || parsed.name)) {
          this.lastUser = parsed;
          this.hasSavedUser = true;
          this.authMode = 'faceid';
          this.form.patchValue({
            email: parsed.email || '',
          });
        }
      } catch {
        this.hasSavedUser = false;
      }
    }

    if (!this.hasSavedUser) {
      this.authMode = 'password';
    }

    this.timer = setInterval(() => {
      if (this.otpSeconds > 1) {
        this.otpSeconds--;
      } else {
        this.otpSeconds = 30;
        this.otpCode = Math.floor(100000 + Math.random() * 900000).toString().replace(/(\d{3})(\d{3})/, '$1 $2');
      }
    }, 1000);
  }

  ngOnDestroy(): void {
    if (this.timer) {
      clearInterval(this.timer);
    }
  }

  toggleAuthMode(mode?: 'faceid' | 'password'): void {
    this.message = '';
    if (mode) {
      this.authMode = mode;
    } else {
      this.authMode = this.authMode === 'faceid' ? 'password' : 'faceid';
    }
  }

  switchAccount(): void {
    this.hasSavedUser = false;
    this.lastUser = null;
    this.authMode = 'password';
    this.form.patchValue({
      email: '',
      password: '',
    });
  }

  loginWithFaceID(): void {
    if (this.submitting || this.isScanningFace) return;
    
    this.isScanningFace = true;
    this.message = '';
    
    setTimeout(() => {
      this.isScanningFace = false;
      this.submit();
    }, 1200);
  }

  submit(): void {
    if (this.form.invalid) {
      this.authMode = 'password';
      this.message = 'Vui lòng nhập đầy đủ tên đăng nhập / email và mật khẩu';
      return;
    }
    this.message = '';
    this.submitting = true;
    this.auth.login(this.form.value.email || '', this.form.value.password || '').subscribe({
      next: (res) => {
        this.submitting = false;
        this.auth.persistSession(res);

        // Save last logged in user for fast login next time
        const userName = res.user?.name || this.form.value.email || 'Người dùng Finora';
        const userEmail = res.user?.email || this.form.value.email || '';
        const savedUser: SavedUser = {
          name: userName,
          email: userEmail,
        };
        localStorage.setItem('finora.last_user', JSON.stringify(savedUser));

        this.router.navigateByUrl('/dashboard');
      },
      error: (err) => {
        this.submitting = false;
        if (err?.error?.code === 'EMAIL_NOT_VERIFIED') {
          this.verificationRequired = true;
          this.message = 'Email chưa được xác thực. Nhập mã 6 số đã gửi tới email của bạn.';
          return;
        }
        this.authMode = 'password';
        this.message = err?.error?.message || 'Đăng nhập không thành công. Vui lòng kiểm tra lại thông tin.';
      },
    });
  }

  verifyEmail(): void {
    const email = this.form.value.email || '';
    const code = this.form.value.code || '';
    if (!email || !/^\d{6}$/.test(code)) {
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
      error: (err) => {
        this.submitting = false;
        this.message = err?.error?.message || 'Mã xác thực không hợp lệ hoặc đã hết hạn.';
      },
    });
  }

  resendCode(): void {
    const email = this.form.value.email || '';
    if (!email) return;
    this.auth.resendVerificationEmail(email).subscribe({
      next: () => (this.message = 'Nếu tài khoản cần xác thực, mã mới đã được gửi.'),
      error: () => (this.message = 'Không thể gửi lại mã ngay bây giờ.'),
    });
  }

  openOtpModal(): void {
    this.showOtpModal = true;
  }

  openQrModal(): void {
    this.showQrModal = true;
  }

  openSupportModal(): void {
    this.showSupportModal = true;
  }

  closeModals(): void {
    this.showOtpModal = false;
    this.showQrModal = false;
    this.showSupportModal = false;
  }

  copyOtp(): void {
    navigator.clipboard?.writeText(this.otpCode.replace(/\s+/g, ''));
    this.otpCopied = true;
    setTimeout(() => {
      this.otpCopied = false;
    }, 2000);
  }

  simulateQrSuccess(): void {
    this.closeModals();
    this.loginWithFaceID();
  }

  changeLanguage(event: Event): void {
    const lang = (event.target as HTMLSelectElement).value as SupportedLanguage;
    this.langService.setLanguage(lang);
  }
}
