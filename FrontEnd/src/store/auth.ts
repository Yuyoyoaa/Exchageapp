// FrontEnd/src/store/auth.ts
import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import axios from '../axios';

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

export interface RegisterPayload {
  username: string;
  password: string;
  nickname?: string;
  email?: string;
  avatar?: string;
}

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

  const fetchProfile = async () => {
    if (!token.value) return;
    try {
      const response = await axios.get<User>('/user/profile');
      user.value = response.data;
    } catch (error) {
      console.error('Failed to fetch profile', error);
      throw error;
    }
  };

  const login = async (username: string, password: string) => {
    try {
      const response = await axios.post('/auth/login', { username, password });
      token.value = response.data.token;
      localStorage.setItem('token', token.value || '');
      await fetchProfile();
    } catch (error: any) {
      if (error.response?.data?.error) {
        throw new Error(error.response.data.error);
      }
      throw error;
    }
  };

  const register = async (payload: RegisterPayload) => {
    try {
      const response = await axios.post('/auth/register', payload);
      token.value = response.data.token;
      localStorage.setItem('token', token.value || '');
      if (response.data.user) {
        user.value = response.data.user;
      }
    } catch (error: any) {
      if (error.response?.data?.error) {
        throw new Error(error.response.data.error);
      }
      throw error;
    }
  };

  const updateProfile = async (payload: UpdateProfilePayload) => {
    try {
      const response = await axios.put<User>('/user/profile', payload);
      user.value = response.data;
      return response.data;
    } catch (error: any) {
      if (error.response?.data?.error) {
        throw new Error(error.response.data.error);
      }
      throw error;
    }
  };

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

  return {
    token,
    user,
    isAuthenticated,
    login,
    register,
    fetchProfile,
    updateProfile,
    logout,
    checkAdminPermission
  };
});