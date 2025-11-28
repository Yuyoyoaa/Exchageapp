<template>
  <el-container class="news-detail-container">
    <el-main>
      <div class="content-wrapper" v-if="article">
        <!-- 文章卡片 -->
        <el-card class="article-card">
          <div class="article-header">
            <h1>{{ article.title }}</h1>
            <div class="meta-row">
              <span>{{ formatDate(article.createdAt) }}</span>
              <span>浏览: {{ article.viewsCount }}</span>
              <span>点赞: {{ article.likesCount }}</span>
            </div>
          </div>

          <div v-if="article.cover" class="article-cover">
            <el-image :src="article.cover" fit="cover" class="cover-image" />
          </div>

          <el-divider />

          <div class="article-content" v-html="article.content"></div>

          <!-- 点赞收藏 -->
          <div class="action-bar">
            <el-button
              :type="hasLiked ? 'primary' : 'default'"
              circle
              size="large"
              @click="likeArticle"
            >
              点赞
            </el-button>
            <span class="action-label">{{ hasLiked ? '已点赞' : '点赞' }}</span>

            <el-button
              :type="isFavorited ? 'warning' : 'default'"
              circle
              size="large"
              @click="toggleFavorite"
              style="margin-left: 20px;"
            >
              收藏
            </el-button>
            <span class="action-label">{{ isFavorited ? '已收藏' : '收藏' }}</span>
          </div>
        </el-card>

        <!-- 评论区 -->
        <el-card class="comments-card">
          <template #header>
            <div>评论 ({{ comments.length }})</div>
          </template>

          <!-- 新评论输入 -->
          <div v-if="authStore.isAuthenticated" class="comment-input-area">
            <el-input
              v-model="newComment"
              type="textarea"
              :rows="3"
              placeholder="写下你的看法..."
            />
            <div class="comment-submit">
              <el-button type="primary" @click="submitComment(null)">发表评论</el-button>
            </div>
          </div>

          <el-alert
            v-else
            title="请登录后参与评论"
            type="info"
            center
            show-icon
            :closable="false"
            style="margin-bottom:20px;"
          />

          <!-- 评论列表 -->
          <div class="comment-list">
            <div v-for="comment in comments" :key="comment.id" class="comment-item">
              <div class="comment-avatar">
                <el-avatar :size="40">{{ comment.user?.avatar?.charAt(0) || comment.userName?.charAt(0) || 'U' }}</el-avatar>
              </div>

              <div class="comment-body">
                <div class="comment-info">
                  <span class="username">{{ comment.userName || '匿名用户' }}</span>
                  <span class="timestamp">{{ formatDate(comment.createdAt) }}</span>
                </div>

                <div class="comment-text">
                  <span v-if="comment.parentId" class="reply-tag">
                    回复 @{{ getParentUser(comment.parentId) }}:
                  </span>
                  {{ comment.content }}
                </div>

                <div class="comment-actions">
                  <el-button
                    link
                    size="small"
                    type="primary"
                    v-if="authStore.isAuthenticated"
                    @click="toggleReply(comment.id)"
                  >
                    回复
                  </el-button>

                  <el-button
                    link
                    size="small"
                    type="danger"
                    v-if="canDelete(comment)"
                    @click="deleteComment(comment.id)"
                  >
                    删除
                  </el-button>
                </div>

                <div v-if="replyBoxId === comment.id" class="reply-box">
                  <el-input v-model="replyContents[comment.id]" size="small" placeholder="回复内容..." />
                  <el-button type="primary" size="small" @click="submitComment(comment.id)" style="margin-top:5px;">
                    发送
                  </el-button>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <div v-else>加载中...</div>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
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
}

interface Comment {
  id: number;
  userId: number;
  user: User;  // 添加 user 对象
  userName: string;
  content: string;
  parentId?: number;
  createdAt: string;
}

const route = useRoute();
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

const formatDate = (dateStr?: string) => dateStr ? new Date(dateStr).toLocaleString('zh-CN') : '';

const getParentUser = (pid: number) => {
  const parent = comments.value.find(c => c.id === pid);
  return parent?.userName || 'Unknown';
};

const toggleReply = (id: number) => {
  replyBoxId.value = replyBoxId.value === id ? null : id;
};

