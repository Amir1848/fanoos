import {
  AfterViewInit,
  Component,
  ElementRef,
  OnDestroy,
  OnInit,
  computed,
  effect,
  inject,
  input,
  signal,
  viewChild,
} from '@angular/core';
import { Router } from '@angular/router';
import { Subscription, forkJoin, interval, startWith, switchMap } from 'rxjs';
import Konva from 'konva';
import { EnvironmentsService } from '../../core/api/environments.service';
import { DevicesService } from '../../core/api/devices.service';
import { DeviceLinksService } from '../../core/api/device-links.service';
import { apiErrorMessage } from '../../core/api/api-error';
import { DeviceLinkListItem, DeviceListItem, EnvironmentListItem, LightState, ControllerState, ControllerFirmware } from '../../core/api/models';
import { DEVICE_TYPES, DeviceType, DeviceTypeOption, deviceTypeLabel } from '../../core/device-types';
import { ConfirmService } from '../../shared/confirm/confirm.service';
import { ToastService } from '../../shared/toast/toast.service';
import { IconComponent } from '../../shared/icon/icon.component';
import { renderMarkdown } from '../../shared/markdown';

const STAGE_WIDTH = 1400;
const STAGE_HEIGHT = 800;
const DEVICE_MARGIN = 120;
const LIGHT_STATE_POLL_MS = 3000;

interface DeviceVisual {
  fill: string;
  radius: number;
}

const DEVICE_VISUALS: Record<number, DeviceVisual> = {
  [DeviceType.Light]: { fill: '#f4b400', radius: 18 },
  [DeviceType.Controller]: { fill: '#3f51b5', radius: 15 },
};
const DEFAULT_VISUAL: DeviceVisual = { fill: '#5b6472', radius: 15 };

const LINK_TYPE_TOGGLE = 1;

interface SelectedDeviceLink {
  link: DeviceLinkListItem;
  otherName: string;
  direction: 'به' | 'از';
}

@Component({
  selector: 'app-environment-editor',
  standalone: true,
  imports: [IconComponent],
  templateUrl: './environment-editor.component.html',
  styleUrl: './environment-editor.component.scss',
})
export class EnvironmentEditorComponent implements OnInit, AfterViewInit, OnDestroy {
  private readonly environmentsService = inject(EnvironmentsService);
  private readonly devicesService = inject(DevicesService);
  private readonly deviceLinksService = inject(DeviceLinksService);
  private readonly confirm = inject(ConfirmService);
  private readonly toast = inject(ToastService);
  private readonly router = inject(Router);

  /** Route param, bound via withComponentInputBinding */
  readonly id = input.required<string>();

  private readonly canvasHost = viewChild.required<ElementRef<HTMLDivElement>>('canvasHost');

  readonly stageWidth = STAGE_WIDTH;
  readonly stageHeight = STAGE_HEIGHT;

  readonly deviceTypes = DEVICE_TYPES;
  readonly DeviceType = DeviceType;

  readonly loading = signal(true);
  readonly environment = signal<EnvironmentListItem | null>(null);
  readonly devices = signal<DeviceListItem[]>([]);
  readonly links = signal<DeviceLinkListItem[]>([]);
  readonly selectedDeviceId = signal<number | null>(null);
  readonly linkMode = signal(false);
  readonly linkSourceId = signal<number | null>(null);

  readonly lightState = signal<LightState | null>(null);
  readonly lightStateLoading = signal(false);
  readonly lightCommandPending = signal(false);
  private lightStatePoll: Subscription | null = null;

  readonly controllerState = signal<ControllerState | null>(null);
  readonly controllerStateLoading = signal(false);
  private controllerStatePoll: Subscription | null = null;

  readonly firmwareLoading = signal(false);
  readonly firmware = signal<ControllerFirmware | null>(null);
  readonly firmwareTab = signal<'ino' | 'readme'>('ino');

  readonly readmeHtml = computed(() => {
    const fw = this.firmware();
    return fw ? renderMarkdown(fw.readme) : '';
  });

  readonly selectedDevice = computed(
    () => this.devices().find((d) => d.id === this.selectedDeviceId()) ?? null,
  );

