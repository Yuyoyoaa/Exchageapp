<template>
  <el-container class="profile-container">
    <div class="profile-wrapper">
      <el-tabs type="border-card">

        <!-- ================= 个人资料 ================= -->
        <el-tab-pane label="个人资料">
          <el-form :model="form" label-width="100px" v-loading="loading" class="profile-form">
            
            <!-- 昵称 -->
            <el-form-item label="昵称">
              <el-input v-model="form.nickname" placeholder="请输入昵称" />
            </el-form-item>

            <!-- 邮箱 -->
            <el-form-item label="邮箱">
              <el-input v-model="form.email" placeholder="请输入邮箱" />
            </el-form-item>

            <!-- 头像 URL -->
            <el-form-item label="头像 URL">
              <el-input
                v-model="form.avatar"
                placeholder="请输入头像 URL"
              />
            </el-form-item>

            <!-- 修改密码 -->
            <el-form-item label="修改密码">
              <el-input v-model="form.password" type="password" placeholder="留空则不修改" />
            </el-form-item>

            <!-- 保存按钮 -->
            <el-form-item>
              <el-button type="primary" @click="handleUpdate">保存</el-button>
            </el-form-item>

          </el-form>
        </el-tab-pane>

        <!-- ================= 我的收藏 ================= -->
        <el-tab-pane label="我的收藏">
          <div v-if="favorites.length > 0" class="fav-list">
            <el-card 
              v-for="fav in favorites"
              :key="fav.ID"
              class="fav-item"
              shadow="hover"
            >
              <div class="fav-content">
                <div class="fav-info" @click="router.push(`/news/${fav.Article.id}`)">
                  <h3>{{ fav.Article.title }}</h3>
                  <p>{{ fav.Article.preview }}</p>
                  <span class="fav-date">{{ formatDate(fav.Article.CreatedAt) }}</span>
                </div>

                <el-button
                  size="small"
                  type="primary"
                  @click.stop="toggleFavorite(fav.Article.id)"
                >
                  取消收藏
                </el-button>

                <div v-if="fav.Article.cover" class="fav-img">
                  <el-image :src="fav.Article.cover" fit="cover" />
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

// 初始化表单
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

// 格式化日期
const formatDate = (dateStr?: string) => {
  if (!dateStr || dateStr.startsWith('0001')) return 'N/A'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

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
  min-height: 80vh;
}
.profile-wrapper { width: 100%; max-width: 800px; }
.profile-form { max-width: 500px; padding: 20px; }

/* ================= 收藏列表样式 ================= */
.fav-list { display: flex; flex-direction: column; gap: 10px; }
.fav-item { cursor: pointer; border-radius: 4px; }
.fav-item:hover { background-color: #fcfcfc; }
.fav-content { display: flex; justify-content: space-between; align-items: center; padding: 10px; cursor: default; }
.fav-info { flex: 1; min-width: 0; padding-right: 15px; }
.fav-info h3 { margin: 0 0 8px; font-size: 1.1rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.fav-info p { color: #909399; height: 3em; margin: 0 0 8px; overflow: hidden; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; }
.fav-date { font-size: 0.8rem; color: #C0C4CC; }
.fav-img { width: 120px; height: 80px; border-radius: 4px; overflow: hidden; margin-left: 10px; }
</style>









