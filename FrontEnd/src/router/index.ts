import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import Login from '../components/Login.vue';
import Register from '../components/Register.vue';
import { useAuthStore } from '../store/auth';
import AdminUserView from '../views/AdminUserView.vue';
import CurrencyExchangeView from '../views/CurrencyExchangeView.vue';
import HomeView from '../views/HomeView.vue';
import NewsDetailView from '../views/NewsDetailView.vue';
import NewsView from '../views/NewsView.vue';
import ProfileView from '../views/ProfileView.vue';

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'Home', component: HomeView },
  { path: '/exchange', name: 'CurrencyExchange', component: CurrencyExchangeView },
  { path: '/news', name: 'News', component: NewsView },
  { path: '/news/:id', name: 'NewsDetail', component: NewsDetailView },
  { path: '/login', name: 'Login', component: Login },
  { path: '/register', name: 'Register', component: Register },
  { path: '/profile', name: 'Profile', component: ProfileView, meta: { requiresAuth: true } },
  { 
    path: '/admin/users', 
    name: 'AdminUsers', 
    component: AdminUserView, 
    meta: { requiresAuth: true, requiresAdmin: true } 
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// 路由守卫 - 认证检查和管理员权限检查
// FrontEnd/src/router/index.ts
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();
  
  // 检查是否需要登录
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login');
    return;
  }
  
  // 检查是否需要管理员权限
  if (to.meta.requiresAdmin) {
    if (!authStore.isAuthenticated) {
      next('/login');
      return;
    }
    
    // 确保用户信息已加载
    if (!authStore.user) {
      try {
        await authStore.fetchProfile();
      } catch (error) {
        next('/login');
        return;
      }
    }
    
    // 检查管理员权限
    if (authStore.user?.role !== 'admin') {
      next('/');
      return;
    }
  }
  
  next();
});

function checkAdminPermission(to: any, next: any) {
  const authStore = useAuthStore();
  if (authStore.user?.role !== 'admin') {
    next('/');
    return;
  }
  next();
}

export default router;
