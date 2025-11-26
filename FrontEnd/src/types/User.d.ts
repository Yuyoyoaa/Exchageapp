export interface User {
  ID: number; // 确保与后端 uint 类型兼容
  username: string;
  nickname: string;
  email: string;
  role: string;
  avatar: string;
  CreatedAt: string;
  UpdatedAt?: string;
}

export interface AdminUserListResponse {
  users: User[];
  total: number;
}

export interface ChangeRoleRequest {
  role: 'admin' | 'user';
}

export interface ChangeRoleResponse {
  message: string;
  user_id: number;
  new_role: string;
}