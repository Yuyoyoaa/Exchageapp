// FrontEnd/src/store/auth.ts
import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import axios from '../axios';

// 定义用户信息接口
export interface User {
  ID: number;
  username: string;
  role: string;
  nickname: string;
  email: string;
  avatar: string;
  CreatedAt?: string;
  UpdatedAt?: string;
}

// 注册所需参数接口
export interface RegisterPayload {
  username: string;
  password: string;
  nickname?: string;
  email?: string;
  avatar?: string;
}

// 更新信息参数接口
export interface UpdateProfilePayload {
  nickname?: string;
  email?: string;
  avatar?: string;
  password?: string;
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'));
  const user = ref<User | null>(null);

  const isAuthenticated = computed(() => !!token.value);

  // 获取当前用户信息
  const fetchProfile = async () => {
    if (!token.value) return;
    try {
      const response = await axios.get<User>('/user/profile');
      user.value = response.data;
    } catch (error) {
      console.error('Failed to fetch profile', error);
      // 如果获取用户信息失败（如token过期），可能需要登出
      // logout(); 
    }
  };

  // 登录
  const login = async (username: string, password: string) => {
    try {
      const response = await axios.post('/auth/login', { username, password });
      token.value = response.data.token;
      localStorage.setItem('token', token.value || '');
      
      // 登录成功后立即获取用户信息
      await fetchProfile();
    } catch (error: any) {
      // 透传后端具体的错误信息
      if (error.response && error.response.data && error.response.data.error) {
        throw new Error(error.response.data.error);
      }
      throw error;
    }
  };

  // 注册
  const register = async (payload: RegisterPayload) => {
    try {
      // 这里的 payload 包含了 username, password, nickname, email, avatar
      const response = await axios.post('/auth/register', payload);
      
      token.value = response.data.token;
      // 注册接口直接返回了 user 对象，直接保存，省去一次请求
      if (response.data.user) {
        user.value = response.data.user;
      }
      localStorage.setItem('token', token.value || '');
    } catch (error: any) {
      if (error.response && error.response.data && error.response.data.error) {
        throw new Error(error.response.data.error);
      }
      throw error;
    }
  };

  // 更新用户信息
  const updateProfile = async (payload: UpdateProfilePayload) => {
    try {
      const response = await axios.put<User>('/user/profile', payload);
      // 更新本地状态
      user.value = response.data;
      return response.data;
    } catch (error: any) {
      if (error.response && error.response.data && error.response.data.error) {
        throw new Error(error.response.data.error);
      }
      throw error;
    }
  }

  const logout = () => {
    token.value = null;
    user.value = null;
    localStorage.removeItem('token');
  };

 const checkAdminPermission = async (): Promise<boolean> => {
  if (!user.value) {
    await fetchProfile();
  }
  return user.value?.role === 'admin';
};

// 在 return 语句中添加这个方法
return {
  token,
  user,
  isAuthenticated,
  login,
  register,
  fetchProfile,
  updateProfile,
  logout,
  checkAdminPermission // 新增方法
};
});