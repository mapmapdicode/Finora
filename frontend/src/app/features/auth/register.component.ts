import { Component } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [ReactiveFormsModule, RouterLink],
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
      email: ['demo@wealthos.vn', [Validators.required, Validators.email]],
      password: ['demo-pass', Validators.required],
      name: ['Demo User', Validators.required],
      workspaceName: ['Cá nhân', Validators.required],
    });
  }

  submit() {
    if (this.form.invalid) return;
    const { email, password, name, workspaceName } = this.form.value;
    this.auth.register({ email, password, name, workspaceName }).subscribe({
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
