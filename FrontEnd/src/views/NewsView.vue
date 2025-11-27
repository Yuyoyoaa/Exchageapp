<template>
  <el-container>
    <el-main>
      <!-- 分类筛选和热门文章 -->
      <div class="filter-section">
        <el-select v-model="categoryFilter" placeholder="选择分类" @change="handleCategoryChange">
          <el-option label="全部" value="" />
          <el-option v-for="category in categories" :key="category.id" 
            :label="category.name" :value="category.id" />
        </el-select>
        <el-button type="primary" @click="fetchHotArticles">热门文章</el-button>
      </div>

      <!-- 文章列表 -->
      <div v-if="articles.length">
        <el-card v-for="article in articles" :key="article.id" class="article-card">
          <div class="article-header">
            <h2>{{ article.title }}</h2>
            <span class="article-meta">
              浏览: {{ article.viewsCount }} | 点赞: {{ article.likesCount }}
            </span>
          </div>
          <p class="article-preview">{{ article.preview }}</p>
          <div class="article-footer">
            <el-button text @click="viewDetail(article.id)">阅读更多</el-button>
            <span class="article-time">{{ formatDate(article.createdAt) }}</span>
          </div>
        </el-card>
      </div>
      <div v-else class="no-data">暂无文章</div>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[5, 10, 20]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
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
const categoryFilter = ref('');
const currentPage = ref(1);
const pageSize = ref(10);
const total = ref(0);
const router = useRouter();

const fetchArticles = async () => {
  try {
    const params: any = { page: currentPage.value, limit: pageSize.value };
    if (categoryFilter.value) params.category = categoryFilter.value;
    const res = await axios.get<{ data: Article[]; total: number }>('/articles', { params });
    // 注意这里和后端返回格式一致
    articles.value = res.data.data;
    total.value = res.data.total;
  } catch (error) {
    console.error(error);
    ElMessage.error('加载文章失败');
  }
};

const fetchCategories = async () => {
  try {
    const res = await axios.get<Category[]>('/categories');
    categories.value = res.data;
  } catch (error) {
    console.error(error);
  }
};

const fetchHotArticles = async () => {
  try {
    const res = await axios.get<Article[]>('/articles/hot');
    articles.value = res.data;
    categoryFilter.value = '';
    currentPage.value = 1;
  } catch (error) {
    console.error(error);
    ElMessage.error('加载热门文章失败');
  }
};

const handleCategoryChange = () => { currentPage.value = 1; fetchArticles(); };
const handleSizeChange = (size: number) => { pageSize.value = size; currentPage.value = 1; fetchArticles(); };
const handleCurrentChange = (page: number) => { currentPage.value = page; fetchArticles(); };

const viewDetail = (id: number) => {
  router.push({ name: 'NewsDetail', params: { id: id.toString() } });
};

const formatDate = (dateStr?: string) => dateStr ? new Date(dateStr).toLocaleDateString('zh-CN') : '';

onMounted(() => {
  fetchCategories();
  fetchArticles();
});
</script>

<style scoped>
.filter-section { display: flex; gap: 10px; margin-bottom: 20px; align-items: center; }
.article-card { margin: 20px 0; padding: 20px; }
.article-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.article-meta { color: #666; font-size: 0.9em; }
.article-preview { color: #555; line-height: 1.6; margin-bottom: 15px; }
.article-footer { display: flex; justify-content: space-between; align-items: center; }
.article-time { color: #999; font-size: 0.8em; }
.pagination-container { display: flex; justify-content: center; margin-top: 20px; }
.no-data { text-align: center; color: #999; font-size: 1.2em; padding: 40px 0; }
</style>