  readonly selectedDeviceLinks = computed<SelectedDeviceLink[]>(() => {
    const id = this.selectedDeviceId();
    if (!id) {
      return [];
    }
    return this.links()
      .filter((l) => l.sourceDeviceId === id || l.targetDeviceId === id)
      .map((l) => {
        const isSource = l.sourceDeviceId === id;
        const otherId = isSource ? l.targetDeviceId : l.sourceDeviceId;
        const other = this.devices().find((d) => d.id === otherId);
        return { link: l, otherName: other?.name ?? '؟', direction: isSource ? 'به' : 'از' };
      });
  });

  private stage: Konva.Stage | null = null;
  private layer: Konva.Layer | null = null;

  constructor() {
    effect(() => {
      const device = this.selectedDevice();
      this.stopLightStatePoll();
      this.stopControllerStatePoll();
      if (device?.deviceType === DeviceType.Light) {
        this.startLightStatePoll(device.id);
      } else if (device?.deviceType === DeviceType.Controller) {
        this.startControllerStatePoll(device.id);
      } else {
        this.lightState.set(null);
        this.controllerState.set(null);
      }
    });
  }

  private environmentId(): number {
    return +this.id();
  }

  ngOnInit(): void {
    const environmentId = this.environmentId();

    forkJoin({
      environment: this.environmentsService.getById(environmentId),
      devices: this.devicesService.listForEnvironment(environmentId),
      links: this.deviceLinksService.listForEnvironment(environmentId),
    }).subscribe({
      next: ({ environment, devices, links }) => {
        this.loading.set(false);
        this.environment.set(environment);
        this.devices.set(devices ?? []);
        this.links.set(links ?? []);
        this.render();
      },
      error: (err) => {
        this.loading.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت اطلاعات محیط'));
      },
    });
  }

  ngAfterViewInit(): void {
    this.stage = new Konva.Stage({
      container: this.canvasHost().nativeElement,
      width: STAGE_WIDTH,
      height: STAGE_HEIGHT,
    });
    this.layer = new Konva.Layer();
    this.stage.add(this.layer);

    this.stage.on('click', (e: Konva.KonvaEventObject<MouseEvent>) => {
      if (e.target !== this.stage) {
        return;
      }
      if (this.linkMode()) {
        this.linkSourceId.set(null);
      } else {
        this.selectedDeviceId.set(null);
      }
      this.render();
    });

    this.render();
  }

  ngOnDestroy(): void {
    this.stage?.destroy();
    this.stopLightStatePoll();
    this.stopControllerStatePoll();
  }

  private startLightStatePoll(deviceId: number): void {
    this.lightStateLoading.set(true);
    this.lightStatePoll = interval(LIGHT_STATE_POLL_MS)
      .pipe(
        startWith(0),
        switchMap(() => this.devicesService.getLightState(deviceId)),
      )
      .subscribe({
        next: (state) => {
          this.lightStateLoading.set(false);
          this.lightState.set(state);
        },
        error: () => {
          this.lightStateLoading.set(false);
        },
      });
  }

  private stopLightStatePoll(): void {
    this.lightStatePoll?.unsubscribe();
    this.lightStatePoll = null;
    this.lightStateLoading.set(false);
  }

  private startControllerStatePoll(deviceId: number): void {
    this.controllerStateLoading.set(true);
    this.controllerStatePoll = interval(LIGHT_STATE_POLL_MS)
      .pipe(
        startWith(0),
        switchMap(() => this.devicesService.getControllerState(deviceId)),
      )
      .subscribe({
        next: (state) => {
          this.controllerStateLoading.set(false);
          this.controllerState.set(state);
        },
        error: () => {
          this.controllerStateLoading.set(false);
        },
      });
  }

  private stopControllerStatePoll(): void {
    this.controllerStatePoll?.unsubscribe();
    this.controllerStatePoll = null;
    this.controllerStateLoading.set(false);
  }

  setLightState(on: boolean): void {
    const device = this.selectedDevice();
    if (!device || this.lightCommandPending()) {
      return;
    }

    this.lightCommandPending.set(true);
    this.devicesService.sendLightCommand(device.id, on).subscribe({
      next: () => {
        this.lightCommandPending.set(false);
        this.lightState.update((s) => ({ on, online: s?.online ?? true, ts: s?.ts ?? 0, known: true }));
      },
      error: (err) => {
        this.lightCommandPending.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در ارسال دستور به چراغ'));
      },
    });
  }

