export interface UserListItem {
  id: number;
  firstName: string;
  lastName: string;
  userName: string;
  phoneNumber: string;
  email: string;
  isAdmin: boolean;
}

export interface UserSave {
  id: number;
  userName: string;
  firstName: string;
  lastName: string;
  email: string;
  phoneNumber: string;
  password: string;
  isAdmin: boolean;
}

export interface RoleListItem {
  id: number;
  title: string;
  code: string;
}

export interface RoleSave {
  id: number;
  title: string;
  code: string;
}

export interface ResourceItem {
  id: number;
  resourceKey: string;
}

export interface EnvironmentListItem {
  id: number;
  name: string;
  description: string;
}

export interface EnvironmentSave {
  id: number;
  name: string;
  description: string;
}

export interface DeviceListItem {
  id: number;
  environmentId: number;
  deviceType: number;
  name: string;
  positionX: number;
  positionY: number;
  rotation: number;
}

export interface DeviceSave {
  id: number;
  environmentId: number;
  deviceType: number;
  name: string;
  positionX: number;
  positionY: number;
  rotation: number;
}

export interface DeviceLinkListItem {
  id: number;
  sourceDeviceId: number;
  targetDeviceId: number;
  linkType: string;
}

export interface DeviceLinkSave {
  sourceDeviceId: number;
  targetDeviceId: number;
  linkType: string;
}
