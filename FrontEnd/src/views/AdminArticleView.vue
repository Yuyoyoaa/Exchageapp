<template>
  <el-container class="admin-container">
    <el-card class="admin-card">
      <template #header>
        <div class="card-header">
          <h2>文章与分类管理</h2>
          <div class="header-actions">
            <el-button type="success" @click="categoryDialogVisible = true">管理分类</el-button>
            <el-button type="primary" @click="openCreateDialog">发布文章</el-button>
          </div>
        </div>
      </template>

      <el-table :data="articles" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="title" label="标题" show-overflow-tooltip />
        <el-table-column label="分类" width="120">
          <template #default="scope">
            <el-tag size="small">{{ getCategoryName(scope.row.categoryId) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="viewsCount" label="浏览" width="80" />
        <el-table-column prop="createdAt" label="发布时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button size="small" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination 
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑文章' : '发布文章'" width="700px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="标题" required>
          <el-input v-model="form.title" placeholder="请输入文章标题" />
        </el-form-item>
        <el-form-item label="分类" required>
          <el-select v-model="form.categoryId" placeholder="选择分类" style="width: 100%">
            <el-option v-for="c in categories" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="封面URL">
          <el-input v-model="form.cover" placeholder="图片地址" />
        </el-form-item>
        <el-form-item label="简介" required>
          <el-input v-model="form.preview" type="textarea" :rows="2" placeholder="文章预览文字" />
        </el-form-item>
        <el-form-item label="内容" required>
          <el-input v-model="form.content" type="textarea" :rows="10" placeholder="支持 HTML 内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitArticle">提交</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="categoryDialogVisible" title="分类管理" width="500px">
      <div class="category-input">
        <el-input v-model="newCategoryName" placeholder="新分类名称" />
        <el-button type="primary" @click="addCategory" :disabled="!newCategoryName">添加</el-button>
      </div>
      <el-table :data="categories" height="300px">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="名称" />
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-button type="danger" link @click="deleteCategory(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </el-container>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import axios from '../axios';

interface Article {
  id: number;
  title: string;
  content: string;
  preview: string;
  cover: string;
  categoryId: number;
  viewsCount: number;
  likesCount: number;
  createdAt: string;
  updatedAt: string;
  authorId?: number;
  status?: string;
}

interface Category {
  id: number;
  name: string;
}

const articles = ref<Article[]>([]);
const categories = ref<Category[]>([]);
const loading = ref(false);
const dialogVisible = ref(false);
const categoryDialogVisible = ref(false);
const isEdit = ref(false);
const currentId = ref<number | null>(null);
const currentPage = ref(1);
const pageSize = ref(10);
const total = ref(0);

const form = reactive({
  title: '',
  content: '',
  preview: '',
  cover: '',
  categoryId: null as number | null
});

const newCategoryName = ref('');

// 调试函数：打印数据详情
const debugData = () => {
  console.log('Articles:', articles.value);
  console.log('Categories:', categories.value);
  if (articles.value.length > 0) {
    console.log('First article:', articles.value[0]);
    console.log('First article ID:', articles.value[0].id);
    console.log('First article title:', articles.value[0].title);
    console.log('First article categoryId:', articles.value[0].categoryId);
  }
};

const fetchArticles = async () => {
  loading.value = true;
  try {
    console.log('正在获取文章列表...');
    // 注意：这里使用公共API /api/articles，不是管理员API
    const response = await axios.get(`/articles?page=${currentPage.value}&limit=${pageSize.value}`);
    console.log('文章API响应:', response);
    
    // 处理不同的响应格式
    if (response.data && Array.isArray(response.data)) {
      articles.value = response.data;
      total.value = response.data.length;
    } else if (response.data && response.data.data) {
      articles.value = response.data.data;
      total.value = response.data.total || response.data.data.length;
    } else {
      articles.value = response.data || [];
      total.value = response.data?.length || 0;
    }
    
    console.log('处理后的文章数据:', articles.value);
    debugData();
  } catch (e: any) {
    console.error('获取文章错误:', e);
    ElMessage.error(e.response?.data?.error || '加载文章失败');
  } finally {
    loading.value = false;
  }
};

const fetchCategories = async () => {
  try {
    console.log('正在获取分类列表...');
    const response = await axios.get('/categories');
    console.log('分类API响应:', response);
    categories.value = response.data;
  } catch (e: any) {
    console.error('获取分类错误:', e);
    ElMessage.error(e.response?.data?.error || '加载分类失败');
  }
};

const fetchData = async () => {
  await Promise.all([fetchArticles(), fetchCategories()]);
};

const getCategoryName = (id: number) => {
  if (!id) return '未分类';
  const cat = categories.value.find(c => c.id === id);
  return cat ? cat.name : '未分类';
};

const openCreateDialog = () => {
  isEdit.value = false;
  currentId.value = null;
  form.title = '';
  form.content = '';
  form.preview = '';
  form.cover = '';
  form.categoryId = null;
  dialogVisible.value = true;
};

const handleEdit = (row: Article) => {
  isEdit.value = true;
  currentId.value = row.id;
  form.title = row.title;
  form.content = row.content;
  form.preview = row.preview;
  form.cover = row.cover;
  form.categoryId = row.categoryId;
  dialogVisible.value = true;
};

const submitArticle = async () => {
  if (!form.title || !form.content || !form.preview || !form.categoryId) {
    ElMessage.warning('请填写完整信息');
    return;
  }

  try {
    const payload = {
      title: form.title,
      content: form.content,
      preview: form.preview,
      cover: form.cover,
      categoryId: form.categoryId,
      status: 'published'
    };

    console.log('提交文章数据:', payload);
    
    if (isEdit.value && currentId.value) {
      console.log('正在更新文章:', currentId.value);
      // 使用管理员API：/api/admin/articles/:id
      await axios.put(`/admin/articles/${currentId.value}`, payload);
      ElMessage.success('更新成功');
    } else {
      console.log('正在创建新文章');
      // 使用管理员API：/api/admin/articles
      await axios.post('/admin/articles', payload);
      ElMessage.success('发布成功');
    }
    dialogVisible.value = false;
    await fetchArticles();
  } catch (e: any) {
    console.error('提交文章错误:', e);
    if (e.response?.status === 401) {
      ElMessage.error('请先登录');
    } else if (e.response?.status === 403) {
      ElMessage.error('没有管理员权限');
    } else {
      ElMessage.error(e.response?.data?.error || '操作失败');
    }
  }
};

const handleDelete = async (row: Article) => {
  try {
    await ElMessageBox.confirm('确定要删除这篇文章吗?', '警告', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    });
    
    console.log('正在删除文章:', row.id);
    // 使用管理员API：/api/admin/articles/:id
    await axios.delete(`/admin/articles/${row.id}`);
    ElMessage.success('删除成功');
    await fetchArticles();
  } catch (error: any) {
    if (error === 'cancel') {
      // 用户取消删除
      return;
    }
   console.error('错误响应:', error.response);
    
    if (error.response?.status === 401) {
      ElMessage.error('请先登录');
    } else if (error.response?.status === 403) {
      ElMessage.error('没有管理员权限');
    } else if (error.response?.status === 500) {
      // 服务器错误，显示具体错误信息
      const errorMsg = error.response?.data?.error || '服务器内部错误';
      ElMessage.error(`服务器错误: ${errorMsg}`);
      console.error('服务器错误详情:', error.response?.data);
    } else {
      ElMessage.error(error.response?.data?.error || '删除失败');
    }
  }
};

const addCategory = async () => {
  if (!newCategoryName.value.trim()) {
    ElMessage.warning('请输入分类名称');
    return;
  }

  try {
    console.log('正在添加分类:', newCategoryName.value);
    // 使用管理员API：/api/admin/categories
    await axios.post('/admin/categories', { name: newCategoryName.value.trim() });
    ElMessage.success('分类添加成功');
    newCategoryName.value = '';
    await fetchCategories();
  } catch (e: any) {
    console.error('添加分类错误:', e);
    if (e.response?.status === 401) {
      ElMessage.error('请先登录');
    } else if (e.response?.status === 403) {
      ElMessage.error('没有管理员权限');
    } else {
      ElMessage.error(e.response?.data?.error || '添加失败');
    }
  }
};

const deleteCategory = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这个分类吗？相关文章将变为未分类', '警告', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    });
    
    console.log('正在删除分类:', id);
    // 使用管理员API：/api/admin/categories/:id
    await axios.delete(`/admin/categories/${id}`);
    ElMessage.success('分类已删除');
    await fetchCategories();
  } catch (error: any) {
    if (error === 'cancel') return;
    console.error('删除分类错误:', error);
    if (error.response?.status === 401) {
      ElMessage.error('请先登录');
    } else if (error.response?.status === 403) {
      ElMessage.error('没有管理员权限');
    } else {
      ElMessage.error(error.response?.data?.error || '删除失败');
    }
  }
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

const formatDate = (dateStr: string) => {
  if (!dateStr || dateStr.startsWith('0001')) return '刚刚';
  return new Date(dateStr).toLocaleString('zh-CN');
};

onMounted(() => {
  console.log('AdminArticleView 组件已挂载');
  fetchData();
});
</script>

<style scoped>
.admin-container { 
  padding: 20px; 
  background-color: #f5f7fa; 
  min-height: calc(100vh - 60px); 
}
.admin-card { 
  max-width: 1200px; 
  margin: 0 auto; 
  width: 100%;
}
.card-header { 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
}
.header-actions { 
  display: flex; 
  gap: 10px; 
}
.pagination { 
  margin-top: 20px; 
  text-align: right; 
}
.category-input { 
  display: flex; 
  gap: 10px; 
  margin-bottom: 20px; 
}
</style>