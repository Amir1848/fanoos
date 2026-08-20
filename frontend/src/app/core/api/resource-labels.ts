/**
 * Persian (fa) labels for backend resource keys (see resources table / resource_key column),
 * shown in the role-access UI. Add an entry whenever a new resource_key is introduced.
 * Unmapped keys fall back to the raw key, so nothing is ever hidden from the UI.
 */
export const RESOURCE_LABELS: Record<string, string> = {
  'users.create': 'ایجاد کاربر',
  'users.delete': 'حذف کاربر',
  'users.view': 'مشاهده کاربران',
  'roles.create': 'ایجاد نقش',
  'roles.update': 'ویرایش نقش',
  'roles.delete': 'حذف نقش',
  'roles.view': 'مشاهده نقش‌ها',
};

/** Persian labels for the group prefix of a resource key (e.g. "users" in "users.create"). */
export const RESOURCE_GROUP_LABELS: Record<string, string> = {
  users: 'کاربران',
  roles: 'نقش‌ها',
};

/** app-icon name for a group prefix; falls back to a generic lock icon. */
export const RESOURCE_GROUP_ICONS: Record<string, string> = {
  users: 'users',
  roles: 'shield',
};

export function resourceLabel(resourceKey: string): string {
  return RESOURCE_LABELS[resourceKey] ?? resourceKey;
}

export function resourceGroupLabel(group: string): string {
  return RESOURCE_GROUP_LABELS[group] ?? group;
}

export function resourceGroupIcon(group: string): string {
  return RESOURCE_GROUP_ICONS[group] ?? 'lock';
}
