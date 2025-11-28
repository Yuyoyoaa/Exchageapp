<template>
  <el-container class="news-container">
    <el-main>
      <el-card shadow="always" class="main-card">
        <!-- 筛选和查询部分 -->
        <div class="filter-section">
          <div class="filter-group">
            <!-- 分类筛选 -->
            <el-select 
              v-model="categoryFilter" 
              placeholder="全部分类"
              class="category-select"
              clearable
              size="large"
            >
              <el-option label="全部" value="" />
              <el-option
                v-for="category in categories"
                :key="category.id"
                :label="category.name"
                :value="category.id"
              />
            </el-select>

            <el-button type="primary" size="large" @click="searchArticles" class="filter-btn">
              <el-icon class="mr-1"><Search /></el-icon> 查询
            </el-button>
            <el-button type="warning" size="large" plain @click="fetchHotArticles" class="filter-btn-hot">
              <el-icon class="mr-1"><StarFilled /></el-icon> 热门文章
            </el-button>
          </div>
        </div>

        <!-- 文章列表 -->
        <div v-if="articles.length" class="articles-list">
          <el-row :gutter="24">
            <el-col
              v-for="article in articles"
              :key="article.id"
              :xs="24" :sm="12" :md="8"
              class="article-col"
            >
              <el-card class="article-card" shadow="hover" :body-style="{ padding: '0' }" @click="viewDetail(article.id)">
                <!-- 如果有封面图可以在这里展示，目前用纯色块或渐变代替演示，或者直接去掉 -->
                <div class="article-cover-placeholder" :style="{ backgroundColor: getRandomColor(article.id) }">
                  <el-tag class="category-tag" effect="dark" size="small">
                    {{ getCategoryName(article.categoryId) }}
                  </el-tag>
                </div>
                
                <div class="article-content">
                  <h3 class="article-title" :title="article.title">
                    {{ article.title }}
                  </h3>
                  
                  <div class="article-meta-row">
                    <span class="article-stat">
                      <el-icon><View /></el-icon> {{ article.viewsCount }}
                    </span>
                    <span class="article-stat">
                      <el-icon><StarFilled /></el-icon> {{ article.likesCount }}
                    </span>
                    <span class="article-time">{{ formatDate(article.createdAt) }}</span>
                  </div>

                  <p class="article-preview">{{ article.preview }}</p>
                  
                  <div class="article-footer">
                    <el-button text type="primary" class="read-btn">
                      阅读全文 <el-icon class="el-icon--right"><ArrowRight /></el-icon>
                    </el-button>
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>
        
        <div v-else class="no-data">
          <el-empty description="暂无文章或未找到匹配结果" :image-size="120" />
        </div>

        <!-- 分页器（热门模式下隐藏） -->
        <div class="pagination-container" v-if="articles.length > 0 && !isHotMode">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[6, 9, 12, 15]"
            :total="total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            background
          />
        </div>
      </el-card>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ArrowRight, Search, StarFilled, View } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import axios from '../axios';

interface Article {
  id: number;
  title: string;
  preview: string;
  content: string;
  cover?: string;
  likesCount: number;
  viewsCount: number;
  categoryId: number;
  createdAt: string;
}

interface Category {
  id: number;
  name: string;
}

const articles = ref<Article[]>([]);
const categories = ref<Category[]>([]);
const categoryFilter = ref<number | ''>(''); // 选中分类
const currentPage = ref(1);
const pageSize = ref(9); // 调整默认每页数量为3的倍数，适应3列布局
const total = ref(0);
const isHotMode = ref(false);
const router = useRouter();

const fetchArticles = async () => {
  try {
    const params: any = { page: currentPage.value, limit: pageSize.value };
    if (typeof categoryFilter.value === 'number') params.category = categoryFilter.value;
    const res = await axios.get<{ data: Article[]; total: number }>('/articles', { params });
    isHotMode.value = false;
    articles.value = res.data.data;
    total.value = res.data.total;
  } catch (error) {
    ElMessage.error('加载文章列表失败');
  }
};

const searchArticles = () => {
  currentPage.value = 1;
  fetchArticles();
};

