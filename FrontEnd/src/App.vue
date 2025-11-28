<template>
  <el-container>
    <el-header>
      <el-menu 
        :default-active="activeIndex" 
        class="el-menu-demo" 
        mode="horizontal" 
        :ellipsis="false" 
        @select="handleSelect"
        router
      >
        <el-menu-item index="/" style="font-weight: bold; color: #409EFF;">
          💰 YOYO兑换基地
        </el-menu-item>
        
        <el-menu-item index="/">首页</el-menu-item>
        <el-menu-item index="/exchange">货币兑换</el-menu-item>
        <el-menu-item index="/news">新闻资讯</el-menu-item>
        
        <el-submenu index="admin" v-if="authStore.user?.role === 'admin'">
          <template #title>
            <el-icon><Setting /></el-icon>
            管理员
          </template>
          <el-menu-item index="/admin/users">用户管理</el-menu-item>
          <el-menu-item index="/admin/articles">文章管理</el-menu-item> </el-submenu>
        
        <div class="flex-grow"></div>
        
        <div class="user-menu" v-if="authStore.isAuthenticated">
          <el-submenu index="user">
            <template #title>
              <el-avatar 
                :size="32" 
                :src="authStore.user?.avatar" 
                style="margin-right: 8px;"
              >
                {{ authStore.user?.username?.charAt(0)?.toUpperCase() || 'U' }}
              </el-avatar>
              {{ authStore.user?.nickname || authStore.user?.username }}
              <el-tag v-if="authStore.user?.role === 'admin'" size="small" type="danger" style="margin-left: 8px;">
                管理员
              </el-tag>
            </template>
            <el-menu-item index="/profile">个人中心</el-menu-item>
            <el-menu-item index="logout">退出登录</el-menu-item>
          </el-submenu>
        </div>
        
        <div class="auth-menu" v-else>
          <el-menu-item index="/login">登录</el-menu-item>
          <el-menu-item index="/register">注册</el-menu-item>
        </div>
      </el-menu>
    </el-header>
    
    <el-main style="min-height: calc(100vh - 60px);">
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
const activeIndex = ref('/');

// 监听路由变化高亮菜单
watch(route, (newRoute) => {
  updateActiveIndex(newRoute);
});

// 应用加载时，如果已登录但没有用户信息，尝试获取一次
onMounted(() => {
  updateActiveIndex(route);
  if (authStore.isAuthenticated && !authStore.user) {
    authStore.fetchProfile().catch(() => {
      // 如果获取用户信息失败，可能是token过期，清除token
      authStore.logout();
    });
  }
});

const updateActiveIndex = (currentRoute: any) => {
  // 如果是嵌套路由或参数路由，确保高亮对应的主菜单
  if (currentRoute.path.startsWith('/admin')) {
    // 保持 admin 子菜单高亮逻辑由 element-plus 自动处理，或者手动指定
    activeIndex.value = currentRoute.path;
  } else if (currentRoute.path.startsWith('/news')) {
    activeIndex.value = '/news';
  } else {
    activeIndex.value = currentRoute.path;
  }
};

const handleSelect = (key: string) => {
  if (key === 'logout') {
    authStore.logout();
    router.push({ name: 'Home' });
  } else if (key.startsWith('/')) {
    // 路由跳转由 router 属性处理
  }
};
</script>

<style scoped>
.el-menu-demo {
  line-height: 60px;
  display: flex;
  align-items: center;
}

.flex-grow {
  flex-grow: 1;
}

.user-menu {
  display: flex;
  align-items: center;
}

.auth-menu {
  display: flex;
}

:deep(.el-menu--horizontal) {
  border-bottom: none;
}

:deep(.el-menu--horizontal > .el-menu-item) {
  border-bottom: 2px solid transparent;
}

:deep(.el-menu--horizontal > .el-menu-item.is-active) {
  border-bottom-color: #409EFF;
}
</style>