import { Component, OnInit, inject, input, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { RolesService } from '../../core/api/roles.service';
import { apiErrorMessage } from '../../core/api/api-error';
import { RoleListItem } from '../../core/api/models';
import { ConfirmService } from '../../shared/confirm/confirm.service';
import { ToastService } from '../../shared/toast/toast.service';
import { IconComponent } from '../../shared/icon/icon.component';

@Component({
  selector: 'app-role-form',
  standalone: true,
  imports: [ReactiveFormsModule, IconComponent],
  templateUrl: './role-form.component.html',
  styleUrl: './role-form.component.scss',
})
export class RoleFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly rolesService = inject(RolesService);
  private readonly router = inject(Router);
  private readonly confirm = inject(ConfirmService);
  private readonly toast = inject(ToastService);

  /** Route param, bound via withComponentInputBinding; absent on /roles/new */
  readonly id = input<string>();

  readonly isNew = signal(true);
  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly notFound = signal(false);
  readonly role = signal<RoleListItem | null>(null);

  readonly form = this.fb.nonNullable.group({
    title: ['', [Validators.required]],
    code: ['', [Validators.required]],
  });

  ngOnInit(): void {
    const idParam = this.id();
    if (!idParam) {
      return;
    }

    this.isNew.set(false);

    const fromState = (history.state as { role?: RoleListItem }).role;
    if (fromState && fromState.id === +idParam) {
      this.showRole(fromState);
      return;
    }

    this.loading.set(true);
    this.rolesService.getById(+idParam).subscribe({
      next: (role) => {
        this.loading.set(false);
        if (role) {
          this.showRole(role);
        } else {
          this.notFound.set(true);
        }
      },
      error: (err) => {
        this.loading.set(false);
        this.notFound.set(true);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت اطلاعات نقش'));
      },
    });
  }

  private showRole(role: RoleListItem): void {
    this.role.set(role);
    this.form.patchValue({ title: role.title, code: role.code });
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

    this.saving.set(true);

    if (this.isNew()) {
      this.rolesService
        .create({ id: 0, title: value.title.trim(), code: value.code.trim() })
        .subscribe({
          next: () => {
            this.toast.success('نقش با موفقیت ایجاد شد');
            this.router.navigate(['/roles']);
          },
          error: (err) => {
            this.saving.set(false);
            this.toast.error(apiErrorMessage(err, 'خطا در ذخیره نقش'));
          },
        });
      return;
    }

    const current = this.role();
    if (!current) {
      this.saving.set(false);
      return;
    }

    this.rolesService
      .update({ id: current.id, title: value.title.trim(), code: value.code.trim() })
      .subscribe({
        next: () => {
          this.toast.success('نقش با موفقیت ویرایش شد');
          this.router.navigate(['/roles']);
        },
        error: (err) => {
          this.saving.set(false);
          this.toast.error(apiErrorMessage(err, 'خطا در ذخیره نقش'));
        },
      });
  }

  async deleteRole(): Promise<void> {
    const current = this.role();
    if (!current) {
      return;
    }

    const ok = await this.confirm.ask({
      title: 'حذف نقش',
      message: `آیا از حذف نقش «${current.title}» مطمئن هستید؟ این عملیات قابل بازگشت نیست.`,
      confirmText: 'حذف',
      danger: true,
    });

    if (!ok) {
      return;
    }

    this.rolesService.delete([current.id]).subscribe({
      next: () => {
        this.toast.success('نقش حذف شد');
        this.router.navigate(['/roles']);
      },
      error: (err) => this.toast.error(apiErrorMessage(err, 'خطا در حذف نقش')),
    });
  }

  back(): void {
    this.router.navigate(['/roles']);
  }
}
