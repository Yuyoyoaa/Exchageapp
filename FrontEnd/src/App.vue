<template>
  <el-container class="layout-container">
    <!-- 顶部导航栏 -->
    <el-header class="app-header">
      <div class="header-content">
        <!-- 左侧 LOGO -->
        <div class="logo-area" @click="router.push('/')">
          <span class="logo-icon">💰</span>
          <span class="logo-text">YOYO兑换基地</span>
        </div>

        <!-- 中间 导航菜单 -->
        <el-menu
          :default-active="activeIndex"
          class="nav-menu"
          mode="horizontal"
          :ellipsis="false"
          @select="handleSelect"
          router
        >
          <el-menu-item index="/">首页</el-menu-item>
          <el-menu-item index="/exchange">货币兑换</el-menu-item>
          <el-menu-item index="/news">新闻资讯</el-menu-item>
          
          <!-- 管理员菜单 -->
          <el-sub-menu index="admin" v-if="authStore.user?.role === 'admin'">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>管理员</span>
            </template>
            <el-menu-item index="/admin/users">用户管理</el-menu-item>
            <el-menu-item index="/admin/articles">文章管理</el-menu-item>
          </el-sub-menu>
        </el-menu>

        <!-- 右侧 用户/登录区 -->
        <div class="right-area">
          <template v-if="authStore.isAuthenticated">
            <el-dropdown trigger="click" @command="handleUserCommand">
              <div class="user-info-trigger">
                <el-avatar 
                  :size="36" 
                  :src="authStore.user?.avatar" 
                  class="user-avatar"
                >
                  {{ authStore.user?.username?.charAt(0)?.toUpperCase() || 'U' }}
                </el-avatar>
                <span class="username">{{ authStore.user?.nickname || authStore.user?.username }}</span>
                <el-tag v-if="authStore.user?.role === 'admin'" size="small" type="danger" effect="plain" round>
                  ADMIN
                </el-tag>
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </div>
              
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                  <el-dropdown-item divided command="logout" style="color: #f56c6c;">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>

          <template v-else>
            <div class="auth-buttons">
              <el-button type="primary" link @click="router.push('/login')">登录</el-button>
              <el-button type="primary" round @click="router.push('/register')">注册账号</el-button>
            </div>
          </template>
        </div>
      </div>
    </el-header>
    
    <!-- 主体内容 -->
    <el-main class="main-content">
      <router-view></router-view>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ArrowDown, Setting } from '@element-plus/icons-vue';
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

// 应用加载时
onMounted(() => {
  updateActiveIndex(route);
  if (authStore.isAuthenticated && !authStore.user) {
    authStore.fetchProfile().catch(() => {
      authStore.logout();
    });
  }
});

const updateActiveIndex = (currentRoute: any) => {
  if (currentRoute.path.startsWith('/admin')) {
    // 让 admin 主菜单高亮（或者你可以精确匹配子菜单）
    activeIndex.value = currentRoute.path; 
  } else if (currentRoute.path.startsWith('/news')) {
    activeIndex.value = '/news';
  } else {
    activeIndex.value = currentRoute.path;
  }
};

// 处理主菜单点击（主要用于路由跳转，router 模式下其实会自动处理，这里主要处理特殊逻辑）
const handleSelect = (key: string) => {
  // 这里不需要处理 logout，因为 logout 移到了 dropdown command 中
};

// 处理用户下拉菜单点击
const handleUserCommand = (command: string) => {
  if (command === 'logout') {
    authStore.logout();
    router.push('/');
  } else if (command === 'profile') {
    router.push('/profile');
  }
};
</script>

<style scoped>
.layout-container {
  min-height: 100vh;
  background-color: #f5f7fa; /* 整体背景色偏灰，突出内容卡片 */
}

/* Header 样式 */
.app-header {
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  height: 64px !important; /* 增加一点高度更大气 */
  padding: 0;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

/* Logo 区域 */
.logo-area {
  display: flex;
  align-items: center;
  cursor: pointer;
  margin-right: 40px;
}

.logo-icon {
  font-size: 24px;
  margin-right: 8px;
}

.logo-text {
  font-size: 20px;
  font-weight: 700;
  color: #409EFF;
  letter-spacing: 0.5px;
}

/* 导航菜单重置 */
.nav-menu {
  flex-grow: 1;
  border-bottom: none !important; /* 去除 Element 默认下划线 */
  background: transparent;
}

/* 菜单项样式微调 */
:deep(.el-menu--horizontal > .el-menu-item) {
  font-size: 15px;
  font-weight: 500;
  color: #606266;
  border-bottom: 3px solid transparent;
  transition: all 0.3s;
}

:deep(.el-menu--horizontal > .el-menu-item.is-active) {
  color: #409EFF !important;
  border-bottom-color: #409EFF !important;
  background: transparent !important;
  font-weight: 600;
}

:deep(.el-menu--horizontal > .el-menu-item:hover) {
  background-color: rgba(64, 158, 255, 0.05) !important;
  color: #409EFF;
}

/* 右侧区域 */
.right-area {
  margin-left: 20px;
  display: flex;
  align-items: center;
}

/* 用户下拉触发区 */
.user-info-trigger {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 20px;
  transition: background 0.3s;
}

.user-info-trigger:hover {
  background: #f0f2f5;
}

.user-avatar {
  border: 1px solid #ebeef5;
}

.username {
  margin: 0 8px;
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

/* 登录注册按钮 */
.auth-buttons .el-button {
  font-weight: 500;
}

.main-content {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}
</style>