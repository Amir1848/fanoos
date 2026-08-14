import { Component, DestroyRef, inject, signal, computed } from '@angular/core';
import { forkJoin, of, switchMap, timer } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { EnvironmentListItem } from '../../core/api/models';
import { EnvironmentsService } from '../../core/api/environments.service';
import { DevicesService } from '../../core/api/devices.service';

interface CircuitVm {
  id: number;
  name: string;
  on: boolean;
  online: boolean;
  known: boolean;
}

interface SensorVm {
  id: number;
  name: string;
  value: string | null;
}

const ENV_KEY = 'marketino.mobile.env';
const POLL_MS = 3000;

@Component({
  selector: 'app-mobile-home',
  standalone: true,
  imports: [],
  templateUrl: './mobile-home.component.html',
  styleUrl: './mobile-home.component.scss',
})
export class MobileHomeComponent {
  private readonly environmentsService = inject(EnvironmentsService);
  private readonly devicesService = inject(DevicesService);
  private readonly destroyRef = inject(DestroyRef);

  readonly loading = signal(true);
  readonly envMenuOpen = signal(false);
  readonly environments = signal<EnvironmentListItem[]>([]);
  readonly circuits = signal<CircuitVm[]>([]);
  readonly sensors = signal<SensorVm[]>([]);
  readonly offlineCount = computed(() => this.circuits().filter((c) => c.known && !c.online).length);

  readonly selectedEnvId = signal<number | null>(this.readStoredEnv());

  constructor() {
    this.loadAll();

    timer(POLL_MS, POLL_MS)
      .pipe(
        switchMap(() => (document.visibilityState === 'visible' ? this.pollStates() : of(null))),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe();
  }

  get selectedEnv(): EnvironmentListItem | null {
    const id = this.selectedEnvId();
    return this.environments().find((e) => e.id === id) ?? null;
  }

  toggleEnvMenu(event: Event): void {
    event.stopPropagation();
    this.envMenuOpen.update((open) => !open);
  }

  closeEnvMenu(): void {
    this.envMenuOpen.set(false);
  }

  pickEnv(env: EnvironmentListItem): void {
    this.selectedEnvId.set(env.id);
    localStorage.setItem(ENV_KEY, String(env.id));
    this.closeEnvMenu();
    this.loading.set(true);
    this.loadDevices();
  }

  toggle(circuit: CircuitVm): void {
    const next = !circuit.on;
    this.patchCircuit(circuit.id, { on: next });

    this.devicesService.sendLightCommand(circuit.id, next).subscribe({
      // The next poll reconciles with the device's reported state.
      error: () => this.patchCircuit(circuit.id, { on: !next }),
    });
  }

  refresh(): void {
    this.loadAll();
  }

  private readStoredEnv(): number | null {
    const raw = localStorage.getItem(ENV_KEY);
    const id = raw === null ? Number.NaN : Number(raw);
    return Number.isFinite(id) ? id : null;
  }

  private loadAll(): void {
    this.environmentsService
      .list(1, 500)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (environments) => {
          this.environments.set(environments);
          const stored = this.selectedEnvId();
          const target = environments.find((e) => e.id === stored) ?? environments[0] ?? null;
          if (target && target.id !== stored) {
            this.selectedEnvId.set(target.id);
            localStorage.setItem(ENV_KEY, String(target.id));
          }
          this.loadDevices();
        },
        error: () => this.loading.set(false),
      });
  }

  private loadDevices(): void {
    const envId = this.selectedEnvId();
    if (envId === null) {
      this.loading.set(false);
      return;
    }

    this.devicesService
      .listForEnvironment(envId)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (devices) => {
          this.circuits.set(
            devices
              .filter((d) => d.deviceType === 1)
              .map((d) => ({ id: d.id, name: d.name, on: false, online: false, known: false })),
          );
          // Sensors (device type 3) have no backend state endpoint yet —
          // shown as read-only cards awaiting values.
          this.sensors.set(
            devices
              .filter((d) => d.deviceType === 3)
              .map((d) => ({ id: d.id, name: d.name, value: null })),
          );
          this.pollStates()
            .pipe(takeUntilDestroyed(this.destroyRef))
            .subscribe();
        },
        error: () => this.loading.set(false),
      });
  }

  private pollStates() {
    const lights = this.circuits();
    if (lights.length === 0) {
      this.loading.set(false);
      return of(null);
    }

    const states$ = forkJoin(lights.map((c) => this.devicesService.getLightState(c.id)));
    return states$.pipe(
      switchMap((states) => {
        lights.forEach((circuit, i) => {
          const state = states[i];
          if (state) {
            this.patchCircuit(circuit.id, { on: state.on, online: state.online, known: state.known });
          }
        });
        this.loading.set(false);
        return of(null);
      }),
      switchMap(() => of(null)),
    );
  }

  private patchCircuit(id: number, patch: Partial<CircuitVm>): void {
    this.circuits.update((list) => list.map((c) => (c.id === id ? { ...c, ...patch } : c)));
  }
}
