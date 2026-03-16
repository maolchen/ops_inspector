<template>
  <div class="page-container" id="report-content">
    <div class="page-header">
      <h2>巡检报告 - {{ report?.project_name }}</h2>
      <div>
        <el-button type="primary" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出PDF
        </el-button>
        <el-button @click="goBack">返回</el-button>
      </div>
    </div>

    <div v-loading="loading">
      <!-- 巡检总览 -->
      <el-card class="section-card">
        <template #header>
          <span class="section-title">巡检总览</span>
        </template>
        <div class="overview-cards">
          <div
            v-for="group in groupStats"
            :key="group.name"
            class="overview-card"
          >
            <div class="card-title">{{ group.name }}</div>
            <div class="card-stats">
              <div class="stat-item">
                <span class="stat-label">总计</span>
                <span class="stat-value">{{ group.total }}</span>
              </div>
              <div class="stat-item" v-if="group.critical > 0">
                <span class="stat-label">严重</span>
                <span class="stat-value critical">{{ group.critical }}</span>
              </div>
              <div class="stat-item" v-if="group.warning > 0">
                <span class="stat-label">告警</span>
                <span class="stat-value warning">{{ group.warning }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">正常</span>
                <span class="stat-value normal">{{ group.normal }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 资源趋势 -->
      <el-card class="section-card">
        <template #header>
          <span class="section-title">资源趋势（最近7天）</span>
        </template>
        <div class="trend-charts">
          <div class="chart-container">
            <div ref="cpuChart" class="chart"></div>
          </div>
          <div class="chart-container">
            <div ref="memChart" class="chart"></div>
          </div>
          <div class="chart-container">
            <div ref="diskChart" class="chart"></div>
          </div>
        </div>
      </el-card>

      <!-- 基础资源详情表 -->
      <el-card class="section-card" v-if="basicTableData.length > 0">
        <template #header>
          <span class="section-title">基础资源详情</span>
        </template>
        <el-table :data="basicTableData" stripe border size="small">
          <el-table-column prop="instance" label="节点IP" width="150" fixed="left" />
          <el-table-column
            v-for="col in basicColumns"
            :key="col.prop"
            :prop="col.prop"
            :label="col.label"
            :width="col.width"
          >
            <template #default="{ row }">
              <span :class="getStatusClass(row[col.prop + '_status'])">
                {{ row[col.prop] }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 分组详情 -->
      <el-card
        v-for="groupDetail in groupDetails"
        :key="groupDetail.group_id"
        class="section-card"
      >
        <template #header>
          <span class="section-title">{{ groupDetail.group_name }}</span>
        </template>
        <div
          v-for="ruleDetail in groupDetail.rules"
          :key="ruleDetail.rule_name"
          class="rule-section"
        >
          <div class="rule-title">{{ ruleDetail.rule_name }}</div>
          <el-table :data="ruleDetail.items" stripe size="small" border>
            <el-table-column
              v-for="col in ruleDetail.columns"
              :key="col.prop"
              :prop="col.prop"
              :label="col.label"
              :width="col.width"
            >
              <template #default="{ row }">
                <span :class="getStatusClass(row.status)">
                  {{ row[col.prop] }}{{ row.unit ? ` ${row.unit}` : '' }}
                </span>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-card>

      <!-- 巡检总结 -->
      <el-card class="section-card">
        <template #header>
          <span class="section-title">巡检总结</span>
        </template>
        <el-form :model="summaryForm" label-width="100px">
          <el-form-item label="巡检人">
            <el-input :value="report?.inspector || ''" disabled style="width: 200px" />
          </el-form-item>
          <el-form-item label="巡检时间">
            <el-input :value="formatDate(report?.start_time)" disabled style="width: 300px" />
          </el-form-item>
          <el-form-item label="巡检详情">
            <el-input
              v-model="summaryForm.summary"
              type="textarea"
              :rows="3"
              placeholder="请输入巡检详情"
            />
          </el-form-item>
          <el-form-item label="备注">
            <el-input
              v-model="summaryForm.remark"
              type="textarea"
              :rows="3"
              placeholder="请输入备注信息"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSaveSummary">保存总结</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import html2canvas from 'html2canvas'
import jsPDF from 'jspdf'
import { inspectionApi, type InspectionReport, type InspectionItem } from '../../api/inspection'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const report = ref<InspectionReport | null>(null)
const items = ref<InspectionItem[]>([])
const cpuChart = ref<HTMLElement>()
const memChart = ref<HTMLElement>()
const diskChart = ref<HTMLElement>()

const summaryForm = ref({
  summary: '',
  remark: ''
})

// 按分组统计
const groupStats = computed(() => {
  const stats: Record<string, { name: string; total: number; critical: number; warning: number; normal: number }> = {}
  
  items.value.forEach(item => {
    if (!stats[item.group_name]) {
      stats[item.group_name] = {
        name: item.group_name,
        total: 0,
        critical: 0,
        warning: 0,
        normal: 0
      }
    }
    
    stats[item.group_name].total++
    if (item.status === 'critical') {
      stats[item.group_name].critical++
    } else if (item.status === 'warning') {
      stats[item.group_name].warning++
    } else {
      stats[item.group_name].normal++
    }
  })
  
  return Object.values(stats)
})

// 基础资源表格数据
const basicTableData = computed(() => {
  const basicItems = items.value.filter(i => i.show_in_table)
  const instanceMap: Record<string, any> = {}
  
  basicItems.forEach(item => {
    if (!instanceMap[item.instance]) {
      instanceMap[item.instance] = { instance: item.instance }
    }
    const value = item.value.toFixed(2)
    instanceMap[item.instance][item.rule_name] = item.unit ? `${value}${item.unit}` : value
    instanceMap[item.instance][item.rule_name + '_status'] = item.status
  })
  
  return Object.values(instanceMap)
})

// 基础资源表格列
const basicColumns = computed(() => {
  const cols: { prop: string; label: string; width: number }[] = []
  const basicItems = items.value.filter(i => i.show_in_table)
  const ruleNames = [...new Set(basicItems.map(i => i.rule_name))]
  
  ruleNames.forEach(name => {
    cols.push({ prop: name, label: name, width: 120 })
  })
  
  return cols
})

// 分组详情
const groupDetails = computed(() => {
  const nonBasicItems = items.value.filter(i => !i.show_in_table)
  const groups: Record<number, { group_id: number; group_name: string; rules: any[] }> = {}
  
  nonBasicItems.forEach(item => {
    if (!groups[item.group_id]) {
      groups[item.group_id] = {
        group_id: item.group_id,
        group_name: item.group_name,
        rules: []
      }
    }
    
    let ruleGroup = groups[item.group_id].rules.find(r => r.rule_name === item.rule_name)
    if (!ruleGroup) {
      // 解析 labels 获取列信息
      let columns = [{ prop: 'instance', label: '实例', width: 150 }]
      if (item.labels) {
        try {
          const labels = JSON.parse(item.labels)
          Object.keys(labels).forEach(key => {
            if (key !== 'instance' && key !== '__name__') {
              columns.push({ prop: key, label: labels[key] || key, width: 120 })
            }
          })
        } catch (e) {}
      }
      columns.push({ prop: 'value', label: '值', width: 100 })
      
      ruleGroup = {
        rule_name: item.rule_name,
        columns,
        items: []
      }
      groups[item.group_id].rules.push(ruleGroup)
    }
    
    // 添加数据行
    let row: any = {
      instance: item.instance,
      value: item.value.toFixed(2),
      unit: item.unit,
      status: item.status
    }
    
    if (item.labels) {
      try {
        const labels = JSON.parse(item.labels)
        Object.assign(row, labels)
      } catch (e) {}
    }
    
    ruleGroup.items.push(row)
  })
  
  return Object.values(groups)
})

const loadReport = async () => {
  const id = Number(route.params.id)
  loading.value = true
  try {
    const res = await inspectionApi.get(id)
    report.value = res.data.report
    items.value = res.data.items
    summaryForm.value.summary = report.value?.summary || ''
    summaryForm.value.remark = report.value?.remark || ''
    
    // 渲染图表
    nextTick(() => {
      renderCharts()
    })
  } finally {
    loading.value = false
  }
}

const renderCharts = () => {
  // 模拟趋势数据
  const dates = []
  const now = new Date()
  for (let i = 6; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    dates.push(d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }))
  }
  
  // CPU 趋势
  if (cpuChart.value) {
    const chart = echarts.init(cpuChart.value)
    chart.setOption({
      title: { text: 'CPU 使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{
        type: 'line',
        data: [45, 52, 48, 62, 58, 55, 60],
        smooth: true,
        itemStyle: { color: '#409EFF' },
        areaStyle: { color: 'rgba(64, 158, 255, 0.2)' }
      }]
    })
  }
  
  // 内存趋势
  if (memChart.value) {
    const chart = echarts.init(memChart.value)
    chart.setOption({
      title: { text: '内存使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{
        type: 'line',
        data: [60, 65, 62, 70, 68, 72, 75],
        smooth: true,
        itemStyle: { color: '#67C23A' },
        areaStyle: { color: 'rgba(103, 194, 58, 0.2)' }
      }]
    })
  }
  
  // 磁盘趋势
  if (diskChart.value) {
    const chart = echarts.init(diskChart.value)
    chart.setOption({
      title: { text: '磁盘使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series: [{
        type: 'line',
        data: [55, 56, 57, 58, 59, 60, 61],
        smooth: true,
        itemStyle: { color: '#E6A23C' },
        areaStyle: { color: 'rgba(230, 162, 60, 0.2)' }
      }]
    })
  }
}

