<template>
  <el-container class="admin-container">
    <el-card class="admin-card">
      <template #header>
        <div class="card-header">
          <h2>文章与分类管理</h2>
          <div class="header-actions">
            <!-- 绑定点击事件打开分类弹窗 -->
            <el-button type="success" @click="categoryDialogVisible = true">管理分类</el-button>
            <el-button type="primary" @click="openCreateDialog">发布文章</el-button>
          </div>
        </div>
      </template>

      <el-table :data="articles" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        
        <el-table-column label="封面" width="100">
          <template #default="scope">
            <el-image 
              v-if="scope.row.cover"
              style="width: 60px; height: 40px; border-radius: 4px"
              :src="getImageUrl(scope.row.cover)"
              :preview-src-list="[getImageUrl(scope.row.cover)]"
              preview-teleported
              fit="cover"
            />
            <span v-else style="color:#909399;font-size:12px">无封面</span>
          </template>
        </el-table-column>

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

    <!-- 文章 编辑/发布 弹窗 -->
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
        
        <el-form-item label="封面">
          <el-upload
            class="cover-uploader"
            action="#" 
            :http-request="handleUpload"
            :show-file-list="false"
            accept="image/jpeg,image/png,image/gif"
          >
            <div v-if="form.cover" class="uploaded-cover-box">
              <img :src="getImageUrl(form.cover)" class="cover-img" />
              <div class="re-upload-mask">点击更换</div>
            </div>
            <el-icon v-else class="cover-uploader-icon"><Plus /></el-icon>
          </el-upload>
          <div style="font-size:12px; color:#999; margin-top:5px;">支持 jpg/png/gif，建议尺寸 16:9</div>
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

    <!-- 【功能补充】分类管理弹窗 -->
    <el-dialog v-model="categoryDialogVisible" title="分类管理" width="500px">
      <div class="category-input">
        <el-input 
          v-model="newCategoryName" 
          placeholder="请输入新分类名称" 
          @keyup.enter="addCategory"
        />
        <el-button type="primary" @click="addCategory" :disabled="!newCategoryName.trim()">添加</el-button>
      </div>
      
      <el-table :data="categories" height="300px" style="width: 100%" stripe border>
        <el-table-column prop="id" label="ID" width="80" align="center" />
        <el-table-column prop="name" label="分类名称" />
        <el-table-column label="操作" width="100" align="center">
          <template #default="scope">
            <el-popconfirm 
              title="确定删除该分类吗？"
              confirm-button-text="删除"
              cancel-button-text="取消"
              @confirm="deleteCategory(scope.row.id)"
            >
              <template #reference>
                <el-button type="danger" link size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </el-container>
</template>

<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue';
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
const categoryDialogVisible = ref(false); // 控制分类弹窗显示
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

// 图片 URL 拼接
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

// 封面上传处理
const handleUpload = async (options: any) => {
  const { file } = options;
  const formData = new FormData();
  formData.append('file', file);

  try {
    const res = await axios.post('/admin/articles/upload/cover', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    form.cover = res.data.url; 
    ElMessage.success('封面上传成功');
  } catch (error: any) {
    console.error('上传失败', error);
    ElMessage.error(error.response?.data?.error || '上传失败');
  }
};

const fetchArticles = async () => {
  loading.value = true;
  try {
    const response = await axios.get(`/articles?page=${currentPage.value}&limit=${pageSize.value}`);
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
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '加载文章失败');
  } finally {
    loading.value = false;
  }
};

const fetchCategories = async () => {
  try {
    const response = await axios.get('/categories');
    categories.value = response.data;
  } catch (e: any) {
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
    
    if (isEdit.value && currentId.value) {
      await axios.put(`/admin/articles/${currentId.value}`, payload);
      ElMessage.success('更新成功');
    } else {
      await axios.post('/admin/articles', payload);
      ElMessage.success('发布成功');
    }
    dialogVisible.value = false;
    await fetchArticles();
  } catch (e: any) {
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
    await axios.delete(`/admin/articles/${row.id}`);
    ElMessage.success('删除成功');
    await fetchArticles();
  } catch (error: any) {
    if (error === 'cancel') return;
    if (error.response?.status === 403) {
      ElMessage.error('没有管理员权限');
    } else {
      ElMessage.error(error.response?.data?.error || '删除失败');
    }
  }
};

// 【功能补充】添加分类
const addCategory = async () => {
  if (!newCategoryName.value.trim()) {
    ElMessage.warning('请输入分类名称');
    return;
  }
  try {
    await axios.post('/admin/categories', { name: newCategoryName.value.trim() });
    ElMessage.success('分类添加成功');
    newCategoryName.value = ''; // 清空输入框
    await fetchCategories(); // 刷新列表
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '分类添加失败');
  }
};

// 【功能补充】删除分类
const deleteCategory = async (id: number) => {
  try {
    await axios.delete(`/admin/categories/${id}`);
    ElMessage.success('分类已删除');
    await fetchCategories(); // 刷新列表
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '删除失败，可能包含关联文章');
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
/* 分类管理弹窗样式 */
.category-input { 
  display: flex; 
  gap: 10px; 
  margin-bottom: 20px; 
}

/* 上传组件样式 */
.cover-uploader {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  width: 178px;
  height: 100px;
  display: flex;
  justify-content: center;
  align-items: center;
  transition: border-color 0.3s;
}
.cover-uploader:hover {
  border-color: #409EFF;
}
.cover-uploader-icon {
  font-size: 28px;
  color: #8c939d;
}
.uploaded-cover-box {
  width: 100%;
  height: 100%;
  position: relative;
}
.cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.re-upload-mask {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(0,0,0,0.5);
  color: #fff;
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  transition: opacity 0.3s;
}
.uploaded-cover-box:hover .re-upload-mask {
  opacity: 1;
}
</style>