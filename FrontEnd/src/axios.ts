import axios from 'axios';
import { ElMessage } from 'element-plus';
import { useAuthStore } from './store/auth';

const instance = axios.create({
  baseURL: 'http://localhost:3080/api',
  timeout: 10000,
});

// 请求拦截器
instance.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = token;
  }
  return config;
}, error => {
  return Promise.reject(error);
});

// 响应拦截器
instance.interceptors.response.use(
  response => {
    return response;
  },
  error => {
    if (error.response?.status === 401) {
      // token 过期或无效
      const authStore = useAuthStore();
      authStore.logout();
      ElMessage.error('登录已过期，请重新登录');
      window.location.href = '/login';
    } else if (error.response?.status === 403) {
      // 权限不足
      ElMessage.error('权限不足，无法访问此功能');
      if (window.location.pathname !== '/') {
        window.location.href = '/';
      }
    } else if (error.response?.status >= 500) {
      ElMessage.error('服务器错误，请稍后重试');
    } else if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时，请检查网络连接');
    }
    return Promise.reject(error);
  }
);


export default instance;