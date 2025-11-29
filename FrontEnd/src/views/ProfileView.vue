<template>
  <el-container class="profile-container">
    <div class="profile-wrapper">
      <el-tabs type="border-card" class="profile-tabs">
        
        <!-- 个人资料部分 (保持不变) -->
        <el-tab-pane label="个人资料">
          <div class="profile-content">
            <el-form :model="form" label-width="90px" v-loading="loading" class="profile-form">
              <el-row :gutter="40">
                <el-col :xs="24" :sm="8" class="avatar-col">
                  <div class="avatar-preview-wrapper">
                    <el-avatar :size="100" :src="getImageUrl(form.avatar || authStore.user?.avatar||'')" class="main-avatar">
                      {{ authStore.user?.username?.charAt(0)?.toUpperCase() || 'U' }}
                    </el-avatar>
                    <div class="role-badge">
                      <el-tag :type="authStore.user!.role === 'admin' ? 'danger' : 'info'" effect="dark" round>
                        {{ authStore.user!.role === 'admin' ? '管理员' : '普通用户' }}
                      </el-tag>
                    </div>
                     <el-upload
                        class="avatar-uploader"
                        action="#"
                        :show-file-list="false"
                        :http-request="handleAvatarUpload"
                        accept="image/jpeg,image/png,image/gif"
                      >
                        <el-button type="primary" link size="small" style="margin-top: 10px;">更换头像</el-button>
                      </el-upload>
                  </div>
                </el-col>
                <el-col :xs="24" :sm="16">
                  <el-form-item label="昵称">
                    <el-input v-model="form.nickname" maxlength="24" />
                  </el-form-item>
                  <el-form-item label="邮箱">
                    <el-input v-model="form.email" type="email" maxlength="50" />
                  </el-form-item>
                  <el-form-item label="新密码">
                    <el-input v-model="form.password" type="password" show-password placeholder="留空不修改" />
                  </el-form-item>
                  <el-form-item style="margin-top: 30px;">
                    <el-button type="primary" size="large" class="save-btn" @click="handleUpdate">保存修改</el-button>
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- ================= 我的收藏 (核心修复) ================= -->
        <el-tab-pane label="我的收藏">
          <div v-if="favorites.length > 0" class="fav-list">
            <div v-for="fav in favorites" :key="fav.id" class="fav-card-wrapper">
              <el-card class="fav-card" shadow="hover" :body-style="{ padding: '15px' }">
                
                <!-- 情况A: 文章正常存在 -->
                <div v-if="fav.Article && fav.Article.id" class="fav-inner">
                  <div v-if="fav.Article.cover" class="fav-img-box" @click="router.push(`/news/${fav.Article.id}`)">
                    <el-image :src="getImageUrl(fav.Article.cover)" fit="cover" class="fav-img" />
                  </div>
                  <div class="fav-info" @click="router.push(`/news/${fav.Article.id}`)">
                    <h3 class="fav-title">{{ fav.Article.title }}</h3>
                    <p class="fav-preview">{{ fav.Article.preview }}</p>
                    <div class="fav-meta">
                      <span class="fav-date">
                        <el-icon><Calendar /></el-icon> {{ formatDate(fav.Article.CreatedAt) }}
                      </span>
                    </div>
                  </div>
                  <div class="fav-action">
                    <!-- 这里的参数一定是 fav.ID (收藏记录ID)，不是 articleID -->
                    <el-button size="small" type="danger" plain round @click.stop="handleDeleteFavorite(fav.id)">
                      取消收藏
                    </el-button>
                  </div>
                </div>

                <!-- 情况B: 文章已失效 (Article为空 或 id为0) -->
                <div v-else class="fav-inner invalid-item">
                  <div class="fav-img-box invalid-img">
                    <span>已删除</span>
                  </div>
                  <div class="fav-info">
                    <h3 class="fav-title" style="color: #999; text-decoration: line-through;">(文章内容已丢失)</h3>
                    <p class="fav-preview" style="color: #ccc;">该文章已被作者删除，无法查看。</p>
                  </div>
                  <div class="fav-action">
                    <!-- 关键点：即使文章没了，fav.ID 依然存在，必须传这个 ID -->
                    <el-button size="small" type="danger" icon="Delete" @click.stop="handleDeleteFavorite(fav.id)">
                      移除记录
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from '../axios'
import { useAuthStore } from '../store/auth'

// 确保这里的类型定义包含 ID
interface Favorite {
  id: number;        // 收藏记录ID (后端返回的 json:"id" 或 "ID")
  ArticleId: number;
  UserID: number;
  Article?: {        // Article 可能为空
    id: number;
    title: string;
    preview: string;
    cover: string;
    CreatedAt: string;
  }; 
}

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

