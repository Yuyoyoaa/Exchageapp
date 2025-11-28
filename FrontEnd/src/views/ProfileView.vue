<template>
  <el-container class="profile-container">
    <div class="profile-wrapper">
      <el-tabs type="border-card">

        <!-- ================= 个人资料 ================= -->
        <el-tab-pane label="个人资料">
          <el-form :model="form" label-width="100px" v-loading="loading" class="profile-form">

            <!-- 用户名 -->
            <el-form-item label="用户名">
              <el-input v-model="authStore.user!.username" disabled />
            </el-form-item>

            <!-- 身份 -->
            <el-form-item label="身份">
              <el-tag :type="authStore.user!.role === 'admin' ? 'danger' : 'success'">
                {{ authStore.user!.role === 'admin' ? '管理员' : '普通用户' }}
              </el-tag>
            </el-form-item>

            <!-- 昵称 -->
            <el-form-item label="昵称">
              <el-input v-model="form.nickname" />
            </el-form-item>

            <!-- 邮箱 -->
            <el-form-item label="邮箱">
              <el-input v-model="form.email" />
            </el-form-item>

            <!-- ================= 头像上传 ================= -->
            <el-form-item label="头像">
              <div class="avatar-upload-container">
                <el-upload
                  class="avatar-uploader"
                  action="/user/upload/avatar"
                  :show-file-list="false"
                  :before-upload="beforeAvatarUpload"
                  :on-success="handleAvatarSuccess"
                  :on-change="handleAvatarChange"
                  accept="image/*"
                  :auto-upload="false"
                  ref="avatarUploader"
                >
                  <div class="avatar-wrapper" style="cursor:pointer;">
                    <img v-if="avatarPreview" :src="avatarPreview" class="avatar" />
                    <div v-else class="avatar-placeholder">
                      <i class="el-icon-plus"></i>
                      <span>选择图片</span>
                    </div>
                  </div>
                </el-upload>
                
                <!-- 预览和操作按钮 -->
                <div v-if="avatarFile" class="preview-section">
                  <span class="preview-label">新头像预览:</span>
                  <img :src="avatarPreview" class="preview-image" />
                  <div class="upload-actions">
                    <el-button type="primary" size="small" @click="submitAvatarUpload" :loading="uploading">
                      确认上传
                    </el-button>
                    <el-button size="small" @click="cancelAvatarUpload">
                      取消
                    </el-button>
                  </div>
                </div>
                
                <div v-else class="preview-section">
                  <span class="preview-label">当前头像:</span>
                  <img :src="form.avatar" class="preview-image" />
                </div>
              </div>
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
const uploading = ref(false)
const avatarUploader = ref()
const avatarFile = ref<File | null>(null)
const avatarPreview = ref('')
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
    avatarPreview.value = authStore.user.avatar || ''
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

// 上传前检查文件
const beforeAvatarUpload = (file: File) => {
  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isImage) ElMessage.error('上传头像只能是图片文件')
  if (!isLt2M) ElMessage.error('图片大小不能超过 2MB')
  return isImage && isLt2M
}

// 处理头像文件选择
const handleAvatarChange = (file: any) => {
  // 验证文件
  if (!beforeAvatarUpload(file.raw)) {
    return false
  }
  
  avatarFile.value = file.raw
  // 生成预览
  avatarPreview.value = URL.createObjectURL(file.raw)
}

// 提交头像上传
const submitAvatarUpload = async () => {
  if (!avatarFile.value) {
    ElMessage.warning('请先选择头像文件')
    return
  }
  
  uploading.value = true
  try {
    // 手动触发上传
    if (avatarUploader.value) {
      // 这里需要手动提交文件
      const formData = new FormData()
      formData.append('avatar', avatarFile.value)
      
      const res = await axios.post('/user/upload/avatar', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
      
      if (res.data.avatar) {
        form.avatar = res.data.avatar
        avatarPreview.value = res.data.avatar
        avatarFile.value = null
        ElMessage.success('头像上传成功')
      } else {
        ElMessage.error('头像上传失败')
      }
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '头像上传失败')
  } finally {
    uploading.value = false
  }
}

// 取消头像上传
const cancelAvatarUpload = () => {
  avatarFile.value = null
  avatarPreview.value = form.avatar
  ElMessage.info('已取消头像上传')
}

// 上传成功回调
const handleAvatarSuccess = (res: any) => {
  // 这个函数现在可能不会被调用，因为我们改为手动上传
  // 但保留以防其他逻辑需要
  if (res.avatar) {
    form.avatar = res.avatar
    ElMessage.success('头像上传成功')
  } else {
    ElMessage.error('头像上传失败')
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

/* ================= 头像上传样式 ================= */
.avatar-upload-container {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 15px;
}

.avatar-uploader {
  display: inline-block;
  width: 100px;
  height: 100px;
  border: 2px dashed #d9d9d9;
  border-radius: 50%;
  cursor: pointer;
  overflow: hidden;
  position: relative;
  transition: all 0.3s;
}

.avatar-uploader:hover {
  border-color: #409EFF;
}

.avatar-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f5f7fa;
}

.avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: #c0c4cc;
}

.avatar-placeholder i {
  font-size: 24px;
  margin-bottom: 5px;
}

.avatar-placeholder span {
  font-size: 12px;
}

.preview-section {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 5px;
}

.preview-label {
  font-size: 14px;
  color: #606266;
}

.preview-image {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #e0e0e0;
}

.upload-actions {
  display: flex;
  gap: 8px;
}

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







