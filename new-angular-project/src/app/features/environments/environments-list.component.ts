import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { EnvironmentsService } from '../../core/api/environments.service';
import { apiErrorMessage } from '../../core/api/api-error';
import { EnvironmentListItem } from '../../core/api/models';
import { ConfirmService } from '../../shared/confirm/confirm.service';
import { ToastService } from '../../shared/toast/toast.service';
import { ColumnDef, DataTableComponent, TableRow } from '../../shared/table/data-table.component';
import { IconComponent } from '../../shared/icon/icon.component';

@Component({
  selector: 'app-environments-list',
  standalone: true,
  imports: [DataTableComponent, IconComponent],
  templateUrl: './environments-list.component.html',
})
export class EnvironmentsListComponent {
  private readonly environmentsService = inject(EnvironmentsService);
  private readonly router = inject(Router);
  private readonly confirm = inject(ConfirmService);
  private readonly toast = inject(ToastService);

  readonly columns: ColumnDef[] = [
    { key: 'name', header: 'نام محیط' },
    { key: 'description', header: 'توضیحات' },
  ];

  readonly rows = signal<EnvironmentListItem[]>([]);
  readonly loading = signal(false);
  readonly deleting = signal(false);
  readonly page = signal(1);
  readonly pageSize = signal(10);
  readonly hasNext = signal(false);
  readonly selectedIds = signal<number[]>([]);

  constructor() {
    this.load();
  }

  load(): void {
    this.loading.set(true);
    this.environmentsService.list(this.page(), this.pageSize()).subscribe({
      next: (rows) => {
        this.rows.set(rows ?? []);
        this.hasNext.set((rows ?? []).length === this.pageSize());
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت فهرست محیط‌ها'));
      },
    });
  }

  onPageChange(page: number): void {
    this.page.set(page);
    this.load();
  }

  onPageSizeChange(size: number): void {
    this.pageSize.set(size);
    this.page.set(1);
    this.load();
  }

  openRow(row: TableRow): void {
    this.router.navigate(['/environments', row.id], { state: { environment: row } });
  }

  newEnvironment(): void {
    this.router.navigate(['/environments/new']);
  }

  openEditor(row: TableRow): void {
    this.router.navigate(['/environments', row.id, 'editor'], { state: { environment: row } });
  }

  async deleteSelected(): Promise<void> {
    const ids = this.selectedIds();
    if (ids.length === 0) {
      return;
    }

    const ok = await this.confirm.ask({
      title: 'حذف محیط‌ها',
      message: `آیا از حذف ${ids.length} محیط انتخاب‌شده مطمئن هستید؟ نقشه دستگاه‌های مرتبط نیز حذف می‌شود.`,
      confirmText: 'حذف',
      danger: true,
    });

    if (!ok) {
      return;
    }

    this.deleting.set(true);
    this.environmentsService.delete(ids).subscribe({
      next: () => {
        this.deleting.set(false);
        this.toast.success('محیط‌های انتخاب‌شده حذف شدند');
        this.load();
      },
      error: (err) => {
        this.deleting.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در حذف محیط‌ها'));
      },
    });
  }
}
