<template>
  <el-container class="profile-container">
    <div class="profile-wrapper">
      <el-tabs type="border-card" class="profile-tabs">
        
        <!-- ================= 个人资料 ================= -->
        <el-tab-pane label="个人资料">
          <div class="profile-content">
            <el-form :model="form" label-width="90px" v-loading="loading" class="profile-form">
              <el-row :gutter="40">
                <!-- 左侧：头像预览 -->
                <el-col :xs="24" :sm="8" class="avatar-col">
                  <div class="avatar-preview-wrapper">
                    <el-avatar :size="100" :src="form.avatar || authStore.user?.avatar" class="main-avatar">
                      {{ authStore.user?.username?.charAt(0)?.toUpperCase() || 'U' }}
                    </el-avatar>
                    <div class="role-badge">
                      <el-tag :type="authStore.user!.role === 'admin' ? 'danger' : 'info'" effect="dark" round>
                        {{ authStore.user!.role === 'admin' ? '管理员' : '普通用户' }}
                      </el-tag>
                    </div>
                  </div>
                </el-col>

                <!-- 右侧：表单信息 -->
                <el-col :xs="24" :sm="16">
                  <el-form-item label="昵称">
                    <el-input v-model="form.nickname" maxlength="24" placeholder="请输入您的昵称" />
                  </el-form-item>
                  
                  <el-form-item label="邮箱">
                    <el-input v-model="form.email" type="email" maxlength="50" placeholder="请输入您的邮箱" />
                  </el-form-item>

                  <el-form-item label="头像URL">
                    <el-input v-model="form.avatar" maxlength="255" placeholder="输入图片链接" />
                  </el-form-item>

                  <el-divider content-position="left" style="margin: 30px 0 20px;">安全设置</el-divider>

                  <el-form-item label="新密码">
                    <el-input 
                      v-model="form.password" 
                      type="password" 
                      show-password 
                      autocomplete="new-password" 
                      placeholder="如果不修改请留空" 
                    />
                  </el-form-item>

                  <el-form-item style="margin-top: 30px;">
                    <el-button type="primary" size="large" class="save-btn" @click="handleUpdate">保存修改</el-button>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- ================= 我的收藏 ================= -->
        <el-tab-pane label="我的收藏">
          <div v-if="favorites.length > 0" class="fav-list">
            <div v-for="fav in favorites" :key="fav.ID" class="fav-card-wrapper">
              <el-card class="fav-card" shadow="hover" :body-style="{ padding: '15px' }">
                <div class="fav-inner">
                  <!-- 文章封面 -->
                  <div v-if="fav.Article.cover" class="fav-img-box" @click="router.push(`/news/${fav.Article.id}`)">
                    <el-image :src="fav.Article.cover" fit="cover" class="fav-img" />
                  </div>

                  <!-- 文章信息 -->
                  <div class="fav-info" @click="router.push(`/news/${fav.Article.id}`)">
                    <h3 class="fav-title">{{ fav.Article.title }}</h3>
                    <p class="fav-preview">{{ fav.Article.preview }}</p>
                    <div class="fav-meta">
                      <span class="fav-date">
                        <el-icon><Calendar /></el-icon> {{ formatDate(fav.Article.CreatedAt) }}
                      </span>
                    </div>
                  </div>

                  <!-- 操作按钮 -->
                  <div class="fav-action">
                    <el-button
                      size="small"
                      type="danger"
                      plain
                      round
                      @click.stop="toggleFavorite(fav.Article.id)"
                    >
                      取消收藏
                    </el-button>
                  </div>
                </div>
              </el-card>
            </div>
          </div>
          <el-empty v-else description="暂无收藏文章" :image-size="120" />
        </el-tab-pane>

      </el-tabs>
    </div>
  </el-container>
</template>

<script setup lang="ts">
import { Calendar } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from '../axios'
import { useAuthStore } from '../store/auth'
// 使用全局 Article 类型
import type { Favorite } from '../types/Article'

const authStore = useAuthStore()
const router = useRouter()
const loading = ref(false)
const favorites = ref<Favorite[]>([])

