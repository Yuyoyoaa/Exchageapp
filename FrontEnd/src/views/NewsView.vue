<template>
  <el-container>
    <el-main>
      <el-card shadow="never" class="main-card">
        <!-- 筛选和查询部分 -->
        <div class="filter-section">
          <!-- 分类筛选：只绑定 model，不再自动触发查询 (@change 事件已移除) -->
          <el-select 
            v-model="categoryFilter" 
            placeholder="选择分类" 
            style="width: 150px;"
          >
            <!-- value="" 表示查询全部 -->
            <el-option label="全部" value="" />
            <el-option 
              v-for="category in categories" 
              :key="category.id" 
              :label="category.name" 
              :value="category.id" 
            />
          </el-select>
          
          <!-- 【需求 1 实现】查询文章按钮：点击后才触发分类查询 -->
          <el-button type="primary" @click="searchArticles">查询文章</el-button>

          <!-- 【需求 2 实现】热门文章按钮：切换到热门文章列表 -->
          <el-button type="warning" @click="fetchHotArticles">
            <el-icon><StarFilled /></el-icon> 热门文章
          </el-button>
        </div>

        <!-- 文章列表 -->
        <div v-if="articles.length">
          <el-card 
            v-for="article in articles" 
            :key="article.id" 
            class="article-card" 
            shadow="hover"
          >
            <div class="article-header">
              <h2 class="article-title">{{ article.title }}</h2>
              <span class="article-meta">
                分类: {{ getCategoryName(article.categoryId) }} | 浏览: {{ article.viewsCount }} | 点赞: {{ article.likesCount }}
              </span>
            </div>
            <p class="article-preview">{{ article.preview }}</p>
            <div class="article-footer">
              <el-button type="primary" link @click="viewDetail(article.id)">阅读更多</el-button>
              <span class="article-time">{{ formatDate(article.createdAt) }}</span>
            </div>
          </el-card>
        </div>
        <div v-else class="no-data">
          <el-empty description="暂无文章或未找到匹配结果" />
        </div>

        <!-- 分页 (当不是热门文章模式时才显示) -->
        <!-- isHotMode 为 false 时才显示分页器 -->
        <div class="pagination-container" v-if="articles.length > 0 && !isHotMode">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[5, 10, 20]"
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
import { StarFilled } from '@element-plus/icons-vue';
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
const categoryFilter = ref<number | ''>(''); // 存储选中的分类ID
const currentPage = ref(1);
const pageSize = ref(10);
const total = ref(0);
const isHotMode = ref(false); // 跟踪是否处于热门文章模式
const router = useRouter();

// 获取所有文章 (分页和筛选逻辑)
const fetchArticles = async () => {
  try {
    const params: any = { page: currentPage.value, limit: pageSize.value };
    
    // 只有当 categoryFilter 是有效的 ID 时才添加 category 参数
    if (typeof categoryFilter.value === 'number') {
      params.category = categoryFilter.value;
    }

    const res = await axios.get<{ data: Article[]; total: number }>('/articles', { params });
    
    // 退出热门模式，并更新分页数据
    isHotMode.value = false;
    articles.value = res.data.data;
    total.value = res.data.total;

  } catch (error) {
    console.error('加载文章失败:', error);
    ElMessage.error('加载文章列表失败');
  }
};

// 【需求 1 实现】查询文章：从第一页开始，根据当前的 categoryFilter 查询
const searchArticles = () => {
  currentPage.value = 1;
  fetchArticles();
};

// 【需求 2 实现】获取热门文章 (非分页逻辑)
const fetchHotArticles = async () => {
  try {
    // 1. 切换到热门模式
    isHotMode.value = true;
    
    // 2. 调用热门文章API
    const res = await axios.get<Article[]>('/articles/hot');
    articles.value = res.data;
    
    // 3. 重置筛选条件，并将总数设置为文章列表长度（隐藏分页器）
    categoryFilter.value = ''; 
    total.value = articles.value.length; 
    
    ElMessage.success(`已加载 ${articles.value.length} 篇热门文章`);
  } catch (error) {
    console.error('加载热门文章失败:', error);
    ElMessage.error('加载热门文章失败');
  }
};

// 获取分类列表
const fetchCategories = async () => {
  try {
    const res = await axios.get<Category[]>('/categories');
    categories.value = res.data;
  } catch (error) {
    console.error('加载分类失败:', error);
  }
};

// 分页大小改变：重置到第一页并查询
const handleSizeChange = (size: number) => { 
  pageSize.value = size; 
  currentPage.value = 1; 
  fetchArticles(); 
};

// 分页页码改变：查询当前页
const handleCurrentChange = (page: number) => { 
  currentPage.value = page; 
  fetchArticles(); 
};

// 根据 ID 获取分类名称
const getCategoryName = (id: number) => {
  const category = categories.value.find(c => c.id === id);
  return category ? category.name : '未知分类';
};

// 跳转到文章详情页
const viewDetail = (id: number) => {
  router.push({ name: 'NewsDetail', params: { id: id.toString() } });
};

// 格式化日期
const formatDate = (dateStr?: string) => dateStr ? new Date(dateStr).toLocaleDateString('zh-CN', {
    year: 'numeric', month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit'
  }) : '';

onMounted(() => {
  fetchCategories();
  fetchArticles(); // 默认加载全部文章
});
</script>

<style scoped>
.main-card {
  max-width: 1200px;
  margin: 20px auto;
  border-radius: 12px;
}
.filter-section { 
  display: flex; 
  gap: 10px; 
  margin-bottom: 20px; 
  align-items: center; 
  padding: 10px 0;
  border-bottom: 1px solid #ebeef5;
}
.article-card { 
  margin-bottom: 20px; 
  border-radius: 8px;
  transition: all 0.3s ease;
}
.article-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}
.article-header { 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  margin-bottom: 10px; 
}
.article-title {
  font-size: 1.5em;
  color: #303133;
  margin: 0;
}
.article-meta { 
  color: #909399; 
  font-size: 0.85em; 
  white-space: nowrap;
}
.article-preview { 
  color: #606266; 
  line-height: 1.7; 
  margin-bottom: 15px; 
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3; /* 限制三行预览 */
  overflow: hidden;
  min-height: 60px; /* 保证排版一致 */
}
.article-footer { 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
}
.article-time { 
  color: #c0c4cc; 
  font-size: 0.8em; 
}
.pagination-container { 
  display: flex; 
  justify-content: center; 
  margin-top: 30px; 
  padding-bottom: 10px;
}
.no-data { 
  text-align: center; 
  padding: 40px 0; 
}
</style>

