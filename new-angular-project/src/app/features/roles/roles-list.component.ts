import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { RolesService } from '../../core/api/roles.service';
import { apiErrorMessage } from '../../core/api/api-error';
import { RoleListItem } from '../../core/api/models';
import { ConfirmService } from '../../shared/confirm/confirm.service';
import { ToastService } from '../../shared/toast/toast.service';
import { ColumnDef, DataTableComponent, TableRow } from '../../shared/table/data-table.component';
import { IconComponent } from '../../shared/icon/icon.component';

@Component({
  selector: 'app-roles-list',
  standalone: true,
  imports: [DataTableComponent, IconComponent],
  templateUrl: './roles-list.component.html',
})
export class RolesListComponent {
  private readonly rolesService = inject(RolesService);
  private readonly router = inject(Router);
  private readonly confirm = inject(ConfirmService);
  private readonly toast = inject(ToastService);

  readonly columns: ColumnDef[] = [
    { key: 'title', header: 'عنوان نقش' },
    { key: 'code', header: 'کد نقش', ltr: true, width: '160px' },
  ];

  readonly rows = signal<RoleListItem[]>([]);
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
    this.rolesService.list(this.page(), this.pageSize()).subscribe({
      next: (rows) => {
        this.rows.set(rows ?? []);
        this.hasNext.set((rows ?? []).length === this.pageSize());
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت فهرست نقش‌ها'));
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
    this.router.navigate(['/roles', row.id], { state: { role: row } });
  }

  newRole(): void {
    this.router.navigate(['/roles/new']);
  }

  async deleteSelected(): Promise<void> {
    const ids = this.selectedIds();
    if (ids.length === 0) {
      return;
    }

    const ok = await this.confirm.ask({
      title: 'حذف نقش‌ها',
      message: `آیا از حذف ${ids.length} نقش انتخاب‌شده مطمئن هستید؟ این عملیات قابل بازگشت نیست.`,
      confirmText: 'حذف',
      danger: true,
    });

    if (!ok) {
      return;
    }

    this.deleting.set(true);
    this.rolesService.delete(ids).subscribe({
      next: () => {
        this.deleting.set(false);
        this.toast.success('نقش‌های انتخاب‌شده حذف شدند');
        this.load();
      },
      error: (err) => {
        this.deleting.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در حذف نقش‌ها'));
      },
    });
  }
}