const fetchHotArticles = async () => {
  try {
    isHotMode.value = true;
    const res = await axios.get<Article[]>('/articles/hot');
    articles.value = res.data;
    categoryFilter.value = '';
    total.value = articles.value.length;
    ElMessage.success(`已加载 ${articles.value.length} 篇热门文章`);
  } catch {
    ElMessage.error('加载热门文章失败');
  }
};

const fetchCategories = async () => {
  try {
    const res = await axios.get<Category[]>('/categories');
    categories.value = res.data;
  } catch {}
};

const handleSizeChange = (size: number) => {
  pageSize.value = size;
  currentPage.value = 1;
  fetchArticles();
};

const handleCurrentChange = (page: number) => {
  currentPage.value = page;
  fetchArticles();
};

const getCategoryName = (id: number) => {
  const category = categories.value.find(c => c.id === id);
  return category ? category.name : '资讯';
};

const viewDetail = (id: number) => {
  router.push({ name: 'NewsDetail', params: { id: id.toString() } });
};

const formatDate = (dateStr?: string) => dateStr
  ? new Date(dateStr).toLocaleDateString('zh-CN', {
    month: '2-digit', day: '2-digit' // 简化日期显示
  }) : '';

// 生成随机柔和颜色作为封面占位
const getRandomColor = (id: number) => {
  const colors = ['#E8F3FF', '#FFF3E0', '#E8F5E9', '#F3E5F5', '#FFF8E1'];
  return colors[id % colors.length];
};

onMounted(() => {
  fetchCategories();
  fetchArticles();
});
</script>

<style scoped>
.news-container {
  background-color: #f5f7fa;
  min-height: calc(100vh - 60px);
}

.main-card {
  max-width: 1200px;
  margin: 20px auto;
  border-radius: 16px;
  border: none;
  background: transparent;
  box-shadow: none !important; /* 去掉外层卡片的阴影，让内部卡片浮起来 */
}

:deep(.el-card__body) {
  padding: 0;
}

.filter-section {
  background: #fff;
  padding: 20px;
  border-radius: 12px;
  margin-bottom: 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.category-select {
  width: 180px;
}

.mr-1 {
  margin-right: 4px;
}

.articles-list {
  margin-top: 10px;
}

.article-col {
  margin-bottom: 24px;
}

.article-card {
  border-radius: 12px;
  border: none;
  height: 100%;
  display: flex;
  flex-direction: column;
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  cursor: pointer;
  background: #fff;
  overflow: hidden;
}

.article-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.1);
}

.article-cover-placeholder {
  height: 120px; /* 封面高度 */
  position: relative;
  /* 如果有真实图片，可以用 background-image 替换 */
}

.category-tag {
  position: absolute;
  top: 12px;
  left: 12px;
  opacity: 0.9;
}

.article-content {
  padding: 20px;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
}

.article-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #2c3e50;
  margin: 0 0 12px 0;
  line-height: 1.5;
  /* 限制标题两行 */
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  height: 3.3em; /* 固定高度防止抖动 */
}

.article-meta-row {
  display: flex;
  align-items: center;
  font-size: 0.85rem;
  color: #909399;
  margin-bottom: 12px;
}

.article-stat {
  display: flex;
  align-items: center;
  margin-right: 12px;
}

.article-stat .el-icon {
  margin-right: 4px;
}

.article-time {
  margin-left: auto; /* 将时间推到最右边 */
}

.article-preview {
  font-size: 0.9rem;
  color: #606266;
  line-height: 1.6;
  margin-bottom: 16px;
  /* 限制预览文三行 */
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  overflow: hidden;
  flex-grow: 1;
}

.article-footer {
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid #f2f6fc;
}

.read-btn {
  padding: 0;
  font-weight: 500;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 30px;
  padding-bottom: 20px;
}

.no-data {
  text-align: center;
  padding: 60px 0;
}

@media (max-width: 768px) {
  .filter-group {
    flex-wrap: wrap;
  }
  .category-select {
    width: 100%;
  }
  .filter-btn, .filter-btn-hot {
    flex: 1;
  }
  .article-cover-placeholder {
    height: 100px;
  }
}
</style>

