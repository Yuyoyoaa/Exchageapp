<template>
  <el-container class="profile-container">
    <div class="profile-wrapper">
      <el-tabs type="border-card">
        <el-tab-pane label="个人资料">
          <el-form :model="form" label-width="100px" v-loading="loading" class="profile-form">
             <el-form-item label="用户名">
                <el-input v-model="authStore.user!.username" disabled />
             </el-form-item>
             <el-form-item label="身份">
                <el-tag :type="authStore.user!.role === 'admin' ? 'danger' : 'success'">
                  {{ authStore.user!.role === 'admin' ? '管理员' : '普通用户' }}
                </el-tag>
             </el-form-item>
             <el-form-item label="昵称">
                <el-input v-model="form.nickname" />
             </el-form-item>
             <el-form-item label="邮箱">
                <el-input v-model="form.email" />
             </el-form-item>
             <el-form-item label="头像URL">
                <el-input v-model="form.avatar" />
             </el-form-item>
             <el-form-item label="修改密码">
                <el-input v-model="form.password" type="password" placeholder="留空则不修改" />
             </el-form-item>
             <el-form-item>
                <el-button type="primary" @click="handleUpdate">保存</el-button>
             </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="我的收藏">
          <div v-if="favorites.length > 0" class="fav-list">
            <el-card 
              v-for="fav in favorites" 
              :key="fav.ID" 
              class="fav-item" 
              shadow="hover"
            >
              <div class="fav-content" @click="router.push(`/news/${fav.ID}`)">
                <div class="fav-info">
                  <h3>{{ fav.Title }}</h3>
                  <p>{{ fav.Preview }}</p>
                  <span class="fav-date">{{ formatDate(fav.CreatedAt) }}</span>
                </div>
                <div v-if="fav.Cover" class="fav-img">
                   <el-image :src="fav.Cover" fit="cover" />
                </div>
              </div>
            </el-card>
          </div>
          <el-empty v-else description="暂无收藏文章" />
        </el-tab-pane>
      </el-tabs>
    </div>
  </el-container>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import axios from '../axios';
import { useAuthStore } from '../store/auth';

const authStore = useAuthStore();
const router = useRouter();

const loading = ref(false);
const favorites = ref<any[]>([]);

const form = reactive({
  nickname: '',
  email: '',
  avatar: '',
  password: '',
});

const initForm = () => {
  if (authStore.user) {
    form.nickname = authStore.user.nickname || '';
    form.email = authStore.user.email || '';
    form.avatar = authStore.user.avatar || '';
    form.password = ''; 
  }
};

const fetchFavorites = async () => {
  try {
    const res = await axios.get('/user/favorites');
    favorites.value = res.data;
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '加载收藏列表失败');
  }
};

const handleUpdate = async () => {
  loading.value = true;
  try {
    const payload: any = {
      nickname: form.nickname,
      email: form.email,
      avatar: form.avatar,
    };
    if (form.password) {
      payload.password = form.password;
    }

    await axios.put('/user/profile', payload);
    await authStore.fetchProfile(); // Reload user info
    ElMessage.success('资料更新成功');
    form.password = ''; // Clear password field after successful update
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '资料更新失败');
  } finally {
    loading.value = false;
  }
};

const formatDate = (dateStr: string) => {
  if (!dateStr || dateStr.startsWith('0001')) return 'N/A';
  return new Date(dateStr).toLocaleDateString('zh-CN');
}

onMounted(async () => {
  if (authStore.isAuthenticated) {
    loading.value = true;
    await authStore.fetchProfile();
    initForm();
    await fetchFavorites(); // Fetch favorites on mount
    loading.value = false;
  }
});
</script>

<style scoped>
.profile-container { display: flex; justify-content: center; padding: 40px 20px; background: #f5f7fa; min-height: 80vh; }
.profile-wrapper { width: 100%; max-width: 800px; }
.profile-form { max-width: 500px; padding: 20px; }

/* My Favorites List Styles */
.fav-list { display: flex; flex-direction: column; gap: 10px; }
.fav-item { margin-bottom: 10px; cursor: pointer; border-radius: 4px; }
.fav-item:hover { background-color: #fcfcfc; }

.fav-content { display: flex; justify-content: space-between; align-items: flex-start; padding: 10px; }
.fav-info { flex: 1; min-width: 0; padding-right: 15px; }

.fav-info h3 { margin: 0 0 8px 0; color: #303133; font-size: 1.1rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.fav-info p { color: #909399; font-size: 0.9rem; margin: 0 0 8px 0; line-height: 1.5; height: 3em; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; }
.fav-date { font-size: 0.8rem; color: #C0C4CC; }

.fav-img { width: 120px; height: 80px; flex-shrink: 0; border-radius: 4px; overflow: hidden; }
.fav-img .el-image { width: 100%; height: 100%; object-fit: cover; }
</style>