const form = reactive({
  nickname: '',
  email: '',
  avatar: '',
  password: '',
})

// 初始化表单数据
const initForm = () => {
  if (authStore.user) {
    form.nickname = authStore.user.nickname || ''
    form.email = authStore.user.email || ''
    form.avatar = authStore.user.avatar || ''
    form.password = ''
  }
}

// 获取收藏列表
const fetchFavorites = async () => {
  try {
    const res = await axios.get('/user/favorites')
    favorites.value = res.data as Favorite[]
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '加载收藏列表失败')
  }
}

// 收藏/取消收藏
const toggleFavorite = async (articleID: number) => {
  if (!articleID) return
  try {
    const res = await axios.post(`/articles/${articleID}/favorite`)
    ElMessage.success(res.data.message)
    await fetchFavorites()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '操作失败')
  }
}

// 更新个人资料
const handleUpdate = async () => {
  loading.value = true
  try {
    const payload: any = {
      nickname: form.nickname,
      email: form.email,
      avatar: form.avatar,
    }
    if (form.password) payload.password = form.password

    await axios.put('/user/profile', payload)
    await authStore.fetchProfile()
    ElMessage.success('资料更新成功')
    form.password = ''
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '资料更新失败')
  } finally {
    loading.value = false
  }
}

// 日期格式化
const formatDate = (dateStr?: string) => {
  if (!dateStr || dateStr.startsWith('0001')) return 'N/A'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

// 首次加载
onMounted(async () => {
  if (authStore.isAuthenticated) {
    loading.value = true
    try {
      await authStore.fetchProfile()
      initForm()
      await fetchFavorites()
    } finally {
      loading.value = false
    }
  }
})
</script>

<style scoped>
.profile-container {
  display: flex;
  justify-content: center;
  padding: 40px 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 64px); /* 减去头部高度 */
}

.profile-wrapper { 
  width: 100%; 
  max-width: 900px; 
}

/* Tabs 样式重写 */
.profile-tabs {
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  background: #fff;
}

:deep(.el-tabs__header) {
  background-color: #fafafa;
}

:deep(.el-tabs__item) {
  font-size: 16px;
  height: 50px;
  line-height: 50px;
}

:deep(.el-tabs__item.is-active) {
  background-color: #fff;
  font-weight: 600;
}

.profile-content {
  padding: 20px 10px;
}

/* 头像区域 */
.avatar-col {
  display: flex;
  justify-content: center;
  padding-top: 10px;
}

.avatar-preview-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.main-avatar {
  border: 4px solid #fff;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

/* 按钮样式 */
.save-btn {
  width: 100%;
  max-width: 200px;
}

/* 收藏卡片样式 */
.fav-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 10px 0;
}

.fav-card-wrapper {
  width: 100%;
}

.fav-card {
  border: none;
  border-radius: 8px;
  transition: transform 0.2s, box-shadow 0.2s;
}

.fav-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0,0,0,0.08);
}

.fav-inner {
  display: flex;
  align-items: center;
  gap: 20px;
}

.fav-img-box {
  width: 140px;
  height: 90px;
  flex-shrink: 0;
  border-radius: 6px;
  overflow: hidden;
  cursor: pointer;
}

.fav-img {
  width: 100%;
  height: 100%;
  transition: transform 0.3s;
}

.fav-card:hover .fav-img {
  transform: scale(1.05);
}

.fav-info {
  flex: 1;
  min-width: 0; /* 防止 flex 子项溢出 */
  cursor: pointer;
}

.fav-title {
  margin: 0 0 8px;
  font-size: 18px;
  color: #303133;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fav-preview {
  margin: 0 0 10px;
  color: #909399;
  font-size: 14px;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.fav-meta {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: #C0C4CC;
}

.fav-date .el-icon {
  vertical-align: -1px;
  margin-right: 4px;
}

.fav-action {
  flex-shrink: 0;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .fav-inner {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .fav-img-box {
    width: 100%;
    height: 160px;
  }
  
  .fav-action {
    align-self: flex-end;
    width: 100%;
    text-align: right;
  }

  .avatar-col {
    margin-bottom: 30px;
  }
}
</style>









