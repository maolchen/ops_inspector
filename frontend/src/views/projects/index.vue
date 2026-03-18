<template>
  <div class="page-container">
    <div class="page-header">
      <h2>项目管理</h2>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        新增项目
      </el-button>
    </div>

    <el-table :data="projects" v-loading="loading" stripe>
      <el-table-column prop="name" label="项目名称" width="200" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="prometheus_url" label="Prometheus 地址" width="300" />
      <el-table-column prop="token" label="Token" width="150">
        <template #default="{ row }">
          <span v-if="row.token">{{ row.token }}</span>
          <span v-else style="color: #999">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
          <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      v-model="dialogVisible"
      :title="editMode ? '编辑项目' : '新增项目'"
      width="500px"
    >
      <el-form :model="form" label-width="120px" :rules="rules" ref="formRef">
        <el-form-item label="项目名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入项目名称" />
        </el-form-item>
        <el-form-item label="项目描述">
          <el-input v-model="form.description" type="textarea" placeholder="请输入项目描述" />
        </el-form-item>
        <el-form-item label="Prometheus地址" prop="prometheus_url">
          <el-input v-model="form.prometheus_url" placeholder="http://localhost:9090" />
        </el-form-item>
        <el-form-item label="Token">
          <el-input
            v-model="form.token"
            type="password"
            placeholder="可选，用于认证"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { projectApi, type Project } from '../../api/project'

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const editMode = ref(false)
const projects = ref<Project[]>([])
const formRef = ref()

const form = ref<Partial<Project>>({
  name: '',
  description: '',
  prometheus_url: '',
  token: ''
})

const rules = {
  name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
  prometheus_url: [{ required: true, message: '请输入 Prometheus 地址', trigger: 'blur' }]
}

const loadProjects = async () => {
  loading.value = true
  try {
    const res = await projectApi.list()
    projects.value = res.data
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  editMode.value = false
  form.value = {
    name: '',
    description: '',
    prometheus_url: '',
    token: ''
  }
  dialogVisible.value = true
}

const handleEdit = (row: Project) => {
  editMode.value = true
  form.value = { ...row }
  dialogVisible.value = true
}

const handleDelete = async (row: Project) => {
  await ElMessageBox.confirm('确定要删除该项目吗？', '提示', {
    type: 'warning'
  })
  await projectApi.delete(row.id)
  ElMessage.success('删除成功')
  loadProjects()
}

const handleSubmit = async () => {
  await formRef.value.validate()
  submitting.value = true
  try {
    if (editMode.value) {
      await projectApi.update(form.value.id!, form.value)
      ElMessage.success('更新成功')
    } else {
      await projectApi.create(form.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadProjects()
  } finally {
    submitting.value = false
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN')
}

onMounted(() => {
  loadProjects()
})
</script>

<style scoped>
/* 页面样式已由全局 shadcn.css 统一管理 */
</style>
