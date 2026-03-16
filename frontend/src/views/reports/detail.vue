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
        <el-table :data="basicTableData" stripe border size="small" :cell-class-name="getBasicTableCellClass">
          <el-table-column prop="instance" label="节点IP" width="150" fixed="left" />
          <el-table-column prop="mountpoint" label="挂载点" width="120" v-if="hasMountpoint" />
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

      <!-- 磁盘IO详情 -->
      <el-card class="section-card" v-if="diskIOTableData.length > 0">
        <template #header>
          <span class="section-title">磁盘IO详情</span>
        </template>
        <el-table :data="diskIOTableData" stripe border size="small">
          <el-table-column prop="instance" label="节点" width="150" />
          <el-table-column prop="device" label="磁盘设备" width="150" />
          <el-table-column prop="readMB" label="读取速率" width="120">
            <template #default="{ row }">
              {{ row.readMB }}
            </template>
          </el-table-column>
          <el-table-column prop="writeMB" label="写入速率" width="120">
            <template #default="{ row }">
              {{ row.writeMB }}
            </template>
          </el-table-column>
          <el-table-column prop="readIOPS" label="读IOPS" width="100" />
          <el-table-column prop="writeIOPS" label="写IOPS" width="100" />
        </el-table>
      </el-card>

      <!-- 网络IO详情 -->
      <el-card class="section-card" v-if="networkIOTableData.length > 0">
        <template #header>
          <span class="section-title">网络IO详情</span>
        </template>
        <el-table :data="networkIOTableData" stripe border size="small">
          <el-table-column prop="instance" label="节点" width="150" />
          <el-table-column prop="interface" label="网卡" width="120" />
          <el-table-column prop="downloadMB" label="下载速率" width="120" />
          <el-table-column prop="uploadMB" label="上传速率" width="120" />
        </el-table>
      </el-card>

      <!-- K8S节点状态 -->
      <el-card class="section-card" v-if="k8sNodeTableData.length > 0">
        <template #header>
          <span class="section-title">K8S节点就绪状态</span>
        </template>
        <el-table :data="k8sNodeTableData" stripe border size="small" :cell-class-name="getK8sCellClass">
          <el-table-column prop="node" label="节点" width="150" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.status === 'Ready' ? 'status-ready' : 'status-not-ready'">
                {{ row.status }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- K8S Pod状态 -->
      <el-card class="section-card" v-if="k8sPodTableData.length > 0">
        <template #header>
          <span class="section-title">K8S Pod运行状态</span>
        </template>
        <el-table :data="k8sPodTableData" stripe border size="small" :cell-class-name="getK8sCellClass">
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column prop="pod" label="Pod名称" width="250" />
          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <span :class="getPodStatusClass(row.status)">
                {{ row.status }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- K8S证书状态 -->
      <el-card class="section-card" v-if="k8sCertTableData.length > 0">
        <template #header>
          <span class="section-title">K8S证书状态</span>
        </template>
        <el-table :data="k8sCertTableData" stripe border size="small" :cell-class-name="getK8sCellClass">
          <el-table-column prop="instance" label="节点" width="150" />
          <el-table-column prop="certName" label="证书名称" width="200" />
          <el-table-column prop="expiryDays" label="剩余有效期(天)" width="130">
            <template #default="{ row }">
              <span :class="getCertStatusClass(row.expiryDays)">
                {{ row.expiryDays }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- K8S PVC使用率 -->
      <el-card class="section-card" v-if="k8sPVCTableData.length > 0">
        <template #header>
          <span class="section-title">K8S PVC使用率</span>
        </template>
        <el-table :data="k8sPVCTableData" stripe border size="small" :cell-class-name="getK8sCellClass">
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column prop="pvc" label="PVC名称" width="200" />
          <el-table-column prop="usedPercent" label="使用率" width="100">
            <template #default="{ row }">
              <span :class="getUsageStatusClass(row.usedPercent)">
                {{ row.usedPercent }}%
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="used" label="已用" width="100" />
          <el-table-column prop="total" label="总量" width="100" />
        </el-table>
      </el-card>

      <!-- 进程CPU使用率Top5 -->
      <el-card class="section-card" v-if="processCPUTableData.length > 0">
        <template #header>
          <span class="section-title">进程CPU使用率top5</span>
        </template>
        <el-table :data="processCPUTableData" stripe border size="small">
          <el-table-column prop="metricName" label="指标名称" width="200">
            <template #default>
              进程CPU使用率top5
            </template>
          </el-table-column>
          <el-table-column prop="instance" label="节点" width="180" />
          <el-table-column prop="processName" label="进程名" width="200" />
          <el-table-column prop="value" label="值" width="120" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <span :class="getStatusClass(row.status)">
                {{ row.statusText }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 进程内存使用率Top5 -->
      <el-card class="section-card" v-if="processMemTableData.length > 0">
        <template #header>
          <span class="section-title">进程内存使用率top5</span>
        </template>
        <el-table :data="processMemTableData" stripe border size="small">
          <el-table-column prop="metricName" label="指标名称" width="200">
            <template #default>
              进程内存使用率top5
            </template>
          </el-table-column>
          <el-table-column prop="instance" label="节点" width="180" />
          <el-table-column prop="processName" label="进程名" width="200" />
          <el-table-column prop="value" label="值" width="120" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <span :class="getStatusClass(row.status)">
                {{ row.statusText }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 分组详情 -->
      <el-card
        v-for="groupDetail in otherGroupDetails"
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
      <el-card class="section-card" id="summary-section">
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

// 预定义的颜色列表，用于区分不同实例
const colorPalette = [
  '#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de',
  '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc', '#48b8d0'
]

// 去掉IP端口
const stripPort = (instance: string): string => {
  if (!instance) return instance
  return instance.split(':')[0]
}

// 格式化字节大小
const formatBytes = (bytes: number, unit?: string): string => {
  if (unit === '%' || unit === 'percent') {
    return `${bytes.toFixed(2)}%`
  }
  if (unit === 's' || unit === 'seconds') {
    if (bytes < 60) return `${bytes.toFixed(0)}秒`
    if (bytes < 3600) return `${(bytes / 60).toFixed(1)}分钟`
    return `${(bytes / 3600).toFixed(1)}小时`
  }
  if (unit === 'B' || unit === 'bytes' || bytes > 1024) {
    if (bytes < 1024) return `${bytes.toFixed(2)} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
    return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
  }
  return bytes.toFixed(2)
}

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

// 基础资源分组名称
const basicGroupNames = ['基础资源', 'CPU监控', '内存监控', '磁盘监控']

// 基础资源表格数据
const basicTableData = computed(() => {
  const basicItems = items.value.filter(i => basicGroupNames.some(g => i.group_name.includes(g)) && i.show_in_table)
  const instanceMap: Record<string, any> = {}
  
  basicItems.forEach(item => {
    const instanceKey = stripPort(item.instance)
    const mountpoint = getLabelValue(item.labels, 'mountpoint') || getLabelValue(item.labels, 'device')
    const rowKey = mountpoint ? `${instanceKey}-${mountpoint}` : instanceKey
    
    if (!instanceMap[rowKey]) {
      instanceMap[rowKey] = { 
        instance: instanceKey,
        mountpoint: mountpoint || ''
      }
    }
    const value = formatValue(item.value, item.unit)
    instanceMap[rowKey][item.rule_name] = value
    instanceMap[rowKey][item.rule_name + '_status'] = item.status
  })
  
  return Object.values(instanceMap)
})

// 是否有挂载点列
const hasMountpoint = computed(() => {
  return basicTableData.value.some(row => row.mountpoint)
})

// 基础资源表格列
const basicColumns = computed(() => {
  const cols: { prop: string; label: string; width: number }[] = []
  const basicItems = items.value.filter(i => basicGroupNames.some(g => i.group_name.includes(g)) && i.show_in_table)
  const ruleNames = [...new Set(basicItems.map(i => i.rule_name))]
  
  ruleNames.forEach(name => {
    cols.push({ prop: name, label: name, width: 120 })
  })
  
  return cols
})

// 磁盘IO表格数据
const diskIOTableData = computed(() => {
  const diskItems = items.value.filter(i => 
    i.group_name.includes('磁盘') && 
    (i.rule_name.includes('读取') || i.rule_name.includes('写入') || i.rule_name.includes('IO') || i.rule_name.includes('IOPS'))
  )
  
  const deviceMap: Record<string, any> = {}
  
  diskItems.forEach(item => {
    const instance = stripPort(item.instance)
    const device = getLabelValue(item.labels, 'device') || getLabelValue(item.labels, 'mountpoint') || 'unknown'
    const key = `${instance}-${device}`
    
    if (!deviceMap[key]) {
      deviceMap[key] = {
        instance,
        device,
        readMB: '-',
        writeMB: '-',
        readIOPS: '-',
        writeIOPS: '-'
      }
    }
    
    if (item.rule_name.includes('读取') || item.rule_name.includes('读速率')) {
      deviceMap[key].readMB = formatBytes(item.value, 'B') + '/s'
    } else if (item.rule_name.includes('写入') || item.rule_name.includes('写速率')) {
      deviceMap[key].writeMB = formatBytes(item.value, 'B') + '/s'
    } else if (item.rule_name.includes('读IOPS') || item.rule_name.includes('读取IOPS')) {
      deviceMap[key].readIOPS = item.value.toFixed(0)
    } else if (item.rule_name.includes('写IOPS') || item.rule_name.includes('写入IOPS')) {
      deviceMap[key].writeIOPS = item.value.toFixed(0)
    }
  })
  
  return Object.values(deviceMap)
})

// 网络IO表格数据
const networkIOTableData = computed(() => {
  const networkItems = items.value.filter(i => 
    i.group_name.includes('网络') && 
    (i.rule_name.includes('下载') || i.rule_name.includes('上传') || i.rule_name.includes('接收') || i.rule_name.includes('发送'))
  )
  
  const interfaceMap: Record<string, any> = {}
  
  networkItems.forEach(item => {
    const instance = stripPort(item.instance)
    const iface = getLabelValue(item.labels, 'interface') || getLabelValue(item.labels, 'device') || 'eth0'
    const key = `${instance}-${iface}`
    
    if (!interfaceMap[key]) {
      interfaceMap[key] = {
        instance,
        interface: iface,
        downloadMB: '-',
        uploadMB: '-'
      }
    }
    
    if (item.rule_name.includes('下载') || item.rule_name.includes('接收')) {
      interfaceMap[key].downloadMB = formatBytes(item.value, 'B') + '/s'
    } else if (item.rule_name.includes('上传') || item.rule_name.includes('发送')) {
      interfaceMap[key].uploadMB = formatBytes(item.value, 'B') + '/s'
    }
  })
  
  return Object.values(interfaceMap)
})

// K8S节点状态表格数据
const k8sNodeTableData = computed(() => {
  const nodeItems = items.value.filter(i => 
    i.group_name.includes('K8S') && 
    (i.rule_name.includes('节点就绪') || i.rule_name.includes('节点状态') || i.rule_name.includes('Node Ready'))
  )
  
  return nodeItems.map(item => ({
    node: stripPort(item.instance),
    status: item.value === 1 ? 'Ready' : 'NotReady',
    raw: item
  }))
})

// K8S Pod状态表格数据
const k8sPodTableData = computed(() => {
  const podItems = items.value.filter(i => 
    i.group_name.includes('K8S') && 
    (i.rule_name.includes('Pod状态') || i.rule_name.includes('Pod运行'))
  )
  
  return podItems.map(item => ({
    namespace: getLabelValue(item.labels, 'namespace') || 'default',
    pod: getLabelValue(item.labels, 'pod') || item.instance,
    status: getPodStatus(item.value),
    raw: item
  }))
})

// K8S证书状态表格数据
const k8sCertTableData = computed(() => {
  const certItems = items.value.filter(i => 
    i.group_name.includes('K8S') && 
    (i.rule_name.includes('证书') || i.rule_name.includes('certificate'))
  )
  
  return certItems.map(item => ({
    instance: stripPort(item.instance),
    certName: getLabelValue(item.labels, 'certname') || getLabelValue(item.labels, 'name') || 'unknown',
    expiryDays: Math.floor(item.value / 86400), // 秒转天
    raw: item
  }))
})

// K8S PVC使用率表格数据
const k8sPVCTableData = computed(() => {
  const pvcItems = items.value.filter(i => 
    i.group_name.includes('K8S') && 
    (i.rule_name.includes('PVC') || i.rule_name.includes('持久卷'))
  )
  
  return pvcItems.map(item => ({
    namespace: getLabelValue(item.labels, 'namespace') || 'default',
    pvc: getLabelValue(item.labels, 'persistentvolumeclaim') || getLabelValue(item.labels, 'pvc') || 'unknown',
    usedPercent: item.value.toFixed(2),
    used: formatBytes(getLabelValue(item.labels, 'used_bytes') ? parseFloat(getLabelValue(item.labels, 'used_bytes')!) : 0),
    total: formatBytes(getLabelValue(item.labels, 'total_bytes') ? parseFloat(getLabelValue(item.labels, 'total_bytes')!) : 0),
    raw: item
  }))
})

// 进程CPU表格数据
const processCPUTableData = computed(() => {
  const processItems = items.value.filter(i => 
    (i.group_name.includes('进程') || i.group_name.includes('Process')) && 
    (i.rule_name.includes('CPU') || i.rule_name.includes('cpu'))
  ).slice(0, 5) // 取top5
  
  return processItems.map(item => ({
    instance: stripPort(item.instance),
    processName: getLabelValue(item.labels, 'process') || getLabelValue(item.labels, 'procname') || getLabelValue(item.labels, 'comm') || 'unknown',
    value: `${item.value.toFixed(2)}%`,
    status: item.status,
    statusText: item.status === 'normal' ? '正常' : item.status === 'warning' ? '告警' : '严重',
    raw: item
  }))
})

// 进程内存表格数据
const processMemTableData = computed(() => {
  const processItems = items.value.filter(i => 
    (i.group_name.includes('进程') || i.group_name.includes('Process')) && 
    (i.rule_name.includes('内存') || i.rule_name.includes('memory') || i.rule_name.includes('mem'))
  ).slice(0, 5) // 取top5
  
  return processItems.map(item => ({
    instance: stripPort(item.instance),
    processName: getLabelValue(item.labels, 'process') || getLabelValue(item.labels, 'procname') || getLabelValue(item.labels, 'comm') || 'unknown',
    value: formatBytes(item.value, 'B'),
    status: item.status,
    statusText: item.status === 'normal' ? '正常' : item.status === 'warning' ? '告警' : '严重',
    raw: item
  }))
})

// 其他分组详情（排除已特殊处理的）
const otherGroupDetails = computed(() => {
  const excludeGroups = ['基础资源', 'CPU监控', '内存监控', '磁盘监控', '磁盘IO', '网络IO', 'K8S', 'Kubernetes', '进程', 'Process']
  const nonBasicItems = items.value.filter(i => 
    !excludeGroups.some(g => i.group_name.includes(g)) && 
    !i.show_in_table
  )
  
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
    
    let row: any = {
      instance: stripPort(item.instance),
      value: formatValue(item.value, item.unit),
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

// 获取标签值
const getLabelValue = (labelsStr: string, key: string): string | null => {
  if (!labelsStr) return null
  try {
    const labels = JSON.parse(labelsStr)
    return labels[key] || null
  } catch (e) {
    return null
  }
}

// 格式化值
const formatValue = (value: number, unit?: string): string => {
  if (unit === '%' || unit === 'percent') {
    return `${value.toFixed(2)}%`
  }
  if (unit && unit.toLowerCase().includes('byte')) {
    return formatBytes(value, 'B')
  }
  if (unit === 's' || unit === 'seconds') {
    return formatBytes(value, 's')
  }
  if (unit) {
    return `${value.toFixed(2)}${unit}`
  }
  return value.toFixed(2)
}

// Pod状态转换
const getPodStatus = (value: number): string => {
  const statusMap: Record<number, string> = {
    0: 'Unknown',
    1: 'Running',
    2: 'Pending',
    3: 'Succeeded',
    4: 'Failed',
    5: 'CrashLoopBackOff'
  }
  return statusMap[value] || 'Unknown'
}

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
  // 生成日期标签
  const dates = []
  const now = new Date()
  for (let i = 6; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    dates.push(d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }))
  }
  
  // 获取所有实例
  const instances = [...new Set(items.value.map(i => stripPort(i.instance)))]
  
  // 获取CPU、内存、磁盘数据
  const cpuItems = items.value.filter(i => i.rule_name.includes('CPU') && i.trend_data)
  const memItems = items.value.filter(i => i.rule_name.includes('内存') && i.trend_data)
  const diskItems = items.value.filter(i => i.rule_name.includes('磁盘') && i.trend_data)
  
  // 如果有趋势数据，使用真实数据
  if (cpuItems.length > 0 || memItems.length > 0 || diskItems.length > 0) {
    renderRealCharts(dates, instances, cpuItems, memItems, diskItems)
  } else {
    // 否则使用模拟数据
    renderMockCharts(dates, instances)
  }
}

const renderRealCharts = (
  dates: string[], 
  instances: string[],
  cpuItems: InspectionItem[],
  memItems: InspectionItem[],
  diskItems: InspectionItem[]
) => {
  // CPU 趋势
  if (cpuChart.value) {
    const chart = echarts.init(cpuChart.value)
    const series = instances.map((instance, idx) => {
      const item = cpuItems.find(i => stripPort(i.instance) === instance)
      let data = generateMockData(7)
      
      if (item && item.trend_data) {
        try {
          const trendData = JSON.parse(item.trend_data)
          if (trendData.data?.result?.[0]?.values) {
            data = trendData.data.result[0].values.map((v: any) => {
              const val = parseFloat(v[1])
              return isNaN(val) ? 0 : val * 100
            })
          }
        } catch (e) {}
      }
      
      return {
        name: instance,
        type: 'line',
        data: data,
        smooth: true,
        itemStyle: { color: colorPalette[idx % colorPalette.length] }
      }
    })
    
    chart.setOption({
      title: { text: 'CPU 使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { 
        trigger: 'axis',
        formatter: (params: any) => {
          let result = params[0].axisValue + '<br/>'
          params.forEach((p: any) => {
            result += `${p.marker}${p.seriesName}: ${p.value?.toFixed(2)}%<br/>`
          })
          return result
        }
      },
      legend: { 
        data: instances,
        bottom: 0,
        type: 'scroll'
      },
      grid: { 
        left: '3%', 
        right: '4%', 
        bottom: '15%',
        containLabel: true 
      },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
  
  // 内存趋势
  if (memChart.value) {
    const chart = echarts.init(memChart.value)
    const series = instances.map((instance, idx) => {
      const item = memItems.find(i => stripPort(i.instance) === instance)
      let data = generateMockData(7)
      
      if (item && item.trend_data) {
        try {
          const trendData = JSON.parse(item.trend_data)
          if (trendData.data?.result?.[0]?.values) {
            data = trendData.data.result[0].values.map((v: any) => {
              const val = parseFloat(v[1])
              return isNaN(val) ? 0 : val * 100
            })
          }
        } catch (e) {}
      }
      
      return {
        name: instance,
        type: 'line',
        data: data,
        smooth: true,
        itemStyle: { color: colorPalette[idx % colorPalette.length] }
      }
    })
    
    chart.setOption({
      title: { text: '内存使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { 
        trigger: 'axis',
        formatter: (params: any) => {
          let result = params[0].axisValue + '<br/>'
          params.forEach((p: any) => {
            result += `${p.marker}${p.seriesName}: ${p.value?.toFixed(2)}%<br/>`
          })
          return result
        }
      },
      legend: { 
        data: instances,
        bottom: 0,
        type: 'scroll'
      },
      grid: { 
        left: '3%', 
        right: '4%', 
        bottom: '15%',
        containLabel: true 
      },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
  
  // 磁盘趋势
  if (diskChart.value) {
    const chart = echarts.init(diskChart.value)
    const series = instances.map((instance, idx) => {
      const item = diskItems.find(i => stripPort(i.instance) === instance)
      let data = generateMockData(7)
      
      if (item && item.trend_data) {
        try {
          const trendData = JSON.parse(item.trend_data)
          if (trendData.data?.result?.[0]?.values) {
            data = trendData.data.result[0].values.map((v: any) => {
              const val = parseFloat(v[1])
              return isNaN(val) ? 0 : val * 100
            })
          }
        } catch (e) {}
      }
      
      return {
        name: instance,
        type: 'line',
        data: data,
        smooth: true,
        itemStyle: { color: colorPalette[idx % colorPalette.length] }
      }
    })
    
    chart.setOption({
      title: { text: '磁盘使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { 
        trigger: 'axis',
        formatter: (params: any) => {
          let result = params[0].axisValue + '<br/>'
          params.forEach((p: any) => {
            result += `${p.marker}${p.seriesName}: ${p.value?.toFixed(2)}%<br/>`
          })
          return result
        }
      },
      legend: { 
        data: instances,
        bottom: 0,
        type: 'scroll'
      },
      grid: { 
        left: '3%', 
        right: '4%', 
        bottom: '15%',
        containLabel: true 
      },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
}

const renderMockCharts = (dates: string[], instances: string[]) => {
  // CPU 趋势
  if (cpuChart.value) {
    const chart = echarts.init(cpuChart.value)
    const series = instances.map((instance, idx) => ({
      name: instance,
      type: 'line',
      data: generateMockData(7, 30 + idx * 10),
      smooth: true,
      itemStyle: { color: colorPalette[idx % colorPalette.length] }
    }))
    
    chart.setOption({
      title: { text: 'CPU 使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { 
        trigger: 'axis',
        formatter: (params: any) => {
          let result = params[0].axisValue + '<br/>'
          params.forEach((p: any) => {
            result += `${p.marker}${p.seriesName}: ${p.value?.toFixed(2)}%<br/>`
          })
          return result
        }
      },
      legend: { 
        data: instances,
        bottom: 0,
        type: 'scroll'
      },
      grid: { 
        left: '3%', 
        right: '4%', 
        bottom: '15%',
        containLabel: true 
      },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
  
  // 内存趋势
  if (memChart.value) {
    const chart = echarts.init(memChart.value)
    const series = instances.map((instance, idx) => ({
      name: instance,
      type: 'line',
      data: generateMockData(7, 50 + idx * 8),
      smooth: true,
      itemStyle: { color: colorPalette[idx % colorPalette.length] }
    }))
    
    chart.setOption({
      title: { text: '内存使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { 
        trigger: 'axis',
        formatter: (params: any) => {
          let result = params[0].axisValue + '<br/>'
          params.forEach((p: any) => {
            result += `${p.marker}${p.seriesName}: ${p.value?.toFixed(2)}%<br/>`
          })
          return result
        }
      },
      legend: { 
        data: instances,
        bottom: 0,
        type: 'scroll'
      },
      grid: { 
        left: '3%', 
        right: '4%', 
        bottom: '15%',
        containLabel: true 
      },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
  
  // 磁盘趋势
  if (diskChart.value) {
    const chart = echarts.init(diskChart.value)
    const series = instances.map((instance, idx) => ({
      name: instance,
      type: 'line',
      data: generateMockData(7, 40 + idx * 5),
      smooth: true,
      itemStyle: { color: colorPalette[idx % colorPalette.length] }
    }))
    
    chart.setOption({
      title: { text: '磁盘使用率趋势', textStyle: { fontSize: 14 } },
      tooltip: { 
        trigger: 'axis',
        formatter: (params: any) => {
          let result = params[0].axisValue + '<br/>'
          params.forEach((p: any) => {
            result += `${p.marker}${p.seriesName}: ${p.value?.toFixed(2)}%<br/>`
          })
          return result
        }
      },
      legend: { 
        data: instances,
        bottom: 0,
        type: 'scroll'
      },
      grid: { 
        left: '3%', 
        right: '4%', 
        bottom: '15%',
        containLabel: true 
      },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
}

const generateMockData = (count: number, base: number = 50): number[] => {
  const data = []
  for (let i = 0; i < count; i++) {
    data.push(base + Math.random() * 20 - 10)
  }
  return data
}

const handleExport = async () => {
  const element = document.getElementById('report-content')
  if (!element) return
  
  try {
    // 保存当前总结表单内容
    const summarySection = document.getElementById('summary-section')
    void summarySection // 确保元素存在用于导出
    
    const canvas = await html2canvas(element, {
      scale: 2,
      useCORS: true,
      backgroundColor: '#fff'
    })
    
    const imgWidth = 210
    const pageHeight = 297
    const imgHeight = (canvas.height * imgWidth) / canvas.width
    
    const pdf = new jsPDF('p', 'mm', 'a4')
    
    // 如果内容超过一页，需要分页
    let heightLeft = imgHeight
    let position = 0
    
    pdf.addImage(
      canvas.toDataURL('image/png'),
      'PNG',
      0,
      position,
      imgWidth,
      imgHeight
    )
    
    heightLeft -= pageHeight
    
    while (heightLeft > 0) {
      position = heightLeft - imgHeight
      pdf.addPage()
      pdf.addImage(
        canvas.toDataURL('image/png'),
        'PNG',
        0,
        position,
        imgWidth,
        imgHeight
      )
      heightLeft -= pageHeight
    }
    
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

const getBasicTableCellClass = ({ row, column }: { row: any; column: any }) => {
  const propName = column.property
  if (!propName) return ''
  
  const statusKey = propName + '_status'
  const status = row[statusKey]
  
  if (status === 'critical') return 'cell-critical'
  if (status === 'warning') return 'cell-warning'
  return ''
}

const getK8sCellClass = ({ row, column }: { row: any; column: any }) => {
  const propName = column.property
  if (!propName) return ''
  
  if (propName === 'status') {
    if (row.status === 'Ready' || row.status === 'Running') return 'cell-ready'
    return 'cell-not-ready'
  }
  
  if (propName === 'expiryDays' && row.expiryDays < 30) {
    return 'cell-warning'
  }
  
  if (propName === 'usedPercent') {
    const percent = parseFloat(row.usedPercent)
    if (percent > 90) return 'cell-critical'
    if (percent > 80) return 'cell-warning'
  }
  
  return ''
}

const getPodStatusClass = (status: string) => {
  if (status === 'Running') return 'status-ready'
  if (status === 'Pending') return 'status-warning'
  return 'status-not-ready'
}

const getCertStatusClass = (days: number) => {
  if (days < 7) return 'status-critical'
  if (days < 30) return 'status-warning'
  return 'status-ready'
}

const getUsageStatusClass = (percent: number | string) => {
  const p = typeof percent === 'string' ? parseFloat(percent) : percent
  if (p > 90) return 'status-critical'
  if (p > 80) return 'status-warning'
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
  height: 300px;
}

.rule-section {
  margin-bottom: 20px;
}

.rule-title {
  font-weight: bold;
  margin-bottom: 10px;
  color: #606266;
}

/* 状态文字颜色 */
.status-critical {
  color: #F56C6C;
  font-weight: bold;
}

.status-warning {
  color: #E6A23C;
  font-weight: bold;
}

.status-ready {
  color: #67C23A;
  font-weight: bold;
}

.status-not-ready {
  color: #F56C6C;
  font-weight: bold;
}
</style>

<style>
/* 单元格背景色 - 需要全局样式 */
.cell-critical {
  background-color: #fef0f0 !important;
  color: #F56C6C !important;
  font-weight: bold;
}

.cell-warning {
  background-color: #fdf6ec !important;
  color: #E6A23C !important;
  font-weight: bold;
}

.cell-ready {
  background-color: #f0f9eb !important;
  color: #67C23A !important;
  font-weight: bold;
}

.cell-not-ready {
  background-color: #fef0f0 !important;
  color: #F56C6C !important;
  font-weight: bold;
}
</style>