// 获取收藏列表
const fetchFavorites = async () => {
  try {
    const res = await axios.get('/user/favorites')
    // 打印一下数据，方便在浏览器控制台调试，确认 ID 字段是否存在
    console.log('收藏列表数据:', res.data) 
    favorites.value = res.data as Favorite[]
  } catch (e: any) {
    ElMessage.error('加载收藏列表失败')
  }
}

// 【核心逻辑】通过收藏ID删除
const handleDeleteFavorite = async (favId: number) => {
  // 防御性检查
  if (!favId) {
    ElMessage.error('无法获取收藏ID，请刷新页面重试')
    return
  }
  
  try {
    // 弹窗确认
    await ElMessageBox.confirm('确定要移除这条收藏吗?', '提示', {
      confirmButtonText: '移除',
      cancelButtonText: '取消',
      type: 'warning'
    })

    // 发送 DELETE 请求到后端 /api/favorites/:id
    await axios.delete(`/favorites/${favId}`)
    
    ElMessage.success('已移除')
    // 成功后重新获取列表
    await fetchFavorites() 
  } catch (e: any) {
    if (e !== 'cancel') {
       console.error(e)
       ElMessage.error(e.response?.data?.error || '操作失败')
    }
  }
}

// 初始化表单
const initForm = () => {
  if (authStore.user) {
    form.nickname = authStore.user.nickname || ''
    form.email = authStore.user.email || ''
    form.avatar = authStore.user.avatar || ''
    form.password = ''
  }
}

// 更新资料
const handleUpdate = async () => {
  loading.value = true
  try {
    const payload: any = { nickname: form.nickname, email: form.email, avatar: form.avatar }
    if (form.password) payload.password = form.password
    await axios.put('/user/profile', payload)
    await authStore.fetchProfile()
    ElMessage.success('资料更新成功')
    form.password = ''
  } catch (e: any) {
    ElMessage.error('资料更新失败')
  } finally {
    loading.value = false
  }
}

// 头像上传
const handleAvatarUpload = async (options: any) => {
  const { file } = options;
  const formData = new FormData();
  formData.append('file', file);
  try {
    const res = await axios.post('/user/upload/avatar', formData, { headers: { 'Content-Type': 'multipart/form-data' } });
    form.avatar = res.data.url;
    ElMessage.success('上传成功，请保存');
  } catch (error) {
    ElMessage.error('上传失败');
  }
};

const formatDate = (dateStr?: string) => {
  if (!dateStr || dateStr.startsWith('0001')) return 'N/A'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

const getImageUrl = (path: string) => {
  if (!path) return '';
  // 如果已经是完整的 http 开头链接，直接返回
  if (path.startsWith('http') || path.startsWith('blob:')) return path;
  
  // 1. 定义后端基础地址 (请确认你的 Go 后端确实运行在 3080 端口，Gin 默认通常是 8080)
  const baseUrl = 'http://localhost:3080';
  
  // 2. 智能处理斜杠：如果 path 不以 / 开头，手动加上 /
  const validPath = path.startsWith('/') ? path : '/' + path;
  
  return `${baseUrl}${validPath}`; 
};

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
  min-height: calc(100vh - 64px);
}
.profile-wrapper { width: 100%; max-width: 900px; }
.profile-tabs { border-radius: 8px; background: #fff; box-shadow: 0 2px 12px rgba(0,0,0,0.05); }
.profile-content { padding: 20px 10px; }
.avatar-col { display: flex; justify-content: center; padding-top: 10px; }
.avatar-preview-wrapper { display: flex; flex-direction: column; align-items: center; gap: 15px; }
.main-avatar { border: 4px solid #fff; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.role-badge { margin-top: 5px; }
.save-btn { width: 100%; max-width: 200px; }

/* 收藏列表样式 */
.fav-list { display: flex; flex-direction: column; gap: 16px; padding: 10px 0; }
.fav-card { border-radius: 8px; transition: all 0.2s; }
.fav-card:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(0,0,0,0.08); }
.fav-inner { display: flex; align-items: center; gap: 20px; }
.fav-img-box { width: 140px; height: 90px; flex-shrink: 0; border-radius: 6px; overflow: hidden; cursor: pointer; }
.fav-img { width: 100%; height: 100%; }
.fav-info { flex: 1; min-width: 0; cursor: pointer; }
.fav-title { margin: 0 0 8px; font-size: 18px; color: #303133; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.fav-preview { margin: 0 0 10px; color: #909399; font-size: 14px; height: 3em; overflow: hidden; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; }
.fav-meta { display: flex; align-items: center; font-size: 13px; color: #C0C4CC; }
.fav-action { flex-shrink: 0; }

/* 失效文章样式 */
.invalid-item { opacity: 0.8; background-color: #fafafa; border-radius: 6px; padding: 5px; }
.invalid-img { background-color: #eee; display: flex; align-items: center; justify-content: center; color: #999; font-size: 12px; cursor: default; }
</style>









