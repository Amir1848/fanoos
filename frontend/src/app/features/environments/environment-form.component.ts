import { Component, OnInit, inject, input, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { EnvironmentsService } from '../../core/api/environments.service';
import { apiErrorMessage } from '../../core/api/api-error';
import { EnvironmentListItem } from '../../core/api/models';
import { ConfirmService } from '../../shared/confirm/confirm.service';
import { ToastService } from '../../shared/toast/toast.service';
import { IconComponent } from '../../shared/icon/icon.component';

@Component({
  selector: 'app-environment-form',
  standalone: true,
  imports: [ReactiveFormsModule, IconComponent],
  templateUrl: './environment-form.component.html',
  styleUrl: './environment-form.component.scss',
})
export class EnvironmentFormComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly environmentsService = inject(EnvironmentsService);
  private readonly router = inject(Router);
  private readonly confirm = inject(ConfirmService);
  private readonly toast = inject(ToastService);

  /** Route param, bound via withComponentInputBinding; absent on /environments/new */
  readonly id = input<string>();

  readonly isNew = signal(true);
  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly notFound = signal(false);
  readonly environment = signal<EnvironmentListItem | null>(null);

  readonly form = this.fb.nonNullable.group({
    name: ['', [Validators.required]],
    description: [''],
  });

  ngOnInit(): void {
    const idParam = this.id();
    if (!idParam) {
      return;
    }

    this.isNew.set(false);
    this.form.disable();

    const fromState = (history.state as { environment?: EnvironmentListItem }).environment;
    if (fromState && fromState.id === +idParam) {
      this.showEnvironment(fromState);
      return;
    }

    this.loading.set(true);
    this.environmentsService.getById(+idParam).subscribe({
      next: (environment) => {
        this.loading.set(false);
        if (environment) {
          this.showEnvironment(environment);
        } else {
          this.notFound.set(true);
        }
      },
      error: (err) => {
        this.loading.set(false);
        this.notFound.set(true);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت اطلاعات محیط'));
      },
    });
  }

  private showEnvironment(environment: EnvironmentListItem): void {
    this.environment.set(environment);
    this.form.patchValue({ name: environment.name, description: environment.description ?? '' });
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
    this.environmentsService
      .create({ id: 0, name: value.name.trim(), description: value.description.trim() })
      .subscribe({
        next: ({ id }) => {
          this.toast.success('محیط با موفقیت ایجاد شد');
          this.router.navigate(['/environments', id, 'editor']);
        },
        error: (err) => {
          this.saving.set(false);
          this.toast.error(apiErrorMessage(err, 'خطا در ذخیره محیط'));
        },
      });
  }

  openEditor(): void {
    const current = this.environment();
    if (!current) {
      return;
    }
    this.router.navigate(['/environments', current.id, 'editor'], { state: { environment: current } });
  }

  async deleteEnvironment(): Promise<void> {
    const current = this.environment();
    if (!current) {
      return;
    }

    const ok = await this.confirm.ask({
      title: 'حذف محیط',
      message: `آیا از حذف محیط «${current.name}» مطمئن هستید؟ نقشه دستگاه‌های مرتبط نیز حذف می‌شود.`,
      confirmText: 'حذف',
      danger: true,
    });

    if (!ok) {
      return;
    }

    this.environmentsService.delete([current.id]).subscribe({
      next: () => {
        this.toast.success('محیط حذف شد');
        this.router.navigate(['/environments']);
      },
      error: (err) => this.toast.error(apiErrorMessage(err, 'خطا در حذف محیط')),
    });
  }

  back(): void {
    this.router.navigate(['/environments']);
  }
}
