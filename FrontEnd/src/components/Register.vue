<template>
  <div class="auth-container">
    <el-form :model="form" class="auth-form" label-position="top" @submit.prevent="register">
      <h2 style="text-align: center; margin-bottom: 20px;">用户注册</h2>
      
      <el-form-item label="用户名 (必填)" required>
        <el-input v-model="form.username" placeholder="唯一用户名" />
      </el-form-item>
      
      <el-form-item label="密码 (必填)" required>
        <el-input v-model="form.password" type="password" placeholder="8位以上，含大小写字母和数字" show-password />
      </el-form-item>

      <el-form-item label="昵称">
        <el-input v-model="form.nickname" placeholder="您的昵称" />
      </el-form-item>

      <el-form-item label="邮箱">
        <el-input v-model="form.email" placeholder="example@email.com" />
      </el-form-item>


      <el-button type="primary" native-type="submit" style="width: 100%" :loading="loading">注册</el-button>
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
  nickname: '',
  email: '',
  avatar: ''
});

const loading = ref(false);
const authStore = useAuthStore();
const router = useRouter();

const register = async () => {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('用户名和密码是必填项');
    return;
  }

  loading.value = true;
  try {
    await authStore.register({
      username: form.value.username,
      password: form.value.password,
      nickname: form.value.nickname || undefined,
      email: form.value.email || undefined,
      avatar: form.value.avatar || undefined
    });
    ElMessage.success('注册成功！');
    router.push({ name: 'News' });
  } catch (error: any) {
    ElMessage.error(error.message || '注册失败，请重试。');
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
  min-height: 80vh; 
  background-color: #f5f5f5; 
  padding: 20px;  
  box-sizing: border-box; 
}  
  
.auth-form {  
  width: 100%;  
  max-width: 400px; 
  padding: 30px;  
  background-color: #fff;  
  border-radius: 8px;  
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);  
}  
</style>
  