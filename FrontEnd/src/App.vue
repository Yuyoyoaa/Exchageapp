<template>
  <el-container>
    <el-header>
      <el-menu :default-active="activeIndex" class="el-menu-demo" mode="horizontal" :ellipsis="true" @select="handleSelect">
        <el-menu-item index="home">首页</el-menu-item>
        <el-menu-item index="currencyExchange">兑换货币</el-menu-item>
        <el-menu-item index="news">查看新闻</el-menu-item>
        
        <el-menu-item index="login" v-if="!authStore.isAuthenticated">登录</el-menu-item>
        <el-menu-item index="register" v-if="!authStore.isAuthenticated">注册</el-menu-item>
        
        <el-menu-item index="profile" v-if="authStore.isAuthenticated">
          个人资料 <span v-if="authStore.user?.nickname">({{ authStore.user.nickname }})</span>
        </el-menu-item>
        <el-menu-item index="logout" v-if="authStore.isAuthenticated">退出</el-menu-item>
      </el-menu>
    </el-header>
    <el-main>
      <router-view></router-view>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from './store/auth';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const activeIndex = ref('home');

// 监听路由变化高亮菜单
watch(route, (newRoute) => {
  // 将路由名称转换为小写以匹配 index
  activeIndex.value = newRoute.name?.toString().toLowerCase() || 'home';
});

// 应用加载时，如果已登录但没有用户信息，尝试获取一次
onMounted(() => {
  if (authStore.isAuthenticated && !authStore.user) {
    authStore.fetchProfile();
  }
});

const handleSelect = (key: string) => {
  if (key === 'logout') {
    authStore.logout();
    router.push({ name: 'Home' });
  } else if (key === 'profile') {
    router.push({ name: 'Profile' });
  } else if (key === 'home') {
    router.push({ name: 'Home' });
  } else {
    // 简单的首字母大写转换 (login -> Login, etc.)
    // 注意：currencyExchange 需要特殊处理或者保持你的命名一致性
    // 建议直接根据 key switch case 跳转最稳妥
    switch(key) {
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
  }
};
</script>

<style scoped>
.el-menu-demo {
  line-height: 60px;
}
</style>