import { Component, OnInit, computed, inject, input, signal } from '@angular/core';
import { Router } from '@angular/router';
import { forkJoin, of } from 'rxjs';
import { RolesService } from '../../core/api/roles.service';
import { RoleResourcesService } from '../../core/api/role-resources.service';
import { apiErrorMessage } from '../../core/api/api-error';
import { ResourceItem, RoleListItem } from '../../core/api/models';
import { resourceGroupIcon, resourceGroupLabel, resourceLabel } from '../../core/api/resource-labels';
import { ToastService } from '../../shared/toast/toast.service';
import { IconComponent } from '../../shared/icon/icon.component';

interface ResourceGroup {
  name: string;
  resources: ResourceItem[];
}

@Component({
  selector: 'app-role-access',
  standalone: true,
  imports: [IconComponent],
  templateUrl: './role-access.component.html',
  styleUrl: './role-access.component.scss',
})
export class RoleAccessComponent implements OnInit {
  private readonly rolesService = inject(RolesService);
  private readonly roleResourcesService = inject(RoleResourcesService);
  private readonly router = inject(Router);
  private readonly toast = inject(ToastService);

  /** Route param, bound via withComponentInputBinding */
  readonly id = input.required<string>();

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly notFound = signal(false);
  readonly role = signal<RoleListItem | null>(null);
  readonly resources = signal<ResourceItem[]>([]);
  readonly assignedIds = signal<ReadonlySet<number>>(new Set());
  readonly selectedIds = signal<ReadonlySet<number>>(new Set());

  readonly groups = computed<ResourceGroup[]>(() => {
    const byGroup = new Map<string, ResourceItem[]>();
    for (const resource of this.resources()) {
      const name = resource.resourceKey.split('.')[0] || resource.resourceKey;
      const list = byGroup.get(name) ?? [];
      list.push(resource);
      byGroup.set(name, list);
    }
    return [...byGroup.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([name, resources]) => ({ name, resources }));
  });

  readonly dirty = computed(() => {
    const before = this.assignedIds();
    const after = this.selectedIds();
    return before.size !== after.size || [...before].some((id) => !after.has(id));
  });

  ngOnInit(): void {
    const roleId = +this.id();

    const fromState = (history.state as { role?: RoleListItem }).role;
    if (fromState && fromState.id === roleId) {
      this.role.set(fromState);
    }

    const role$ = this.role() ? of(this.role()) : this.rolesService.getById(roleId);

    forkJoin({
      role: role$,
      resources: this.roleResourcesService.listAllResources(),
      assigned: this.roleResourcesService.listForRole(roleId),
    }).subscribe({
      next: ({ role, resources, assigned }) => {
        this.loading.set(false);

        if (!role) {
          this.notFound.set(true);
          return;
        }

        this.role.set(role);
        this.resources.set(resources);

        const ids = new Set(assigned.map((r) => r.id));
        this.assignedIds.set(ids);
        this.selectedIds.set(new Set(ids));
      },
      error: (err) => {
        this.loading.set(false);
        this.notFound.set(true);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت اطلاعات دسترسی‌ها'));
      },
    });
  }

  groupLabel(group: ResourceGroup): string {
    return resourceGroupLabel(group.name);
  }

  groupIcon(group: ResourceGroup): string {
    return resourceGroupIcon(group.name);
  }

  selectedCount(group: ResourceGroup): number {
    const selected = this.selectedIds();
    return group.resources.filter((r) => selected.has(r.id)).length;
  }

  keyLabel(resource: ResourceItem): string {
    return resourceLabel(resource.resourceKey);
  }

  isChecked(resourceId: number): boolean {
    return this.selectedIds().has(resourceId);
  }

  toggle(resourceId: number, checked: boolean): void {
    const next = new Set(this.selectedIds());
    checked ? next.add(resourceId) : next.delete(resourceId);
    this.selectedIds.set(next);
  }

  isGroupFullySelected(group: ResourceGroup): boolean {
    const selected = this.selectedIds();
    return group.resources.every((r) => selected.has(r.id));
  }

  toggleGroup(group: ResourceGroup, checked: boolean): void {
    const next = new Set(this.selectedIds());
    for (const resource of group.resources) {
      checked ? next.add(resource.id) : next.delete(resource.id);
    }
    this.selectedIds.set(next);
  }

  save(): void {
    const roleId = +this.id();
    const before = this.assignedIds();
    const after = this.selectedIds();

    const toAdd = [...after].filter((id) => !before.has(id));
    const toRemove = [...before].filter((id) => !after.has(id));

    if (toAdd.length === 0 && toRemove.length === 0) {
      return;
    }

    this.saving.set(true);
    forkJoin({
      assign: this.roleResourcesService.assign(roleId, toAdd),
      remove: this.roleResourcesService.remove(roleId, toRemove),
    }).subscribe({
      next: () => {
        this.saving.set(false);
        this.assignedIds.set(new Set(after));
        this.toast.success('دسترسی‌های نقش ذخیره شد');
      },
      error: (err) => {
        this.saving.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در ذخیره دسترسی‌ها'));
      },
    });
  }

  back(): void {
    this.router.navigate(['/roles']);
  }
}