  private render(): void {
    if (!this.layer) {
      return;
    }
    this.layer.destroyChildren();

    for (const link of this.links()) {
      const source = this.devices().find((d) => d.id === link.sourceDeviceId);
      const target = this.devices().find((d) => d.id === link.targetDeviceId);
      if (!source || !target) {
        continue;
      }
      this.layer.add(
        new Konva.Line({
          points: [source.positionX, source.positionY, target.positionX, target.positionY],
          stroke: '#a8aebd',
          strokeWidth: 2,
          dash: [6, 4],
        }),
      );
    }

    for (const device of this.devices()) {
      this.layer.add(this.buildDeviceNode(device));
    }

    this.layer.draw();
  }

  private buildDeviceNode(device: DeviceListItem): Konva.Group {
    const visual = DEVICE_VISUALS[device.deviceType] ?? DEFAULT_VISUAL;
    const isSelected = this.selectedDeviceId() === device.id;
    const isLinkSource = this.linkSourceId() === device.id;

    const group = new Konva.Group({ x: device.positionX, y: device.positionY, draggable: true });

    const circle = new Konva.Circle({
      radius: visual.radius,
      fill: visual.fill,
      stroke: isLinkSource ? '#d32f2f' : isSelected ? '#24292f' : '#ffffff',
      strokeWidth: isSelected || isLinkSource ? 3 : 2,
    });

    const label = new Konva.Text({
      text: device.name,
      y: visual.radius + 6,
      fontSize: 12,
      fill: '#24292f',
    });
    label.offsetX(label.width() / 2);

    group.add(circle);
    group.add(label);

    group.on('dragend', () => {
      const pos = group.position();
      this.updateDevicePosition(device.id, pos.x, pos.y);
    });

    group.on('click tap', (e: Konva.KonvaEventObject<MouseEvent>) => {
      e.cancelBubble = true;
      this.onDeviceClicked(device.id);
    });

    return group;
  }

  private onDeviceClicked(deviceId: number): void {
    if (this.linkMode()) {
      const sourceId = this.linkSourceId();
      if (!sourceId) {
        this.linkSourceId.set(deviceId);
        this.render();
        return;
      }
      if (sourceId === deviceId) {
        this.linkSourceId.set(null);
        this.render();
        return;
      }

      if (!this.isLightControllerPair(sourceId, deviceId)) {
        this.toast.error('اتصال فقط بین یک چراغ و یک کنترلر ممکن است');
        this.linkSourceId.set(null);
        this.render();
        return;
      }

      this.createLink(sourceId, deviceId);
      return;
    }

    this.selectedDeviceId.set(deviceId);
    this.render();
  }

  private updateDevicePosition(deviceId: number, x: number, y: number): void {
    const device = this.devices().find((d) => d.id === deviceId);
    if (!device) {
      return;
    }

    this.devices.update((list) =>
      list.map((d) => (d.id === deviceId ? { ...d, positionX: x, positionY: y } : d)),
    );

    this.devicesService.update({ ...device, positionX: x, positionY: y }).subscribe({
      error: (err) => this.toast.error(apiErrorMessage(err, 'خطا در ذخیره موقعیت دستگاه')),
    });
  }

  toggleLinkMode(): void {
    this.linkMode.update((v) => !v);
    this.linkSourceId.set(null);
    this.render();
  }

  addDevice(type: DeviceTypeOption): void {
    const environmentId = this.environmentId();
    const index = this.devices().filter((d) => d.deviceType === type.code).length + 1;
    const positionX = DEVICE_MARGIN + Math.random() * (STAGE_WIDTH - 2 * DEVICE_MARGIN);
    const positionY = DEVICE_MARGIN + Math.random() * (STAGE_HEIGHT - 2 * DEVICE_MARGIN);
    const name = `${type.label} ${index}`;

    this.devicesService
      .create({ id: 0, environmentId, deviceType: type.code, name, positionX, positionY, rotation: 0 })
      .subscribe({
        next: ({ id }) => {
          this.devices.update((list) => [
            ...list,
            { id, environmentId, deviceType: type.code, name, positionX, positionY, rotation: 0 },
          ]);
          this.render();
        },
        error: (err) => this.toast.error(apiErrorMessage(err, 'خطا در افزودن دستگاه')),
      });
  }

  renameSelectedDevice(value: string): void {
    const device = this.selectedDevice();
    const name = value.trim();
    if (!device || !name || name === device.name) {
      return;
    }

    this.devices.update((list) => list.map((d) => (d.id === device.id ? { ...d, name } : d)));
    this.devicesService.update({ ...device, name }).subscribe({
      error: (err) => this.toast.error(apiErrorMessage(err, 'خطا در تغییر نام دستگاه')),
    });
    this.render();
  }

