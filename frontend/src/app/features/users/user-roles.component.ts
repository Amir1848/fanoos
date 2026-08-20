import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { forkJoin } from 'rxjs';
import { UsersService } from '../../core/api/users.service';
import { UserRolesService } from '../../core/api/user-roles.service';
import { apiErrorMessage } from '../../core/api/api-error';
import { RoleListItem, UserListItem } from '../../core/api/models';
import { ToastService } from '../../shared/toast/toast.service';
import { IconComponent } from '../../shared/icon/icon.component';

/** Users and roles are small catalogs in this system — one page each covers all. */
const PAGE_SIZE = 500;

@Component({
  selector: 'app-user-roles',
  standalone: true,
  imports: [IconComponent],
  templateUrl: './user-roles.component.html',
  styleUrl: './user-roles.component.scss',
})
export class UserRolesComponent implements OnInit {
  private readonly usersService = inject(UsersService);
  private readonly userRolesService = inject(UserRolesService);
  private readonly toast = inject(ToastService);

  readonly loading = signal(true);
  readonly rolesLoading = signal(false);
  readonly saving = signal(false);

  readonly users = signal<UserListItem[]>([]);
  readonly roles = signal<RoleListItem[]>([]);
  readonly selectedUser = signal<UserListItem | null>(null);
  readonly assignedRoleIds = signal<ReadonlySet<number>>(new Set());
  readonly selectedRoleIds = signal<ReadonlySet<number>>(new Set());

  readonly dirty = computed(() => {
    if (!this.selectedUser()) {
      return false;
    }
    const before = this.assignedRoleIds();
    const after = this.selectedRoleIds();
    return before.size !== after.size || [...before].some((id) => !after.has(id));
  });

  readonly selectedCount = computed(() => this.selectedRoleIds().size);

  ngOnInit(): void {
    forkJoin({
      users: this.usersService.list(1, PAGE_SIZE),
      roles: this.userRolesService.listAllRoles(),
    }).subscribe({
      next: ({ users, roles }) => {
        // Admins bypass the permission system entirely (is_admin in auth
        // middleware), so role assignment is meaningless for them.
        this.users.set((users ?? []).filter((u) => !u.isAdmin));
        this.roles.set(roles ?? []);
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت کاربران و نقش‌ها'));
      },
    });
  }

  fullName(user: UserListItem): string {
    return `${user.firstName ?? ''} ${user.lastName ?? ''}`.trim();
  }

  selectUser(user: UserListItem): void {
    if (this.selectedUser()?.id === user.id) {
      return;
    }

    this.selectedUser.set(user);
    this.assignedRoleIds.set(new Set());
    this.selectedRoleIds.set(new Set());
    this.rolesLoading.set(true);

    this.userRolesService.listForUser(user.id).subscribe({
      next: (assigned) => {
        this.rolesLoading.set(false);
        const ids = new Set((assigned ?? []).map((r) => r.id));
        this.assignedRoleIds.set(ids);
        this.selectedRoleIds.set(new Set(ids));
      },
      error: (err) => {
        this.rolesLoading.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت نقش‌های کاربر'));
      },
    });
  }

  isChecked(roleId: number): boolean {
    return this.selectedRoleIds().has(roleId);
  }

  toggle(roleId: number, checked: boolean): void {
    const next = new Set(this.selectedRoleIds());
    checked ? next.add(roleId) : next.delete(roleId);
    this.selectedRoleIds.set(next);
  }

  save(): void {
    const user = this.selectedUser();
    if (!user) {
      return;
    }

    const before = this.assignedRoleIds();
    const after = this.selectedRoleIds();

    const toAdd = [...after].filter((id) => !before.has(id));
    const toRemove = [...before].filter((id) => !after.has(id));

    if (toAdd.length === 0 && toRemove.length === 0) {
      return;
    }

    this.saving.set(true);
    forkJoin({
      assign: this.userRolesService.assign(user.id, toAdd),
      remove: this.userRolesService.remove(user.id, toRemove),
    }).subscribe({
      next: () => {
        this.saving.set(false);
        this.assignedRoleIds.set(new Set(after));
        this.toast.success('نقش‌های کاربر ذخیره شد');
      },
      error: (err) => {
        this.saving.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در ذخیره نقش‌ها'));
      },
    });
  }
}
