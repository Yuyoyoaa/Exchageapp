<template>
  <el-container class="news-detail-container">
    <el-main>
      <!-- 加载状态：骨架屏 -->
      <div v-if="!article" class="content-wrapper loading-wrapper">
        <el-skeleton :rows="10" animated />
      </div>

      <!-- 内容区域 -->
      <div class="content-wrapper" v-else>
        <!-- 文章卡片 -->
        <el-card class="article-card" shadow="never">
          <div class="article-header">
            <div class="category-tag" v-if="article.categoryId">
              <el-tag effect="dark" size="small">资讯</el-tag>
            </div>
            <h1 class="article-title">{{ article.title }}</h1>
            <div class="meta-row">
              <span class="meta-item">
                <el-icon><Calendar /></el-icon> {{ formatDate(article.createdAt) }}
              </span>
              <span class="meta-item">
                <el-icon><View /></el-icon> {{ article.viewsCount }} 阅读
              </span>
              <span class="meta-item">
                <el-icon><StarFilled /></el-icon> {{ article.likesCount }} 点赞
              </span>
            </div>
          </div>

          <div v-if="article.cover" class="article-cover-wrapper">
            <el-image :src="article.cover" fit="cover" class="cover-image" lazy />
          </div>

          <el-divider class="content-divider" />

          <div class="article-content typograhy" v-html="article.content"></div>

          <!-- 底部交互区 -->
          <div class="article-footer">
            <div class="action-buttons">
              <el-tooltip content="点赞支持作者" placement="top">
                <el-button 
                  :type="hasLiked ? 'primary' : 'info'" 
                  :plain="!hasLiked"
                  round 
                  size="large" 
                  @click="likeArticle"
                  class="action-btn"
                >
                  <!-- 修正：根据状态切换图标，未点赞用空心Star，点赞用实心StarFilled -->
                  <el-icon class="mr-1">
                    <component :is="hasLiked ? 'StarFilled' : 'Star'" />
                  </el-icon>
                  {{ hasLiked ? '已点赞' : '点赞' }}
                </el-button>
              </el-tooltip>

              <el-tooltip content="收藏以便稍后阅读" placement="top">
                <el-button 
                  :type="isFavorited ? 'warning' : 'info'" 
                  :plain="!isFavorited"
                  round 
                  size="large" 
                  @click="toggleFavorite"
                  class="action-btn"
                >
                  <!-- 修正：使用 CollectionTag 图标代表收藏 -->
                  <el-icon class="mr-1"><CollectionTag /></el-icon>
                  {{ isFavorited ? '已收藏' : '收藏' }}
                </el-button>
              </el-tooltip>
            </div>
          </div>
        </el-card>

        <!-- 评论区 -->
        <div class="comments-section">
          <div class="section-header">
            <h3>全部评论 <span class="comment-count">({{ comments.length }})</span></h3>
          </div>

          <!-- 新评论输入框 -->
          <el-card class="comment-input-card" shadow="hover" :body-style="{ padding: '20px' }">
            <div v-if="authStore.isAuthenticated" class="comment-input-area">
              <div class="user-avatar-mini">
                 <el-avatar :size="40" :src="authStore.user?.avatar">
                   {{ authStore.user?.username?.charAt(0)?.toUpperCase() }}
                 </el-avatar>
              </div>
              <div class="input-wrapper">
                <el-input
                  v-model="newComment"
                  type="textarea"
                  :rows="3"
                  resize="none"
                  placeholder="发表您的友善评论..."
                />
                <div class="input-footer">
                  <el-button type="primary" @click="submitComment(null)" :disabled="!newComment.trim()">发表评论</el-button>
                </div>
              </div>
            </div>
            
            <el-result v-else icon="info" title="登录后发表评论" sub-title="参与讨论需要先登录您的账号">
              <template #extra>
                <el-button type="primary" @click="router.push('/login')">去登录</el-button>
              </template>
            </el-result>
          </el-card>

          <!-- 评论列表 -->
          <div class="comment-list">
            <transition-group name="list">
              <div v-for="comment in comments" :key="comment.id" class="comment-item">
                <div class="comment-avatar-col">
                  <el-avatar :size="48" :src="comment.user?.avatar" class="comment-avatar">
                    {{ comment.user?.avatar?.charAt(0) || comment.userName?.charAt(0) || 'U' }}
                  </el-avatar>
                </div>

                <div class="comment-content-col">
                  <div class="comment-header-row">
                    <span class="username">
                      {{ comment.userName || '匿名用户' }}
                      <el-tag v-if="comment.user?.role === 'admin'" size="small" type="danger" effect="plain" round style="transform: scale(0.8);">管理员</el-tag>
                    </span>
                    <span class="timestamp">{{ formatDate(comment.createdAt) }}</span>
                  </div>

                  <div class="comment-text">
                    <span v-if="comment.parentId" class="reply-target">
                      回复 @{{ getParentUser(comment.parentId) }} :
                    </span>
                    {{ comment.content }}
                  </div>

                  <div class="comment-actions-row">
                    <el-button link type="primary" size="small" v-if="authStore.isAuthenticated" @click="toggleReply(comment.id)">
                      <el-icon><ChatLineRound /></el-icon> 回复
                    </el-button>
                    <el-popconfirm 
                      title="确定删除这条评论吗？" 
                      @confirm="deleteComment(comment.id)"
                      v-if="canDelete(comment)"
                    >
                      <template #reference>
                        <el-button link type="danger" size="small">
                          <el-icon><Delete /></el-icon> 删除
                        </el-button>
                      </template>
                    </el-popconfirm>
                  </div>

                  <!-- 回复框 -->
                  <div v-if="replyBoxId === comment.id" class="reply-box fade-in">
                    <el-input v-model="replyContents[comment.id]" type="textarea" :rows="2" placeholder="回复内容..." />
                    <div class="reply-footer">
                      <el-button size="small" @click="replyBoxId = null">取消</el-button>
                      <el-button type="primary" size="small" @click="submitComment(comment.id)">发送</el-button>
                    </div>
                  </div>
                </div>
              </div>
            </transition-group>
            
            <el-empty v-if="comments.length === 0" description="暂无评论，快来抢沙发吧~" :image-size="100" />
          </div>
        </div>
      </div>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import {
  Calendar, // 修正了这里
  ChatLineRound,
  CollectionTag,
  Delete,
  StarFilled,
  View
} from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import axios from '../axios';
import { useAuthStore } from '../store/auth';

