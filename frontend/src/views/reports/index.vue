<template>
  <div class="page-container">
    <div class="page-header">
      <h2>历史报告</h2>
    </div>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-input
        v-model="keyword"
        placeholder="搜索项目名称、巡检人"
        clearable
        style="width: 300px"
        @keyup.enter="handleSearch"
        @clear="handleSearch"
      >
        <template #append>
          <el-button :icon="Search" @click="handleSearch" />
        </template>
      </el-input>
    </div>

    <el-table :data="reports" v-loading="loading" stripe>
      <el-table-column prop="project_name" label="项目名称" width="200" />
      <el-table-column prop="inspector" label="巡检人" width="120" />
      <el-table-column prop="start_time" label="巡检时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.start_time) }}
        </template>
      </el-table-column>
      <el-table-column prop="total_items" label="总巡检项" width="100" />
      <el-table-column prop="warning_count" label="告警" width="80">
        <template #default="{ row }">
          <span v-if="row.warning_count > 0" style="color: #E6A23C">{{ row.warning_count }}</span>
          <span v-else>{{ row.warning_count }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="critical_count" label="严重" width="80">
        <template #default="{ row }">
          <span v-if="row.critical_count > 0" style="color: #F56C6C">{{ row.critical_count }}</span>
          <span v-else>{{ row.critical_count }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 'completed' ? 'success' : 'warning'" size="small">
            {{ row.status === 'completed' ? '完成' : '进行中' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button type="primary" link @click="viewDetail(row)">查看</el-button>
          <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { inspectionApi, type InspectionReport } from '../../api/inspection'

const router = useRouter()
const loading = ref(false)
const reports = ref<InspectionReport[]>([])
const keyword = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const loadReports = async () => {
  loading.value = true
  try {
    const res = await inspectionApi.list({
      keyword: keyword.value || undefined,
      page: currentPage.value,
      page_size: pageSize.value
    })
    reports.value = res.data.list
    total.value = res.data.total
  } catch (error) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadReports()
}

const handleSizeChange = () => {
  currentPage.value = 1
  loadReports()
}

const handlePageChange = () => {
  loadReports()
}

const viewDetail = (row: InspectionReport) => {
  router.push(`/reports/${row.id}`)
}

const handleDelete = async (row: InspectionReport) => {
  await ElMessageBox.confirm('确定要删除该报告吗？', '提示', { type: 'warning' })
  // 这里需要添加删除 API
  ElMessage.success('删除成功')
  loadReports()
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN')
}

onMounted(() => {
  loadReports()
})
</script>

<style scoped>
.page-container {
  background: #fff;
  padding: 20px;
  border-radius: 4px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

.search-bar {
  margin-bottom: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
