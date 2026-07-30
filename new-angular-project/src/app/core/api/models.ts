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
