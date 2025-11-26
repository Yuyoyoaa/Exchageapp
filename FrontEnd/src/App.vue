<template>
  <el-container>
    <el-header>
      <el-menu 
        :default-active="activeIndex" 
        class="el-menu-demo" 
        mode="horizontal" 
        :ellipsis="true" 
        @select="handleSelect"
      >
        <el-menu-item index="home">首页</el-menu-item>
        <el-menu-item index="currencyExchange">货币兑换</el-menu-item>
        <el-menu-item index="news">新闻资讯</el-menu-item>
        
        <!-- 管理员菜单 -->
        <el-submenu index="admin" v-if="authStore.user?.role === 'admin'">
          <template #title>
            <el-icon><Setting /></el-icon>
            管理员
          </template>
          <el-menu-item index="adminUsers">用户管理</el-menu-item>
        </el-submenu>
        
        <!-- 用户相关菜单 -->
        <div class="user-menu" v-if="authStore.isAuthenticated">
          <el-submenu index="user">
            <template #title>
              <el-avatar 
                :size="32" 
                :src="authStore.user?.avatar" 
                style="margin-right: 8px;"
              >
                {{ authStore.user?.username?.charAt(0)?.toUpperCase() }}
              </el-avatar>
              {{ authStore.user?.nickname || authStore.user?.username }}
              <el-tag v-if="authStore.user?.role === 'admin'" size="small" type="danger" style="margin-left: 8px;">
                管理员
              </el-tag>
            </template>
            <el-menu-item index="profile">个人资料</el-menu-item>
            <el-menu-item index="logout">退出登录</el-menu-item>
          </el-submenu>
        </div>
        
        <div class="auth-menu" v-else>
          <el-menu-item index="login">登录</el-menu-item>
          <el-menu-item index="register">注册</el-menu-item>
        </div>
      </el-menu>
    </el-header>
    
    <el-main>
      <router-view></router-view>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { Setting } from '@element-plus/icons-vue';
import { onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from './store/auth';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const activeIndex = ref('home');

// 监听路由变化高亮菜单
watch(route, (newRoute) => {
  updateActiveIndex(newRoute);
});

// 应用加载时，如果已登录但没有用户信息，尝试获取一次
onMounted(() => {
  updateActiveIndex(route);
  if (authStore.isAuthenticated && !authStore.user) {
    authStore.fetchProfile();
  }
});

const updateActiveIndex = (currentRoute: any) => {
  const routeName = currentRoute.name?.toString().toLowerCase();
  switch (routeName) {
    case 'home':
      activeIndex.value = 'home';
      break;
    case 'currencyexchange':
      activeIndex.value = 'currencyExchange';
      break;
    case 'news':
    case 'newsdetail':
      activeIndex.value = 'news';
      break;
    case 'profile':
      activeIndex.value = 'profile';
      break;
    case 'adminusers':
      activeIndex.value = 'adminUsers';
      break;
    default:
      activeIndex.value = 'home';
  }
};

const handleSelect = (key: string) => {
  switch (key) {
    case 'logout':
      authStore.logout();
      router.push({ name: 'Home' });
      break;
    case 'profile':
      router.push({ name: 'Profile' });
      break;
    case 'adminUsers':
      router.push({ name: 'AdminUsers' });
      break;
    case 'home':
      router.push({ name: 'Home' });
      break;
    case 'currencyExchange':
      router.push({ name: 'CurrencyExchange' });
      break;
    case 'news':
      router.push({ name: 'News' });
      break;
    case 'login':
      router.push({ name: 'Login' });
      break;
    case 'register':
      router.push({ name: 'Register' });
      break;
  }
};
</script>

<style scoped>
.el-menu-demo {
  line-height: 60px;
  display: flex;
  justify-content: space-between;
}

.user-menu {
  display: flex;
  align-items: center;
}

.auth-menu {
  display: flex;
}
</style>