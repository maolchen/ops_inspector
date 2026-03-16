<template>
  <div class="page-container">
    <div class="page-header">
      <h2>系统设置</h2>
    </div>

    <el-card v-loading="loading">
      <template #header>
        <span>报告保留设置</span>
      </template>

      <el-form :model="form" label-width="200px">
        <el-form-item label="报告保留天数">
          <el-input-number
            v-model="form.retentionDays"
            :min="1"
            :max="365"
            controls-position="right"
          />
          <span class="form-tip">天（超过此天数的报告将自动清理）</span>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">保存设置</el-button>
          <el-button @click="handleCleanup" :loading="cleaning">立即清理过期报告</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>系统信息</span>
      </template>

      <el-descriptions :column="1" border>
        <el-descriptions-item label="数据库类型">SQLite</el-descriptions-item>
        <el-descriptions-item label="数据库路径">./data/inspection.db</el-descriptions-item>
        <el-descriptions-item label="自动清理">每天凌晨3点执行</el-descriptions-item>
        <el-descriptions-item label="清理状态">
          {{ cleanupInfo }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { systemApi } from '../../api/inspection'

const loading = ref(false)
const saving = ref(false)
const cleaning = ref(false)
const cleanupInfo = ref('未执行')

const form = ref({
  retentionDays: 30
})

const loadConfig = async () => {
  loading.value = true
  try {
    const res = await systemApi.getConfigs()
    form.value.retentionDays = parseInt(res.data.report_retention_days) || 30
  } catch (error) {
    ElMessage.error('加载配置失败')
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  saving.value = true
  try {
    await systemApi.updateConfig('report_retention_days', form.value.retentionDays.toString())
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const handleCleanup = async () => {
  cleaning.value = true
  try {
    const res = await systemApi.cleanupReports(form.value.retentionDays)
    cleanupInfo.value = `已清理 ${res.count} 个过期报告`
    ElMessage.success(`清理完成，共删除 ${res.count} 个过期报告`)
  } catch (error) {
    ElMessage.error('清理失败')
  } finally {
    cleaning.value = false
  }
}

onMounted(() => {
  loadConfig()
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

.form-tip {
  margin-left: 10px;
  color: #999;
  font-size: 12px;
}
</style>