const handleExport = async () => {
  const element = document.getElementById('report-content')
  if (!element) return
  
  try {
    const canvas = await html2canvas(element, {
      scale: 2,
      useCORS: true,
      backgroundColor: '#fff'
    })
    
    const imgWidth = 210
    const imgHeight = (canvas.height * imgWidth) / canvas.width
    
    const pdf = new jsPDF('p', 'mm', 'a4')
    pdf.addImage(
      canvas.toDataURL('image/png'),
      'PNG',
      0,
      0,
      imgWidth,
      imgHeight
    )
    
    pdf.save(`巡检报告_${report.value?.project_name}_${new Date().toISOString().slice(0, 10)}.pdf`)
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

const handleSaveSummary = async () => {
  if (!report.value) return
  
  await inspectionApi.updateSummary(
    report.value.id,
    summaryForm.value.summary,
    summaryForm.value.remark
  )
  ElMessage.success('保存成功')
}

const getStatusClass = (status?: string) => {
  if (status === 'critical') return 'status-critical'
  if (status === 'warning') return 'status-warning'
  return ''
}

const formatDate = (date?: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const goBack = () => {
  router.push('/reports')
}

onMounted(() => {
  loadReport()
})
</script>

<style scoped>
.page-container {
  background: #fff;
  padding: 20px;
  border-radius: 4px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

.section-card {
  margin-bottom: 20px;
}

.section-title {
  font-weight: bold;
  font-size: 16px;
}

.overview-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.overview-card {
  flex: 1;
  min-width: 200px;
  padding: 15px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  background: #fafafa;
}

.card-title {
  font-weight: bold;
  margin-bottom: 10px;
  font-size: 14px;
}

.card-stats {
  display: flex;
  gap: 15px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-label {
  font-size: 12px;
  color: #999;
}

.stat-value {
  font-size: 20px;
  font-weight: bold;
}

.stat-value.critical {
  color: #F56C6C;
}

.stat-value.warning {
  color: #E6A23C;
}

.stat-value.normal {
  color: #67C23A;
}

.trend-charts {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.chart-container {
  flex: 1;
  min-width: 300px;
}

.chart {
  width: 100%;
  height: 250px;
}

.rule-section {
  margin-bottom: 20px;
}

.rule-title {
  font-weight: bold;
  margin-bottom: 10px;
  color: #606266;
}

.status-critical {
  color: #F56C6C;
  font-weight: bold;
}

.status-warning {
  color: #E6A23C;
  font-weight: bold;
}
</style>
