import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { AuthService } from '../../core/auth/auth.service';
import { IconComponent } from '../../shared/icon/icon.component';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [ReactiveFormsModule, IconComponent],
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss',
})
export class LoginComponent {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  readonly loading = signal(false);
  readonly errorMessage = signal('');
  readonly showPassword = signal(false);

  readonly form = this.fb.nonNullable.group({
    username: ['', [Validators.required]],
    password: ['', [Validators.required]],
  });

  submit(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const { username, password } = this.form.getRawValue();

    this.loading.set(true);
    this.errorMessage.set('');

    this.auth.login(username.trim(), password).subscribe({
      next: () => this.router.navigate(['/']),
      error: () => {
        this.loading.set(false);
        this.errorMessage.set('نام کاربری یا رمز عبور نادرست است');
      },
    });
  }

  invalid(name: 'username' | 'password'): boolean {
    const control = this.form.controls[name];
    return control.invalid && control.touched;
  }
}
