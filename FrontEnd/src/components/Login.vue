<template>  
  <div class="auth-container">  
    <el-form :model="form" class="auth-form" label-position="top" @submit.prevent="login">  
      <h2 style="text-align: center; margin-bottom: 20px;">用户登录</h2>
      <el-form-item label="用户名">  
        <el-input v-model="form.username" placeholder="请输入用户名" />  
      </el-form-item>  
      <el-form-item label="密码">  
        <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />  
      </el-form-item>  
      <el-form-item>  
        <el-button type="primary" native-type="submit" style="width: 100%" :loading="loading">登录</el-button>  
      </el-form-item>  
    </el-form>  
  </div>  
</template>  
  
<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../store/auth';

const form = ref({
  username: '',
  password: '',
});

const loading = ref(false);
const authStore = useAuthStore();
const router = useRouter();

const login = async () => {
  loading.value = true;
  try {
    await authStore.login(form.value.username, form.value.password);
    ElMessage.success('登录成功');
    router.push({ name: 'News' });
  } catch (error: any) {
    // 这里的 error.message 已经是 store 中处理过的后端 error 字段
    ElMessage.error(error.message || '登录失败，请检查网络或重试。');
  } finally {
    loading.value = false;
  }
};
</script>
  
<style scoped>
.auth-container {  
  display: flex;  
  justify-content: center;  
  align-items: center;  
  height: 80vh; 
  background-color: #f5f5f5; 
  padding: 20px;  
  box-sizing: border-box; 
}  
  
.auth-form {  
  width: 100%;  
  max-width: 360px; 
  padding: 30px;  
  background-color: #fff;  
  border-radius: 8px;  
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);  
}  
</style>
  