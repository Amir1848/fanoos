import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DataviewService } from './dataview.service';
import { DeviceListItem, DeviceSave, LightState } from './models';

/** A single home's device layout — one page comfortably covers all of them. */
const ALL_DEVICES_PAGE_SIZE = 500;

@Injectable({ providedIn: 'root' })
export class DevicesService {
  private readonly http = inject(HttpClient);
  private readonly dataviews = inject(DataviewService);

  listForEnvironment(environmentId: number): Observable<DeviceListItem[]> {
    return this.dataviews.fetch<DeviceListItem>('all-devices', 1, ALL_DEVICES_PAGE_SIZE, { environmentId });
  }

  create(device: DeviceSave): Observable<{ id: number }> {
    return this.http.post<{ id: number }>('/api/devices/create', device);
  }

  update(device: DeviceSave): Observable<void> {
    return this.http.post<void>('/api/devices/update', device);
  }

  delete(ids: number[]): Observable<void> {
    return this.http.post<void>('/api/devices/delete', { ids });
  }

  getLightState(deviceId: number): Observable<LightState> {
    return this.http.get<LightState>(`/api/devices/${deviceId}/state/light`);
  }

  sendLightCommand(deviceId: number, on: boolean): Observable<{ ok: boolean }> {
    return this.http.post<{ ok: boolean }>('/api/devices/command/light', { deviceId, on });
  }
}
