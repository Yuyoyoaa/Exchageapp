<template>  
  <el-container class="home-container">  
    <el-main>
      <div class="welcome-section">
        <div class="welcome-content">
          <h1 class="title">🚀 欢迎使用YOYO兑换系统</h1>
          <p class="description">一站式货币兑换与金融资讯平台</p>
          
          <div class="feature-cards">
            <el-row :gutter="20">
              <el-col :xs="24" :sm="8" class="feature-col">
                <el-card class="feature-card" shadow="hover">
                  <div class="feature-icon">💱</div>
                  <h3>实时汇率</h3>
                  <p>获取最新的货币兑换汇率，支持多种货币对</p>
                  <el-button type="primary" @click="router.push('/exchange')">开始兑换</el-button>
                </el-card>
              </el-col>
              
              <el-col :xs="24" :sm="8" class="feature-col">
                <el-card class="feature-card" shadow="hover">
                  <div class="feature-icon">📰</div>
                  <h3>金融资讯</h3>
                  <p>浏览最新的金融新闻和市场分析文章</p>
                  <el-button type="primary" @click="router.push('/news')">阅读资讯</el-button>
                </el-card>
              </el-col>
              
              <el-col :xs="24" :sm="8" class="feature-col">
                <el-card class="feature-card" shadow="hover">
                  <div class="feature-icon">👤</div>
                  <h3>个人中心</h3>
                  <p>管理您的个人信息和交易记录</p>
                  <el-button 
                    type="primary" 
                    @click="authStore.isAuthenticated ? router.push('/profile') : router.push('/login')"
                  >
                    {{ authStore.isAuthenticated ? '个人中心' : '立即登录' }}
                  </el-button>
                </el-card>
              </el-col>
            </el-row>
          </div>

          <!-- 热门文章预览 -->
          <div class="hot-articles" v-if="hotArticles.length">
            <h2 class="section-title">🔥 热门文章</h2>
            <el-row :gutter="20">
              <el-col 
                v-for="article in hotArticles" 
                :key="article.ID" 
                :xs="24" 
                :sm="8" 
                class="article-col"
              >
                <el-card class="article-preview-card" shadow="hover" @click="viewArticle(article.ID)">
                  <div class="article-preview">
                    <h4>{{ article.Title }}</h4>
                    <p class="article-excerpt">{{ article.Preview }}</p>
                    <div class="article-meta">
                      <span>👍 {{ article.LikesCount }}</span>
                      <span>👁 {{ article.ViewsCount }}</span>
                    </div>
                  </div>
                </el-card>
              </el-col>
            </el-row>
          </div>

          <!-- 快速操作 -->
          <div class="quick-actions">
            <h2 class="section-title">⚡ 快速操作</h2>
            <div class="action-buttons">
              <el-button 
                type="primary" 
                size="large" 
                @click="router.push('/exchange')"
                icon="ShoppingCart"
              >
                货币兑换
              </el-button>
              <el-button 
                type="success" 
                size="large" 
                @click="router.push('/news')"
                icon="Reading"
              >
                浏览资讯
              </el-button>
              <el-button 
                v-if="!authStore.isAuthenticated"
                type="warning" 
                size="large" 
                @click="router.push('/register')"
                icon="User"
              >
                立即注册
              </el-button>
              <el-button 
                v-if="authStore.user?.role === 'admin'"
                type="danger" 
                size="large" 
                @click="router.push('/admin/users')"
                icon="Setting"
              >
                管理后台
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </el-main>
  </el-container>  
</template>  
  
<script setup lang="ts">
import { ElMessage } from 'element-plus';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import axios from '../axios';
import { useAuthStore } from '../store/auth';
import type { Article } from "../types/Article";

const router = useRouter();
const authStore = useAuthStore();
const hotArticles = ref<Article[]>([]);

const fetchHotArticles = async () => {
  try {
    const response = await axios.get<Article[]>('/articles/hot');
    hotArticles.value = response.data.slice(0, 3); // 只显示前3篇
  } catch (error) {
    console.error('Failed to load hot articles:', error);
  }
};

const viewArticle = (id: number) => {
  if (!authStore.isAuthenticated) {
    ElMessage.warning('请先登录后再查看文章');
    return;
  }
  router.push({ name: 'NewsDetail', params: { id: id.toString() } });
};

onMounted(() => {
  fetchHotArticles();
});
</script>  
  
<style scoped>
.home-container {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: calc(100vh - 60px);
}

.welcome-section {
  padding: 60px 20px;
}

.welcome-content {
  max-width: 1200px;
  margin: 0 auto;
  text-align: center;
}

.title {
  color: white;
  font-size: 3rem;
  font-weight: bold;
  margin-bottom: 1rem;
  text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
}

.description {
  color: rgba(255, 255, 255, 0.9);
  font-size: 1.5rem;
  margin-bottom: 3rem;
}

.feature-cards {
  margin: 4rem 0;
}

.feature-col {
  margin-bottom: 2rem;
}

.feature-card {
  height: 100%;
  text-align: center;
  transition: transform 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
}

.feature-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.feature-card h3 {
  margin: 1rem 0;
  color: #333;
}

.feature-card p {
  color: #666;
  margin-bottom: 1.5rem;
  line-height: 1.6;
}

.hot-articles {
  margin: 4rem 0;
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.1);
}

.section-title {
  text-align: center;
  color: #333;
  margin-bottom: 2rem;
  font-size: 2rem;
}

.article-col {
  margin-bottom: 1.5rem;
}

.article-preview-card {
  height: 100%;
  cursor: pointer;
  transition: all 0.3s ease;
}

.article-preview-card:hover {
  box-shadow: 0 6px 25px rgba(0,0,0,0.15);
  transform: translateY(-2px);
}

.article-preview h4 {
  margin-bottom: 1rem;
  color: #333;
  line-height: 1.4;
  height: 2.8em;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.article-excerpt {
  color: #666;
  line-height: 1.6;
  height: 4.8em;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  margin-bottom: 1rem;
}

.article-meta {
  display: flex;
  justify-content: space-between;
  color: #999;
  font-size: 0.9em;
}

.quick-actions {
  margin: 4rem 0;
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.1);
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .title {
    font-size: 2rem;
  }
  
  .description {
    font-size: 1.2rem;
  }
  
  .action-buttons {
    flex-direction: column;
    align-items: center;
  }
  
  .action-buttons .el-button {
    width: 100%;
    max-width: 300px;
  }
}
</style>