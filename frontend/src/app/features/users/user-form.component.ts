import { Component, OnInit, inject, input, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { UsersService } from '../../core/api/users.service';
import { apiErrorMessage } from '../../core/api/api-error';
import { UserListItem } from '../../core/api/models';
import { ConfirmService } from '../../shared/confirm/confirm.service';
import { ToastService } from '../../shared/toast/toast.service';
import { IconComponent } from '../../shared/icon/icon.component';

@Component({
  selector: 'app-user-form',
  standalone: true,
  imports: [ReactiveFormsModule, IconComponent],
  templateUrl: './user-form.component.html',
  styleUrl: './user-form.component.scss',
})
export class UserFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly usersService = inject(UsersService);
  private readonly router = inject(Router);
  private readonly confirm = inject(ConfirmService);
  private readonly toast = inject(ToastService);

  /** Route param, bound via withComponentInputBinding; absent on /users/new */
  readonly id = input<string>();

  readonly isNew = signal(true);
  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly notFound = signal(false);
  readonly user = signal<UserListItem | null>(null);
  readonly showPassword = signal(false);

  readonly form = this.fb.nonNullable.group({
    userName: ['', [Validators.required, Validators.minLength(3)]],
    firstName: ['', [Validators.required]],
    lastName: ['', [Validators.required]],
    email: ['', [Validators.email]],
    phoneNumber: [''],
    password: ['', [Validators.required, Validators.minLength(6)]],
    isAdmin: [false],
  });

  ngOnInit(): void {
    const idParam = this.id();
    if (!idParam) {
      return;
    }

    this.isNew.set(false);
    this.form.controls.password.clearValidators();
    this.form.controls.password.setValidators([Validators.minLength(6)]);
    this.form.controls.password.updateValueAndValidity();

    const fromState = (history.state as { user?: UserListItem }).user;
    if (fromState && fromState.id === +idParam) {
      this.showUser(fromState);
      return;
    }

    this.loading.set(true);
    this.usersService.getById(+idParam).subscribe({
      next: (user) => {
        this.loading.set(false);
        if (user) {
          this.showUser(user);
        } else {
          this.notFound.set(true);
        }
      },
      error: (err) => {
        this.loading.set(false);
        this.notFound.set(true);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت اطلاعات کاربر'));
      },
    });
  }

  private showUser(user: UserListItem): void {
    this.user.set(user);
    this.form.patchValue({
      userName: user.userName,
      firstName: user.firstName,
      lastName: user.lastName,
      email: user.email ?? '',
      phoneNumber: user.phoneNumber ?? '',
    });
    this.form.patchValue({ isAdmin: user.isAdmin });
  }

  invalid(name: keyof typeof this.form.controls): boolean {
    const control = this.form.controls[name];
    return control.invalid && control.touched;
  }

  save(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const value = this.form.getRawValue();
    const password = this.isNew() || value.password ? value.password : '';

    this.saving.set(true);

    const request = this.isNew()
      ? this.usersService.create({
          id: 0,
          userName: value.userName.trim(),
          firstName: value.firstName.trim(),
          lastName: value.lastName.trim(),
          email: value.email.trim(),
          phoneNumber: value.phoneNumber.trim(),
          password: value.password,
          isAdmin: value.isAdmin,
        })
      : this.usersService.update({
          id: this.user()?.id ?? 0,
          userName: value.userName.trim(),
          firstName: value.firstName.trim(),
          lastName: value.lastName.trim(),
          email: value.email.trim(),
          phoneNumber: value.phoneNumber.trim(),
          password,
          isAdmin: value.isAdmin,
        });

    request.subscribe({
      next: () => {
        this.toast.success(this.isNew() ? 'کاربر با موفقیت ایجاد شد' : 'کاربر با موفقیت ویرایش شد');
        this.router.navigate(['/users']);
      },
      error: (err) => {
        this.saving.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در ذخیره کاربر'));
      },
    });
  }

  async deleteUser(): Promise<void> {
    const current = this.user();
    if (!current) {
      return;
    }

    const ok = await this.confirm.ask({
      title: 'حذف کاربر',
      message: `آیا از حذف کاربر «${current.firstName} ${current.lastName}» مطمئن هستید؟ این عملیات قابل بازگشت نیست.`,
      confirmText: 'حذف',
      danger: true,
    });

    if (!ok) {
      return;
    }

    this.usersService.delete([current.id]).subscribe({
      next: () => {
        this.toast.success('کاربر حذف شد');
        this.router.navigate(['/users']);
      },
      error: (err) => this.toast.error(apiErrorMessage(err, 'خطا در حذف کاربر')),
    });
  }

  back(): void {
    this.router.navigate(['/users']);
  }
}
