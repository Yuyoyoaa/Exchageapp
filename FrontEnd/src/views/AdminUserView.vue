<template>
  <el-container class="admin-container">
    <el-card class="admin-card">
      <template #header>
        <div class="card-header">
          <h2>用户管理</h2>
          <el-button type="primary" @click="refreshUsers">刷新</el-button>
        </div>
      </template>

      <el-table :data="users" v-loading="loading" style="width: 100%">
        <el-table-column prop="ID" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.role === 'admin' ? 'danger' : 'primary'">
              {{ scope.row.role === 'admin' ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="CreatedAt" label="注册时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button 
              size="small" 
              :type="scope.row.role === 'admin' ? 'warning' : 'success'"
              @click="handleChangeRole(scope.row)"
              :disabled="scope.row.username === authStore.user?.username"
            >
              {{ scope.row.role === 'admin' ? '降为用户' : '设为管理员' }}
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="handleDeleteUser(scope.row)"
              :disabled="scope.row.username === authStore.user?.username"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 修改角色确认对话框 -->
    <el-dialog v-model="roleDialogVisible" title="修改用户角色" width="400px">
      <p>确定要将用户 <strong>{{ selectedUser?.username }}</strong> 的角色修改为 <strong>{{ selectedUser?.role === 'admin' ? '普通用户' : '管理员' }}</strong> 吗？</p>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmChangeRole" :loading="roleLoading">确认修改</el-button>
      </template>
    </el-dialog>

    <!-- 删除用户确认对话框 -->
    <el-dialog v-model="deleteDialogVisible" title="删除用户" width="400px">
      <p style="color: #f56c6c;">警告：此操作将永久删除用户 <strong>{{ selectedUser?.username }}</strong>，且无法恢复！</p>
      <template #footer>
        <el-button @click="deleteDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDeleteUser" :loading="deleteLoading">确认删除</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { onMounted, ref } from 'vue';
import axios from '../axios';
import { useAuthStore } from '../store/auth';

interface User {
  ID: number;
  username: string;
  nickname: string;
  email: string;
  role: string;
  avatar: string;
  CreatedAt: string;
}

const authStore = useAuthStore();
const loading = ref(false);
const users = ref<User[]>([]);
const roleDialogVisible = ref(false);
const deleteDialogVisible = ref(false);
const selectedUser = ref<User | null>(null);
const roleLoading = ref(false);
const deleteLoading = ref(false);

const fetchUsers = async () => {
  loading.value = true;
  try {
    const response = await axios.get('/admin/users');
    users.value = response.data;
  } catch (error: any) {
    ElMessage.error('获取用户列表失败: ' + (error.response?.data?.error || error.message));
  } finally {
    loading.value = false;
  }
};

const refreshUsers = () => {
  fetchUsers();
};

const handleChangeRole = (user: User) => {
  selectedUser.value = user;
  roleDialogVisible.value = true;
};

const confirmChangeRole = async () => {
  if (!selectedUser.value) return;
  
  roleLoading.value = true;
  try {
    const newRole = selectedUser.value.role === 'admin' ? 'user' : 'admin';
    await axios.patch(`/admin/users/${selectedUser.value.ID}/role`, { role: newRole });
    ElMessage.success('角色修改成功');
    roleDialogVisible.value = false;
    await fetchUsers();
  } catch (error: any) {
    ElMessage.error('修改角色失败: ' + (error.response?.data?.error || error.message));
  } finally {
    roleLoading.value = false;
  }
};

const handleDeleteUser = (user: User) => {
  selectedUser.value = user;
  deleteDialogVisible.value = true;
};

const confirmDeleteUser = async () => {
  if (!selectedUser.value) return;
  
  deleteLoading.value = true;
  try {
    await axios.delete(`/admin/users/${selectedUser.value.ID}`);
    ElMessage.success('用户删除成功');
    deleteDialogVisible.value = false;
    await fetchUsers();
  } catch (error: any) {
    ElMessage.error('删除用户失败: ' + (error.response?.data?.error || error.message));
  } finally {
    deleteLoading.value = false;
  }
};

const formatDate = (dateString: string) => {
  if (!dateString) return '';
  return new Date(dateString).toLocaleString('zh-CN');
};

onMounted(() => {
  fetchUsers();
});
</script>

<style scoped>
.admin-container {
  padding: 20px;
  min-height: calc(100vh - 60px);
  background-color: #f5f7fa;
}

.admin-card {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>