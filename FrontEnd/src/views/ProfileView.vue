<template>
  <el-container class="profile-container">
    <el-card class="profile-card">
      <template #header>
        <div class="card-header">
          <span>个人资料</span>
        </div>
      </template>
      
      <el-form :model="form" label-width="100px" v-loading="loading">
        <el-form-item label="用户名">
          <el-input v-model="authStore.user!.username" disabled />
        </el-form-item>
        <el-form-item label="角色">
          <el-tag>{{ authStore.user?.role }}</el-tag>
        </el-form-item>

        <el-form-item label="昵称">
          <el-input v-model="form.nickname" placeholder="设置昵称" />
        </el-form-item>

        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="绑定邮箱" />
        </el-form-item>

        <el-form-item label="头像URL">
          <el-input v-model="form.avatar" placeholder="输入图片地址" />
          <div v-if="form.avatar" class="avatar-preview">
            <el-avatar :size="50" :src="form.avatar" />
          </div>
        </el-form-item>

        <el-divider content-position="left">修改密码 (留空则不修改)</el-divider>

        <el-form-item label="新密码">
          <el-input 
            v-model="form.password" 
            type="password" 
            placeholder="8位以上，含大小写字母和数字" 
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleUpdate">保存修改</el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </el-container>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { useAuthStore } from '../store/auth';

const authStore = useAuthStore();
const loading = ref(false);

const form = reactive({
  nickname: '',
  email: '',
  avatar: '',
  password: ''
});

// 初始化表单数据
const initForm = () => {
  if (authStore.user) {
    form.nickname = authStore.user.nickname || '';
    form.email = authStore.user.email || '';
    form.avatar = authStore.user.avatar || '';
    form.password = ''; // 密码默认不回显
  }
};

onMounted(async () => {
  loading.value = true;
  // 确保获取最新数据
  await authStore.fetchProfile();
  initForm();
  loading.value = false;
});

const handleUpdate = async () => {
  // 构建 payload，如果密码为空字符串则不发送 password 字段
  const payload: any = {
    nickname: form.nickname,
    email: form.email,
    avatar: form.avatar
  };
  
  if (form.password) {
    payload.password = form.password;
  }

  try {
    loading.value = true;
    await authStore.updateProfile(payload);
    ElMessage.success('个人资料更新成功');
    form.password = ''; // 更新成功后清空密码框
  } catch (error: any) {
    ElMessage.error(error.message || '更新失败');
  } finally {
    loading.value = false;
  }
};

const resetForm = () => {
  initForm();
};
</script>

<style scoped>
.profile-container {
  display: flex;
  justify-content: center;
  padding: 40px 20px;
}
.profile-card {
  width: 100%;
  max-width: 600px;
}
.avatar-preview {
  margin-top: 10px;
}
</style>