  async deleteSelectedDevice(): Promise<void> {
    const device = this.selectedDevice();
    if (!device) {
      return;
    }

    const ok = await this.confirm.ask({
      title: 'حذف دستگاه',
      message: `آیا از حذف «${device.name}» مطمئن هستید؟ اتصالات مرتبط با آن نیز حذف می‌شود.`,
      confirmText: 'حذف',
      danger: true,
    });
    if (!ok) {
      return;
    }

    this.devicesService.delete([device.id]).subscribe({
      next: () => {
        this.devices.update((list) => list.filter((d) => d.id !== device.id));
        this.links.update((list) =>
          list.filter((l) => l.sourceDeviceId !== device.id && l.targetDeviceId !== device.id),
        );
        this.selectedDeviceId.set(null);
        this.render();
      },
      error: (err) => this.toast.error(apiErrorMessage(err, 'خطا در حذف دستگاه')),
    });
  }

  private isLightControllerPair(aId: number, bId: number): boolean {
    const a = this.devices().find((d) => d.id === aId)?.deviceType;
    const b = this.devices().find((d) => d.id === bId)?.deviceType;
    return (
      (a === DeviceType.Light && b === DeviceType.Controller) ||
      (a === DeviceType.Controller && b === DeviceType.Light)
    );
  }

  private createLink(firstId: number, secondId: number): void {
    const linkType = LINK_TYPE_TOGGLE;
    const firstDevice = this.devices().find((d) => d.id === firstId);
    const secondDevice = this.devices().find((d) => d.id === secondId);

    const sourceDeviceId = firstDevice?.deviceType === DeviceType.Controller ? firstId : secondId;
    const targetDeviceId = firstDevice?.deviceType === DeviceType.Controller ? secondId : firstId;

    this.deviceLinksService.create({ sourceDeviceId, targetDeviceId, linkType }).subscribe({
      next: ({ id }) => {
        this.links.update((list) => [
          ...list,
          { id, sourceDeviceId, targetDeviceId, linkType },
        ]);
        this.linkSourceId.set(null);
        this.render();
      },
      error: (err) => {
        this.linkSourceId.set(null);
        this.toast.error(apiErrorMessage(err, 'خطا در ایجاد اتصال (شاید این اتصال قبلاً ثبت شده است)'));
        this.render();
      },
    });
  }

  deleteLink(linkId: number): void {
    this.deviceLinksService.delete([linkId]).subscribe({
      next: () => {
        this.links.update((list) => list.filter((l) => l.id !== linkId));
        this.render();
      },
      error: (err) => this.toast.error(apiErrorMessage(err, 'خطا در حذف اتصال')),
    });
  }

  deviceTypeLabel(device: DeviceListItem): string {
    return deviceTypeLabel(device.deviceType);
  }

  loadFirmware(): void {
    const device = this.selectedDevice();
    if (!device || this.firmwareLoading()) {
      return;
    }

    this.firmwareLoading.set(true);
    this.devicesService.getFirmware(device.id).subscribe({
      next: (fw) => {
        this.firmwareLoading.set(false);
        this.firmwareTab.set('ino');
        this.firmware.set(fw);
      },
      error: (err) => {
        this.firmwareLoading.set(false);
        this.toast.error(apiErrorMessage(err, 'خطا در دریافت سورس کد کنترلر'));
      },
    });
  }

  closeFirmware(): void {
    this.firmware.set(null);
  }

  private firmwareFile(): { name: string; content: string } | null {
    const fw = this.firmware();
    if (!fw) {
      return null;
    }
    return this.firmwareTab() === 'ino'
      ? { name: 'controller.ino', content: fw.ino }
      : { name: 'README.md', content: fw.readme };
  }

  downloadFirmwareFile(): void {
    const file = this.firmwareFile();
    if (!file) {
      return;
    }

    const url = URL.createObjectURL(new Blob([file.content], { type: 'text/plain;charset=utf-8' }));
    const anchor = document.createElement('a');
    anchor.href = url;
    anchor.download = file.name;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  async copyFirmwareFile(): Promise<void> {
    const file = this.firmwareFile();
    if (!file) {
      return;
    }

    try {
      await navigator.clipboard.writeText(file.content);
      this.toast.success(`${file.name} کپی شد`);
    } catch {
      this.toast.error('کپی در کلیپ‌بورد ممکن نشد');
    }
  }

  back(): void {
    this.router.navigate(['/environments']);
  }
}
