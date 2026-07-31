export const DeviceType = {
  Light: 1,
  Controller: 2,
} as const;

export interface DeviceTypeOption {
  code: number;
  label: string;
}

export const DEVICE_TYPES: DeviceTypeOption[] = [
  { code: DeviceType.Light, label: 'چراغ' },
  { code: DeviceType.Controller, label: 'کنترلر' },
];

export function deviceTypeLabel(code: number): string {
  return DEVICE_TYPES.find((t) => t.code === code)?.label ?? `${code}`;
}
