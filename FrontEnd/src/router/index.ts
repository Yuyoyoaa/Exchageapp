import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import { useAuthStore } from '../store/auth';

// 页面组件
import Login from '../components/Login.vue';
import Register from '../components/Register.vue';
import AdminArticleView from '../views/AdminArticleView.vue';
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
  { 
    path: '/profile', 
    name: 'Profile', 
    component: ProfileView, 
    meta: { requiresAuth: true } 
  },
  { 
    path: '/admin/users', 
    name: 'AdminUsers', 
    component: AdminUserView, 
    meta: { requiresAuth: true, requiresAdmin: true } 
  },
  { 
    path: '/admin/articles', 
    name: 'AdminArticles', 
    component: AdminArticleView, 
    meta: { requiresAuth: true, requiresAdmin: true } 
  },
  // 可选：404页面
  { path: '/:pathMatch(.*)*', redirect: '/' },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// 全局前置守卫：权限控制
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore();

  // 检查登录权限
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login' });
    return;
  }

  // 检查管理员权限
  if (to.meta.requiresAdmin) {
    if (!authStore.isAuthenticated) {
      next({ name: 'Login' });
      return;
    }

    // 确保用户信息已加载
    if (!authStore.user) {
      try {
        await authStore.fetchProfile();
      } catch {
        next({ name: 'Login' });
        return;
      }
    }

    // 非管理员禁止访问
    if (authStore.user?.role !== 'admin') {
      next({ name: 'Home' });
      return;
    }
  }

  next();
});

export default router;