const canDelete = (comment: Comment) => {
  if (!authStore.user) return false;
  // 检查是否为管理员 (admin) 或评论的创建者 (comment.userId === authStore.user.ID)
  return authStore.user.role === 'admin' || authStore.user.ID === comment.userId;
};

const deleteComment = async (id: number) => {
  // 0. 添加防御性检查，防止 ID 缺失导致请求发送到 /comments/undefined
  if (!id) {
    console.error('Error: Comment ID is missing or invalid.');
    return ElMessage.error('评论 ID 缺失，无法执行删除操作。');
  }

  // 1. 添加确认对话框
  try {
    // 使用 Element Plus 的 ElMessageBox.confirm 进行用户确认
    await ElMessageBox.confirm('确定要删除这条评论吗？', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
    });
  } catch {
    // 用户点击取消或关闭对话框
    ElMessage.info('已取消删除操作');
    return;
  }
  
  // 2. 执行删除操作
  try {
    // 使用传入的有效 id 执行 DELETE 请求
    await axios.delete(`/comments/${id}`); 
    ElMessage.success('删除成功');
    
    // 3. 重新获取评论列表以刷新 UI
    fetchComments();
  } catch (e) {
    console.error('删除评论失败:', e);
    // 给出更详细的错误提示，可能是权限不足或网络问题
    ElMessage.error('删除失败，请检查您的权限或网络连接');
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
    ElMessage.success('评论成功');
  } catch {
    ElMessage.error('评论失败');
  }
};

const likeArticle = async () => {
  if (!authStore.isAuthenticated) return ElMessage.warning('请登录');
  try {
    // 捕获 API 响应
    const res = await axios.post(`/articles/${route.params.id}/like`); 
    const action = res.data.action;
    const newLikesCount = res.data.likes_count;

    // 根据后端返回的 action 更新前端状态
    if (action === 'like') {
      hasLiked.value = true;
      ElMessage.success('点赞成功');
    } else if (action === 'unlike') {
      hasLiked.value = false;
      ElMessage.success('已取消点赞');
    }

    // 使用后端返回的最新点赞数更新 UI，避免前端手动增减的错误
    if (article.value) {
      article.value.likesCount = newLikesCount;
    }
  } catch (err: any) {
    console.error(err);
    ElMessage.error(err.response?.data?.error || '操作失败'); // 统一为“操作失败”
  }
};

const toggleFavorite = async () => {
  if (!authStore.isAuthenticated) return ElMessage.warning('请登录');
  try {
    await axios.post(`/articles/${route.params.id}/favorite`);
    isFavorited.value = !isFavorited.value;
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
.news-detail-container { background: #f5f7fa; min-height: calc(100vh - 60px); padding: 20px; }
.content-wrapper { max-width: 800px; margin: 0 auto; }
.article-card { margin-bottom: 20px; border-radius: 8px; padding: 20px; }
.article-header h1 { font-size: 2rem; margin-bottom: 15px; }
.meta-row { display: flex; gap: 15px; color: #999; font-size: 0.9rem; margin-bottom: 20px; }
.article-cover { margin-bottom: 20px; border-radius: 4px; overflow: hidden; }
.article-content { font-size: 1.1rem; line-height: 1.8; color: #444; min-height: 200px; }
.action-bar { margin-top: 20px; display: flex; align-items: center; }
.action-label { margin-left: 8px; color: #666; }

.comments-card { border-radius: 8px; margin-top: 30px; padding: 15px; }
.comment-item { display: flex; gap: 15px; padding: 10px 0; border-bottom: 1px solid #eee; }
.comment-avatar { flex-shrink: 0; }
.comment-body { flex: 1; }
.comment-info { display: flex; justify-content: space-between; margin-bottom: 5px; }
.username { font-weight: bold; color: #333; }
.timestamp { font-size: 0.8rem; color: #999; }
.comment-text { color: #555; line-height: 1.5; }
.reply-tag { color: #409EFF; margin-right: 5px; }
.comment-actions { margin-top: 5px; }
.reply-box { margin-top: 10px; background: #f9f9f9; padding: 10px; border-radius: 4px; }
.comment-submit { margin-top: 10px; text-align: right; }
</style>






