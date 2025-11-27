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
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="Title" label="标题" show-overflow-tooltip />
        <el-table-column label="分类" width="120">
          <template #default="scope">
            <el-tag size="small">{{ getCategoryName(scope.row.CategoryID) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ViewsCount" label="浏览" width="80" />
        <el-table-column prop="CreatedAt" label="发布时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.CreatedAt) }}
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
         <el-pagination layout="prev, pager, next" :total="100" /> </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑文章' : '发布文章'" width="700px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="标题">
          <el-input v-model="form.title" placeholder="请输入文章标题" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="form.categoryId" placeholder="选择分类" style="width: 100%">
            <el-option v-for="c in categories" :key="c.ID" :label="c.Name" :value="c.ID" />
          </el-select>
        </el-form-item>
        <el-form-item label="封面URL">
          <el-input v-model="form.cover" placeholder="图片地址" />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="form.preview" type="textarea" :rows="2" placeholder="文章预览文字" />
        </el-form-item>
        <el-form-item label="内容">
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
        <el-table-column prop="ID" label="ID" width="60" />
        <el-table-column prop="Name" label="名称" />
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-button type="danger" link @click="deleteCategory(scope.row.ID)">删除</el-button>
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

const articles = ref([]);
const categories = ref<any[]>([]);
const loading = ref(false);
const dialogVisible = ref(false);
const categoryDialogVisible = ref(false);
const isEdit = ref(false);
const currentId = ref<number | null>(null);

const form = reactive({
  title: '',
  content: '',
  preview: '',
  cover: '',
  categoryId: null as number | null
});

const newCategoryName = ref('');

const fetchData = async () => {
  loading.value = true;
  try {
    const [artRes, catRes] = await Promise.all([
      axios.get('/articles?limit=100'), // Get all for admin demo
      axios.get('/categories')
    ]);
    articles.value = artRes.data;
    categories.value = catRes.data;
  } catch (e) {
    ElMessage.error('加载数据失败');
  } finally {
    loading.value = false;
  }
};

const getCategoryName = (id: number) => {
  const cat = categories.value.find(c => c.ID === id);
  return cat ? cat.Name : '未分类';
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

const handleEdit = (row: any) => {
  isEdit.value = true;
  currentId.value = row.ID;
  form.title = row.Title;
  form.content = row.Content;
  form.preview = row.Preview;
  form.cover = row.Cover;
  form.categoryId = row.CategoryID;
  dialogVisible.value = true;
};

const submitArticle = async () => {
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
    fetchData();
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '操作失败');
  }
};

const handleDelete = (row: any) => {
  ElMessageBox.confirm('确定要删除这篇文章吗?', '警告', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      await axios.delete(`/admin/articles/${row.ID}`);
      ElMessage.success('删除成功');
      fetchData();
    } catch (e) {
      ElMessage.error('删除失败');
    }
  });
};

const addCategory = async () => {
  try {
    await axios.post('/admin/categories', { name: newCategoryName.value });
    ElMessage.success('分类添加成功');
    newCategoryName.value = '';
    const res = await axios.get('/categories');
    categories.value = res.data;
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '添加失败');
  }
};

const deleteCategory = async (id: number) => {
  try {
    await axios.delete(`/admin/categories/${id}`);
    ElMessage.success('分类已删除');
    const res = await axios.get('/categories');
    categories.value = res.data;
  } catch (e) {
    ElMessage.error('删除失败');
  }
};

const formatDate = (dateStr: string) => {
  if (!dateStr || dateStr.startsWith('0001')) return '刚刚';
  return new Date(dateStr).toLocaleString();
};

onMounted(fetchData);
</script>

<style scoped>
.admin-container { padding: 20px; background-color: #f5f7fa; min-height: calc(100vh - 60px); }
.admin-card { max-width: 1200px; margin: 0 auto; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.header-actions { display: flex; gap: 10px; }
.pagination { margin-top: 20px; text-align: right; }
.category-input { display: flex; gap: 10px; margin-bottom: 20px; }
</style>