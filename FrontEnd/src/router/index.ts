// FrontEnd/src/router/index.ts
import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
// ... 保持原有引用 ...
import Login from '../components/Login.vue';
import Register from '../components/Register.vue';
import CurrencyExchangeView from '../views/CurrencyExchangeView.vue';
import HomeView from '../views/HomeView.vue';
import NewsDetailView from '../views/NewsDetailView.vue';
import NewsView from '../views/NewsView.vue';
// 新增引用
import { useAuthStore } from '../store/auth';
import ProfileView from '../views/ProfileView.vue';

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'Home', component: HomeView },
  { path: '/exchange', name: 'CurrencyExchange', component: CurrencyExchangeView },
  { path: '/news', name: 'News', component: NewsView },
  { path: '/news/:id', name: 'NewsDetail', component: NewsDetailView },
  { path: '/login', name: 'Login', component: Login },
  { path: '/register', name: 'Register', component: Register },
  // 新增路由
  { path: '/profile', name: 'Profile', component: ProfileView, meta: { requiresAuth: true } },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// 简单的路由守卫（可选，但推荐）
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore();
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'Login' });
  } else {
    next();
  }
});

export default router;
