<template>
  <div class="page-container">
    <div class="page-header">
      <h2>执行巡检</h2>
    </div>

    <el-card>
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item label="选择项目" prop="project_id">
          <el-select v-model="form.project_id" placeholder="请选择要巡检的项目" style="width: 300px">
            <el-option
              v-for="p in projects"
              :key="p.id"
              :label="p.name"
              :value="p.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="巡检人">
          <el-input v-model="form.inspector" placeholder="请输入巡检人姓名" style="width: 300px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleStart" :loading="inspecting">
            <el-icon><Search /></el-icon>
            开始巡检
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="currentReport" style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>巡检结果</span>
          <el-button type="primary" @click="viewReport">查看详情</el-button>
        </div>
      </template>

      <el-descriptions :column="4" border>
        <el-descriptions-item label="巡检项目">{{ currentReport.project_name }}</el-descriptions-item>
        <el-descriptions-item label="巡检人">{{ currentReport.inspector || '-' }}</el-descriptions-item>
        <el-descriptions-item label="巡检时间">{{ formatDate(currentReport.start_time) }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentReport.status === 'completed' ? 'success' : 'warning'">
            {{ currentReport.status === 'completed' ? '已完成' : '进行中' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="总巡检项">{{ currentReport.total_items }}</el-descriptions-item>
        <el-descriptions-item label="告警数量">
          <span style="color: #E6A23C">{{ currentReport.warning_count }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="严重数量">
          <span style="color: #F56C6C">{{ currentReport.critical_count }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="正常数量">
          <span style="color: #67C23A">{{ currentReport.total_items - currentReport.warning_count - currentReport.critical_count }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { projectApi, type Project } from '../../api/project'
import { inspectionApi, type InspectionReport } from '../../api/inspection'

const router = useRouter()
const formRef = ref()
const projects = ref<Project[]>([])
const inspecting = ref(false)
const currentReport = ref<InspectionReport | null>(null)

const form = ref({
  project_id: undefined as number | undefined,
  inspector: ''
})

const rules = {
  project_id: [{ required: true, message: '请选择项目', trigger: 'change' }]
}

const loadProjects = async () => {
  const res = await projectApi.list()
  projects.value = res.data
}

const handleStart = async () => {
  await formRef.value.validate()
  
  inspecting.value = true
  try {
    const res = await inspectionApi.start(form.value.project_id!, form.value.inspector)
    currentReport.value = res.data
    ElMessage.success('巡检完成')
  } catch (error) {
    console.error(error)
  } finally {
    inspecting.value = false
  }
}

const viewReport = () => {
  if (currentReport.value) {
    router.push(`/reports/${currentReport.value.id}`)
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
/* 页面样式由全局 shadcn.css 统一管理 */
</style>