interface Article {
  id: number;
  title: string;
  content: string;
  cover?: string;
  likesCount: number;
  viewsCount: number;
  categoryId: number;
  createdAt: string;
}

interface User {
  id: number;
  userName: string;
  avatar?: string;
  role?: string;
}

interface Comment {
  id: number;
  userId: number;
  user: User;
  userName: string;
  content: string;
  parentId?: number;
  createdAt: string;
}

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const article = ref<Article | null>(null);
const comments = ref<Comment[]>([]);
const newComment = ref('');
const replyContents = ref<{ [key: number]: string }>({});
const replyBoxId = ref<number | null>(null);
const hasLiked = ref(false);
const isFavorited = ref(false);

const fetchArticle = async () => {
  try {
    const res = await axios.get<Article>(`/articles/${route.params.id}`);
    article.value = res.data;
  } catch (error) {
    console.error(error);
  }
};

const fetchComments = async () => {
  try {
    const res = await axios.get<Comment[]>(`/articles/${route.params.id}/comments`);
    comments.value = res.data.map(c => ({
      ...c,
      userName: c.userName || `用户${c.userId}`
    }));
  } catch (error) {
    console.error(error);
  }
};

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return `${date.getFullYear()}-${(date.getMonth()+1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')} ${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
};

const getParentUser = (pid: number) => {
  const parent = comments.value.find(c => c.id === pid);
  return parent?.userName || 'Unknown';
};

const toggleReply = (id: number) => {
  replyBoxId.value = replyBoxId.value === id ? null : id;
};

const canDelete = (comment: Comment) => {
  if (!authStore.user) return false;
  return authStore.user.role === 'admin' || authStore.user.ID === comment.userId;
};

const deleteComment = async (id: number) => {
  if (!id) return;
  try {
    await axios.delete(`/comments/${id}`); 
    ElMessage.success('删除成功');
    fetchComments();
  } catch (e) {
    ElMessage.error('删除失败');
  }
};

const submitComment = async (parentId: number | null) => {
  const content = parentId ? replyContents.value[parentId!] : newComment.value;
  if (!content?.trim()) return ElMessage.warning('请输入内容');

  try {
    await axios.post(`/articles/${route.params.id}/comments`, { content, parentId });
    if (parentId) replyContents.value[parentId] = '';
    else newComment.value = '';
    replyBoxId.value = null;
    fetchComments();
    ElMessage.success('评论发布成功');
  } catch {
    ElMessage.error('评论失败');
  }
};

const likeArticle = async () => {
  if (!authStore.isAuthenticated) return ElMessage.warning('请登录后操作');
  try {
    const res = await axios.post(`/articles/${route.params.id}/like`); 
    const action = res.data.action;
    const newLikesCount = res.data.likes_count;

    if (action === 'like') {
      hasLiked.value = true;
      ElMessage.success('点赞成功');
    } else if (action === 'unlike') {
      hasLiked.value = false;
      ElMessage.info('已取消点赞');
    }

    if (article.value) {
      article.value.likesCount = newLikesCount;
    }
  } catch (err: any) {
    ElMessage.error('操作失败');
  }
};

const toggleFavorite = async () => {
  if (!authStore.isAuthenticated) return ElMessage.warning('请登录后操作');
  try {
    await axios.post(`/articles/${route.params.id}/favorite`);
    isFavorited.value = !isFavorited.value;
    ElMessage.success(isFavorited.value ? '收藏成功' : '已取消收藏');
  } catch {
    ElMessage.error('操作失败');
  }
};

onMounted(() => {
  fetchArticle();
  fetchComments();
});
</script>

<style scoped>
/* 样式部分保持不变，因为没有错误 */
.news-detail-container { 
  background-color: #f4f5f7; 
  min-height: calc(100vh - 60px); 
  padding: 30px 0 60px; 
}

.content-wrapper { 
  max-width: 860px; 
  margin: 0 auto; 
  width: 100%; 
}

.loading-wrapper {
  background: #fff;
  padding: 40px;
  border-radius: 8px;
}

/* 文章卡片 */
.article-card { 
  border-radius: 12px; 
  border: none;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03) !important;
  padding: 20px 10px;
  margin-bottom: 30px;
  background: #fff;
}

.article-header {
  text-align: center;
  margin-bottom: 30px;
}

.category-tag {
  margin-bottom: 15px;
}

.article-title { 
  font-size: 2.2rem; 
  font-weight: 700;
  color: #1a1a1a;
  line-height: 1.4;
  margin: 0 0 20px; 
}

.meta-row { 
  display: flex; 
  justify-content: center;
  gap: 25px; 
  color: #909399; 
  font-size: 0.9rem; 
}

.meta-item {
  display: flex;
  align-items: center;
}

.meta-item .el-icon {
  margin-right: 6px;
  font-size: 1rem;
}

.article-cover-wrapper {
  margin: 20px 0 30px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

.cover-image {
  width: 100%;
  max-height: 400px;
  object-fit: cover;
}

.content-divider {
  margin: 30px 0;
}

/* 文章内容排版 */
.article-content { 
  font-size: 1.1rem; 
  line-height: 1.8; 
  color: #333; 
  text-align: justify;
  padding: 0 20px;
}

:deep(.article-content img) {
  max-width: 100%;
  border-radius: 8px;
  margin: 10px 0;
}

:deep(.article-content h2) {
  margin-top: 30px;
  margin-bottom: 15px;
  font-size: 1.5rem;
  border-left: 4px solid #409EFF;
  padding-left: 12px;
}

:deep(.article-content p) {
  margin-bottom: 20px;
}

.article-footer {
  margin-top: 50px;
  padding-top: 30px;
  border-top: 1px dashed #eee;
  display: flex;
  justify-content: center;
}

.action-buttons {
  display: flex;
  gap: 20px;
}

.action-btn {
  padding: 12px 30px;
  font-weight: 500;
}

.mr-1 { margin-right: 6px; }

/* 评论区 */
.comments-section {
  max-width: 860px;
  margin: 0 auto;
}

.section-header {
  margin-bottom: 20px;
  border-left: 4px solid #409EFF;
  padding-left: 12px;
}

.section-header h3 {
  font-size: 1.4rem;
  margin: 0;
}

.comment-count {
  font-size: 1rem;
  color: #909399;
  font-weight: normal;
  margin-left: 5px;
}

/* 评论输入框 */
.comment-input-card {
  border-radius: 12px;
  border: none;
  margin-bottom: 30px;
  background: #fff;
}

.comment-input-area {
  display: flex;
  gap: 20px;
}

.user-avatar-mini {
  flex-shrink: 0;
}

.input-wrapper {
  flex-grow: 1;
}

.input-footer {
  margin-top: 12px;
  text-align: right;
}

/* 评论列表 */
.comment-list {
  background: #fff;
  border-radius: 12px;
  padding: 10px 20px;
}

.comment-item {
  display: flex;
  gap: 20px;
  padding: 25px 0;
  border-bottom: 1px solid #f0f2f5;
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-avatar-col {
  flex-shrink: 0;
}

.comment-content-col {
  flex-grow: 1;
}

.comment-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.username {
  font-size: 1rem;
  font-weight: 600;
  color: #303133;
}

.timestamp {
  font-size: 0.85rem;
  color: #C0C4CC;
}

.comment-text {
  font-size: 1rem;
  color: #606266;
  line-height: 1.6;
  margin-bottom: 12px;
}

.reply-target {
  color: #409EFF;
  font-weight: 500;
  margin-right: 4px;
}

.comment-actions-row {
  display: flex;
  gap: 15px;
}

/* 回复框动画 */
.reply-box {
  margin-top: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.reply-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 10px;
}

.fade-in {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.list-enter-active,
.list-leave-active {
  transition: all 0.5s ease;
}
.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(30px);
}

@media (max-width: 768px) {
  .article-title { font-size: 1.6rem; }
  .meta-row { flex-wrap: wrap; gap: 10px; }
  .article-content { padding: 0; }
  .comment-item { gap: 12px; }
  .comment-avatar { width: 36px; height: 36px; }
}
</style>






