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
      <el-card class="section-card" v-if="basicResourceData.length > 0">
        <template #header>
          <span class="section-title">基础资源详情</span>
        </template>
        <el-table :data="basicResourceData" stripe border size="small" :span-method="basicResourceSpanMethod" :cell-class-name="getBasicTableCellClass">
          <el-table-column prop="ip" label="IP地址" width="130" fixed="left" />
          <el-table-column prop="cpuCores" label="CPU核心数" width="100" />
          <el-table-column prop="cpuUsage" label="CPU使用率" width="110" />
          <el-table-column prop="uptime" label="运行时间" width="130" />
          <el-table-column prop="load5" label="5分钟负载" width="100" />
          <el-table-column prop="memTotal" label="内存总量" width="100" />
          <el-table-column prop="memUsed" label="内存使用量" width="110" />
          <el-table-column prop="memUsage" label="内存使用率" width="110" />
          <el-table-column prop="mountpoint" label="挂载点" width="100" />
          <el-table-column prop="diskTotal" label="磁盘总量" width="100" />
          <el-table-column prop="diskUsed" label="磁盘使用量" width="110" />
          <el-table-column prop="diskUsage" label="磁盘使用率" width="110" />
          <el-table-column prop="tcpConn" label="TCP连接数" width="100" />
          <el-table-column prop="tcpTw" label="TCP_TW数" width="100" />
        </el-table>
      </el-card>

      <!-- K8S节点就绪状态 -->
      <el-card class="section-card" v-if="k8sNodeTableData.length > 0">
        <template #header>
          <span class="section-title">K8S节点就绪状态</span>
        </template>
        <el-table :data="k8sNodeTableData" stripe border size="small">
          <el-table-column prop="node" label="节点" width="150" />
          <el-table-column prop="statusType" label="状态类型" width="120">
            <template #default>Ready</template>
          </el-table-column>
          <el-table-column prop="value" label="值" width="100" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.value === 0 ? 'status-normal' : 'status-critical'">
                {{ row.value === 0 ? '正常' : '异常' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
        <div class="table-note">说明：值=0表示正常，值≠0表示异常</div>
      </el-card>

      <!-- K8S Pod运行状态 -->
      <el-card class="section-card" v-if="k8sPodTableData.length > 0">
        <template #header>
          <span class="section-title">K8S Pod运行状态</span>
        </template>
        <el-table :data="k8sPodTableData" stripe border size="small">
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column prop="pod" label="Pod名" min-width="300" />
          <el-table-column prop="value" label="运行状态" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.value === 1 ? 'status-normal' : 'status-critical'">
                {{ row.value === 1 ? '正常' : '异常' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
        <div class="table-note">说明：值=1表示正常，值≠1表示异常</div>
      </el-card>

      <!-- K8S PVC使用率 -->
      <el-card class="section-card" v-if="k8sPVCTableData.length > 0">
        <template #header>
          <span class="section-title">K8S PVC使用率</span>
        </template>
        <el-table :data="k8sPVCTableData" stripe border size="small">
          <el-table-column prop="pvc" label="PVC名称" min-width="300" />
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column prop="usedPercent" label="PVC使用率" width="120">
            <template #default="{ row }">
              <span :class="row.usedPercent >= 90 ? 'status-critical' : 'status-normal'">
                {{ row.usedPercent.toFixed(2) }}%
              </span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.usedPercent >= 90 ? 'status-critical' : 'status-normal'">
                {{ row.usedPercent >= 90 ? '异常' : '正常' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
        <div class="table-note">说明：值>=90%表示异常，值<90%表示正常</div>
      </el-card>

      <!-- K8S证书状态 - 按证书类型分组 -->
      <template v-for="(certGroup, certType) in k8sCertGroupedData" :key="certType">
        <el-card class="section-card" v-if="certGroup.length > 0">
          <template #header>
            <span class="section-title">{{ certType }}</span>
          </template>
          <el-table :data="certGroup" stripe border size="small">
            <el-table-column prop="node" label="节点" width="150" />
            <el-table-column prop="value" label="值" width="150">
              <template #default="{ row }">
                {{ row.value }} 天
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <span :class="row.value >= 30 ? 'status-normal' : 'status-warning'">
                  {{ row.value >= 30 ? '正常' : '异常' }}
                </span>
              </template>
            </el-table-column>
          </el-table>
          <div class="table-note">说明：值>=30天表示正常，值<30天表示异常</div>
        </el-card>
      </template>

      <!-- 30分钟内磁盘平均写入值 -->
      <el-card class="section-card" v-if="diskWriteTableData.length > 0">
        <template #header>
          <span class="section-title">30分钟内磁盘平均写入值</span>
        </template>
        <el-table :data="diskWriteTableData" stripe border size="small">
          <el-table-column prop="instance" label="节点" width="150" />
          <el-table-column prop="device" label="设备" width="150" />
          <el-table-column prop="valueFormatted" label="值" width="150" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.status === 'normal' ? 'status-normal' : 'status-warning'">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 30分钟内磁盘平均读取值 -->
      <el-card class="section-card" v-if="diskReadTableData.length > 0">
        <template #header>
          <span class="section-title">30分钟内磁盘平均读取值</span>
        </template>
        <el-table :data="diskReadTableData" stripe border size="small">
          <el-table-column prop="instance" label="节点" width="150" />
          <el-table-column prop="device" label="设备" width="150" />
          <el-table-column prop="valueFormatted" label="值" width="150" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.status === 'normal' ? 'status-normal' : 'status-warning'">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 30分钟内上传速率 -->
      <el-card class="section-card" v-if="networkUploadTableData.length > 0">
        <template #header>
          <span class="section-title">30分钟内上传速率</span>
        </template>
        <el-table :data="networkUploadTableData" stripe border size="small">
          <el-table-column prop="instance" label="节点" width="150" />
          <el-table-column prop="device" label="设备" width="150" />
          <el-table-column prop="valueFormatted" label="值" width="150" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.status === 'normal' ? 'status-normal' : 'status-warning'">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 30分钟内下载速率 -->
      <el-card class="section-card" v-if="networkDownloadTableData.length > 0">
        <template #header>
          <span class="section-title">30分钟内下载速率</span>
        </template>
        <el-table :data="networkDownloadTableData" stripe border size="small">
          <el-table-column prop="instance" label="节点" width="150" />
          <el-table-column prop="device" label="设备" width="150" />
          <el-table-column prop="valueFormatted" label="值" width="150" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.status === 'normal' ? 'status-normal' : 'status-warning'">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 进程CPU使用率Top5 -->
      <el-card class="section-card" v-if="processCPUTableData.length > 0">
        <template #header>
          <span class="section-title">进程CPU使用率Top5</span>
        </template>
        <el-table :data="processCPUTableData" stripe border size="small">
          <el-table-column prop="processName" label="进程名" width="200" />
          <el-table-column prop="instance" label="所在机器" width="180" />
          <el-table-column prop="value" label="值" width="120" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.status === 'normal' ? 'status-normal' : 'status-warning'">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 进程内存使用率Top5 -->
      <el-card class="section-card" v-if="processMemTableData.length > 0">
        <template #header>
          <span class="section-title">进程内存使用率Top5</span>
        </template>
        <el-table :data="processMemTableData" stripe border size="small">
          <el-table-column prop="processName" label="进程名" width="200" />
          <el-table-column prop="instance" label="所在机器" width="180" />
          <el-table-column prop="value" label="值" width="120" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <span :class="row.status === 'normal' ? 'status-normal' : 'status-warning'">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 其他分组详情 -->
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

// 预定义的颜色列表
const colorPalette = [
  '#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de',
  '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc', '#48b8d0'
]

// 去掉IP端口
const stripPort = (instance: string): string => {
  if (!instance) return instance
  return instance.split(':')[0]
}

// 格式化字节大小为人类可读格式
const formatBytesToHuman = (bytes: number): string => {
  if (!bytes || isNaN(bytes)) return '-'
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const k = 1024
  const i = Math.floor(Math.log(Math.abs(bytes)) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + units[i]
}

// 格式化运行时间为"X天Y小时"格式
const formatUptime = (seconds: number): string => {
  if (!seconds || isNaN(seconds)) return '-'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  
  let result = ''
  if (days > 0) result += `${days}天`
  if (hours > 0 || days > 0) result += `${hours}小时`
  if (minutes > 0 && days === 0) result += `${minutes}分钟`
  
  return result || '<1分钟'
}

// 解析labels JSON
const parseLabels = (labelsStr: string): Record<string, string> => {
  if (!labelsStr) return {}
  try {
    return JSON.parse(labelsStr)
  } catch (e) {
    return {}
  }
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

// ==================== 基础资源详情表格 ====================

// 判断是否为基础资源分组
const isBasicResourceGroup = (groupName: string): boolean => {
  const lower = groupName.toLowerCase()
  return lower.includes('基础资源') || 
         lower.includes('cpu') || 
         lower.includes('内存') || 
         lower.includes('磁盘') ||
         lower.includes('disk') ||
         lower.includes('memory') ||
         lower.includes('基础')
}

// 基础资源详情表格数据 - 支持单元格合并
// 非磁盘数据合并为一行，磁盘数据按挂载点分行
const basicResourceData = computed(() => {
  // 只取基础资源分组且show_in_table的数据
  const basicItems = items.value.filter(i => isBasicResourceGroup(i.group_name) && i.show_in_table)
  
  // 先按IP分组，收集每个IP的所有数据
  const ipDataMap: Record<string, Record<string, { value: number; status: string; labels: Record<string, string> }>> = {}
  const ipMountpoints: Record<string, Set<string>> = {}
  
  basicItems.forEach(item => {
    const ip = stripPort(item.instance)
    const labels = parseLabels(item.labels)
    const mountpoint = labels.mountpoint || labels.device || ''
    
    if (!ipDataMap[ip]) {
      ipDataMap[ip] = {}
      ipMountpoints[ip] = new Set()
    }
    
    // 记录该IP的挂载点
    if (mountpoint && !mountpoint.includes('pod') && !mountpoint.startsWith('/run') && !mountpoint.startsWith('/boot')) {
      ipMountpoints[ip].add(mountpoint)
    }
    
    // 存储数据，用 规则名_挂载点 作为key来区分不同挂载点的磁盘数据
    const dataKey = mountpoint ? `${item.rule_name}_${mountpoint}` : item.rule_name
    ipDataMap[ip][dataKey] = {
      value: item.value,
      status: item.status,
      labels
    }
    
    // 也存储一份不带挂载点的数据（用于非磁盘指标）
    if (!ipDataMap[ip][item.rule_name]) {
      ipDataMap[ip][item.rule_name] = {
        value: item.value,
        status: item.status,
        labels
      }
    }
  })
  
  // 生成表格行数据 - 每个挂载点一行，非磁盘数据需要在第一行显示
  const result: any[] = []
  
  // 记录每个IP的起始行索引和行数，用于单元格合并
  const ipRowInfo: Record<string, { startIndex: number; rowCount: number }> = {}
  let currentIndex = 0
  
  Object.keys(ipDataMap).sort().forEach(ip => {
    const data = ipDataMap[ip]
    const mountpoints = Array.from(ipMountpoints[ip]).sort()
    
    // 记录该IP的起始行
    ipRowInfo[ip] = { startIndex: currentIndex, rowCount: 0 }
    
    if (mountpoints.length === 0) {
      // 没有挂载点，添加一行
      result.push(createBasicRow(ip, data, '', true))
      ipRowInfo[ip].rowCount = 1
      currentIndex++
    } else {
      // 每个挂载点一行
      mountpoints.forEach((mountpoint, idx) => {
        result.push(createBasicRow(ip, data, mountpoint, idx === 0))
        ipRowInfo[ip].rowCount++
        currentIndex++
      })
    }
  })
  
  // 将合并信息附加到每行数据
  result.forEach((row, idx) => {
    const ip = row._ip
    if (ip && ipRowInfo[ip]) {
      const info = ipRowInfo[ip]
      row._rowIndex = idx
      row._rowSpan = idx === info.startIndex ? info.rowCount : 0
    }
  })
  
  return result
})

// 创建基础资源行数据
const createBasicRow = (ip: string, data: Record<string, any>, mountpoint: string, isFirst: boolean = true) => {
  const getData = (keys: string[]): { value: number; status: string } | undefined => {
    for (const key of keys) {
      // 先尝试带挂载点的key
      const mpKey = `${key}_${mountpoint}`
      if (data[mpKey]) return data[mpKey]
      if (data[key]) return data[key]
    }
    return undefined
  }
  
  // CPU相关
  const cpuCores = getData(['CPU核心数', 'cpu_cores', 'CPU核心', 'cpu核心'])
  const cpuUsage = getData(['CPU使用率', 'cpu使用率', 'CPU Usage'])
  const uptime = getData(['运行时间', 'uptime', '系统运行时间'])
  const load5 = getData(['5分钟负载', '负载5', 'load5', '系统负载'])
  
  // 内存相关
  const memTotal = getData(['内存总量', 'memory_total', 'Memory Total', '节点内存'])
  const memUsed = getData(['内存使用量', 'memory_used', 'Memory Used'])
  const memUsage = getData(['内存使用率', 'memory_usage', 'Memory Usage'])
  
  // 磁盘相关
  const diskTotal = getData(['磁盘总量', 'disk_total', 'Disk Total'])
  const diskUsed = getData(['磁盘使用量', 'disk_used', 'Disk Used'])
  const diskUsage = getData(['磁盘使用率', 'disk_usage', 'Disk Usage'])
  
  // 网络相关
  const tcpConn = getData(['TCP连接数', 'tcp_conn', 'TCP Connections'])
  const tcpTw = getData(['TCP_TW数', 'tcp_tw', 'TCP TimeWait'])
  
  return {
    _ip: ip,
    _isFirst: isFirst,
    ip: isFirst ? ip : '',
    mountpoint: mountpoint || '-',
    cpuCores: isFirst ? (cpuCores ? Math.round(cpuCores.value) : '-') : '',
    cpuUsage: isFirst ? (cpuUsage ? cpuUsage.value.toFixed(2) + '%' : '-') : '',
    cpuUsageStatus: isFirst ? (cpuUsage?.status || '') : '',
    uptime: isFirst ? (uptime ? formatUptime(uptime.value) : '-') : '',
    load5: isFirst ? (load5 ? load5.value.toFixed(2) : '-') : '',
    load5Status: isFirst ? (load5?.status || '') : '',
    memTotal: isFirst ? (memTotal ? formatBytesToHuman(memTotal.value) : '-') : '',
    memUsed: isFirst ? (memUsed ? formatBytesToHuman(memUsed.value) : '-') : '',
    memUsage: isFirst ? (memUsage ? memUsage.value.toFixed(2) + '%' : '-') : '',
    memUsageStatus: isFirst ? (memUsage?.status || '') : '',
    diskTotal: diskTotal ? formatBytesToHuman(diskTotal.value) : '-',
    diskUsed: diskUsed ? formatBytesToHuman(diskUsed.value) : '-',
    diskUsage: diskUsage ? diskUsage.value.toFixed(2) + '%' : '-',
    diskUsageStatus: diskUsage?.status || '',
    tcpConn: isFirst ? (tcpConn ? Math.round(tcpConn.value) : '-') : '',
    tcpTw: isFirst ? (tcpTw ? Math.round(tcpTw.value) : '-') : ''
  }
}

// 基础资源表格单元格合并方法
// 非磁盘列（IP到内存使用率、TCP连接数、TCP_TW数）需要合并
// 磁盘列（挂载点、磁盘总量、磁盘使用量、磁盘使用率）不合并
const basicResourceSpanMethod = ({ row, column, rowIndex, columnIndex }: { row: any; column: any; rowIndex: number; columnIndex: number }) => {
  // 需要合并的列索引：0-7（IP到内存使用率），12-13（TCP连接数、TCP_TW数）
  // 不需要合并的列索引：8-11（挂载点、磁盘总量、磁盘使用量、磁盘使用率）
  const mergeColumns = [0, 1, 2, 3, 4, 5, 6, 7, 12, 13]
  
  if (mergeColumns.includes(columnIndex)) {
    if (row._rowSpan > 0) {
      return {
        rowspan: row._rowSpan,
        colspan: 1
      }
    } else {
      return {
        rowspan: 0,
        colspan: 0
      }
    }
  }
}

// 基础资源表格单元格样式
const getBasicTableCellClass = ({ row, column }: { row: any; column: any }) => {
  const propName = column.property
  if (!propName) return ''
  const statusKey = propName + 'Status'
  const status = row[statusKey]
  if (status === 'critical') return 'cell-critical'
  if (status === 'warning') return 'cell-warning'
  return ''
}

// ==================== K8S相关表格 ====================

// 判断是否为K8S分组
const isK8SGroup = (groupName: string): boolean => {
  const lower = groupName.toLowerCase()
  return lower.includes('k8s') || lower.includes('kubernetes') || lower.includes('容器') || lower.includes('pod') || lower.includes('container')
}

// K8S节点就绪状态表格数据
// 值=0表示正常，值≠0表示异常
const k8sNodeTableData = computed(() => {
  const nodeItems = items.value.filter(i => 
    isK8SGroup(i.group_name) && 
    (i.rule_name.includes('节点就绪') || i.rule_name.includes('节点状态') || 
     i.rule_name.toLowerCase().includes('node ready') || i.rule_name.toLowerCase().includes('node status'))
  )
  
  // 节点去重
  const nodeMap: Record<string, { node: string; statusType: string; value: number }> = {}
  
  nodeItems.forEach(item => {
    const labels = parseLabels(item.labels)
    const node = labels.node || stripPort(item.instance)
    
    if (!nodeMap[node]) {
      nodeMap[node] = {
        node,
        statusType: 'Ready',
        value: item.value
      }
    }
  })
  
  return Object.values(nodeMap)
})

// K8S Pod运行状态表格数据
// 值=1表示正常，值≠1表示异常
const k8sPodTableData = computed(() => {
  const podItems = items.value.filter(i => 
    isK8SGroup(i.group_name) && 
    (i.rule_name.includes('Pod状态') || i.rule_name.includes('Pod运行') || 
     i.rule_name.toLowerCase().includes('pod status'))
  )
  
  return podItems.map(item => {
    const labels = parseLabels(item.labels)
    return {
      namespace: labels.namespace || 'default',
      pod: labels.pod || stripPort(item.instance),
      value: item.value
    }
  })
})

// K8S PVC使用率表格数据
// 列顺序：PVC名称、命名空间、PVC使用率、状态
// 值>=90%表示异常
const k8sPVCTableData = computed(() => {
  const pvcItems = items.value.filter(i => 
    isK8SGroup(i.group_name) && 
    (i.rule_name.includes('PVC') || i.rule_name.includes('持久卷') || i.rule_name.toLowerCase().includes('pvc') || i.rule_name.includes('存储卷'))
  )
  
  return pvcItems.map(item => {
    const labels = parseLabels(item.labels)
    return {
      pvc: labels.persistentvolumeclaim || 'unknown',
      namespace: labels.namespace || 'default',
      usedPercent: item.value
    }
  })
})

// K8S证书状态表格数据 - 按证书类型分组
const k8sCertGroupedData = computed(() => {
  const certItems = items.value.filter(i => 
    isK8SGroup(i.group_name) && 
    (i.rule_name.includes('证书') || i.rule_name.toLowerCase().includes('certificate'))
  )
  
  const grouped: Record<string, Array<{ node: string; value: number }>> = {}
  
  certItems.forEach(item => {
    const labels = parseLabels(item.labels)
    const node = labels.node || labels.instance || stripPort(item.instance)
    const value = Math.floor(item.value)
    
    // 从规则名获取证书类型
    let certType = item.rule_name
    if (item.rule_name.includes('Kubelet')) {
      certType = 'Kubelet证书状态'
    } else if (item.rule_name.includes('Kubeproxy')) {
      certType = 'Kubeproxy证书状态'
    } else if (item.rule_name.includes('Kubecontroller') || item.rule_name.includes('Controller')) {
      certType = 'Kubecontroller证书状态'
    }
    
    if (!grouped[certType]) {
      grouped[certType] = []
    }
    
    grouped[certType].push({ node, value })
  })
  
  return grouped
})

// ==================== 进程相关表格 ====================

// 判断是否为进程分组
const isProcessGroup = (groupName: string): boolean => {
  const lower = groupName.toLowerCase()
  return lower.includes('进程') || lower.includes('process')
}

// 进程CPU表格数据 - 每台机器的top5（后端查询已按instance分组）
const processCPUTableData = computed(() => {
  const processItems = items.value.filter(i => 
    isProcessGroup(i.group_name) && 
    (i.rule_name.toLowerCase().includes('cpu') || i.rule_name.includes('CPU'))
  )
  
  // 按instance分组，每组按值排序
  const groupedByInstance: Record<string, any[]> = {}
  processItems.forEach(item => {
    const instance = stripPort(item.instance)
    if (!groupedByInstance[instance]) {
      groupedByInstance[instance] = []
    }
    groupedByInstance[instance].push(item)
  })
  
  // 每个instance按值降序排序，然后合并所有结果
  const result: any[] = []
  Object.keys(groupedByInstance).sort().forEach(instance => {
    const items = groupedByInstance[instance].sort((a, b) => b.value - a.value)
    items.forEach(item => {
      const labels = parseLabels(item.labels)
      result.push({
        processName: labels.groupname || 'unknown',
        instance: instance,
        value: `${(item.value * 100).toFixed(2)}%`,
        status: item.status
      })
    })
  })
  
  return result
})

// 进程内存表格数据 - 每台机器的top5（后端查询已按instance分组）
const processMemTableData = computed(() => {
  const processItems = items.value.filter(i => 
    isProcessGroup(i.group_name) && 
    (i.rule_name.toLowerCase().includes('内存') || i.rule_name.includes('内存'))
  )
  
  // 按instance分组，每组按值排序
  const groupedByInstance: Record<string, any[]> = {}
  processItems.forEach(item => {
    const instance = stripPort(item.instance)
    if (!groupedByInstance[instance]) {
      groupedByInstance[instance] = []
    }
    groupedByInstance[instance].push(item)
  })
  
  // 每个instance按值降序排序，然后合并所有结果
  const result: any[] = []
  Object.keys(groupedByInstance).sort().forEach(instance => {
    const items = groupedByInstance[instance].sort((a, b) => b.value - a.value)
    items.forEach(item => {
      const labels = parseLabels(item.labels)
      result.push({
        processName: labels.groupname || 'unknown',
        instance: instance,
        value: formatBytesToHuman(item.value),
        status: item.status
      })
    })
  })
  
  return result
})

// ==================== 磁盘IO和网络IO相关表格 ====================

// 磁盘平均写入值表格数据
// 直接根据规则名称识别，不依赖分组名称
const diskWriteTableData = computed(() => {
  const diskWriteItems = items.value.filter(i => 
    i.rule_name.includes('磁盘平均写入') || 
    i.rule_name.includes('磁盘写入') ||
    i.rule_name.toLowerCase().includes('disk write') ||
    i.rule_name.toLowerCase().includes('diskwrite')
  )
  
  return diskWriteItems.map(item => {
    const labels = parseLabels(item.labels)
    return {
      instance: stripPort(item.instance),
      device: labels.device || '-',
      value: item.value,
      valueFormatted: `${item.value.toFixed(2)} MB/s`,
      status: item.status
    }
  })
})

// 磁盘平均读取值表格数据
const diskReadTableData = computed(() => {
  const diskReadItems = items.value.filter(i => 
    i.rule_name.includes('磁盘平均读取') || 
    i.rule_name.includes('磁盘读取') ||
    i.rule_name.toLowerCase().includes('disk read') ||
    i.rule_name.toLowerCase().includes('diskread')
  )
  
  return diskReadItems.map(item => {
    const labels = parseLabels(item.labels)
    return {
      instance: stripPort(item.instance),
      device: labels.device || '-',
      value: item.value,
      valueFormatted: `${item.value.toFixed(2)} MB/s`,
      status: item.status
    }
  })
})

// 上传速率表格数据
const networkUploadTableData = computed(() => {
  const uploadItems = items.value.filter(i => 
    i.rule_name.includes('上传速率') || 
    i.rule_name.includes('上传') ||
    i.rule_name.toLowerCase().includes('upload rate') ||
    i.rule_name.toLowerCase().includes('network transmit')
  )
  
  return uploadItems.map(item => {
    const labels = parseLabels(item.labels)
    return {
      instance: stripPort(item.instance),
      device: labels.device || labels.interface || labels.if || '-',
      value: item.value,
      valueFormatted: `${item.value.toFixed(2)} MB/s`,
      status: item.status
    }
  })
})

// 下载速率表格数据
const networkDownloadTableData = computed(() => {
  const downloadItems = items.value.filter(i => 
    i.rule_name.includes('下载速率') || 
    i.rule_name.includes('下载') ||
    i.rule_name.toLowerCase().includes('download rate') ||
    i.rule_name.toLowerCase().includes('network receive')
  )
  
  return downloadItems.map(item => {
    const labels = parseLabels(item.labels)
    return {
      instance: stripPort(item.instance),
      device: labels.device || labels.interface || labels.if || '-',
      value: item.value,
      valueFormatted: `${item.value.toFixed(2)} MB/s`,
      status: item.status
    }
  })
})

// ==================== 其他分组详情 ====================

// 判断是否为磁盘IO相关规则
const isDiskIORule = (ruleName: string): boolean => {
  const lower = ruleName.toLowerCase()
  return ruleName.includes('磁盘平均写入') || 
         ruleName.includes('磁盘平均读取') ||
         ruleName.includes('磁盘写入') ||
         ruleName.includes('磁盘读取') ||
         lower.includes('disk write') ||
         lower.includes('disk read') ||
         lower.includes('diskwrite') ||
         lower.includes('diskread')
}

// 判断是否为网络IO相关规则
const isNetworkIORule = (ruleName: string): boolean => {
  const lower = ruleName.toLowerCase()
  return ruleName.includes('上传速率') || 
         ruleName.includes('下载速率') ||
         ruleName.includes('上传') ||
         ruleName.includes('下载') ||
         lower.includes('upload rate') ||
         lower.includes('download rate') ||
         lower.includes('network transmit') ||
         lower.includes('network receive')
}

// 其他分组详情（排除已特殊处理的）
const otherGroupDetails = computed(() => {
  const nonBasicItems = items.value.filter(i => {
    if (i.show_in_table) return false
    if (isK8SGroup(i.group_name)) return false
    if (isProcessGroup(i.group_name)) return false
    if (isDiskIORule(i.rule_name)) return false
    if (isNetworkIORule(i.rule_name)) return false
    return true
  })
  
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
      const labels = parseLabels(item.labels)
      let columns = [{ prop: 'instance', label: '实例', width: 150 }]
      Object.keys(labels).forEach(key => {
        if (key !== 'instance' && key !== '__name__') {
          columns.push({ prop: key, label: key, width: 120 })
        }
      })
      columns.push({ prop: 'value', label: '值', width: 100 })
      
      ruleGroup = {
        rule_name: item.rule_name,
        columns,
        items: []
      }
      groups[item.group_id].rules.push(ruleGroup)
    }
    
    const labels = parseLabels(item.labels)
    const row: any = {
      instance: stripPort(item.instance),
      value: item.value.toFixed(2),
      unit: item.unit,
      status: item.status,
      ...labels
    }
    
    ruleGroup.items.push(row)
  })
  
  return Object.values(groups)
})

// ==================== 数据加载与图表渲染 ====================

const loadReport = async () => {
  const id = Number(route.params.id)
  loading.value = true
  try {
    const res = await inspectionApi.get(id)
    report.value = res.data.report
    items.value = res.data.items
    summaryForm.value.summary = report.value?.summary || ''
    summaryForm.value.remark = report.value?.remark || ''
    
    nextTick(() => {
      renderCharts()
    })
  } finally {
    loading.value = false
  }
}

const renderCharts = () => {
  const dates = []
  const now = new Date()
  for (let i = 6; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    dates.push(d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' }))
  }
  
  const serverItems = items.value.filter(i => !isProcessGroup(i.group_name) && i.show_in_table)
  const instances = [...new Set(serverItems.map(i => stripPort(i.instance)))]
  
  const cpuItems = serverItems.filter(i => 
    i.rule_name === 'CPU使用率' || 
    i.rule_name.toLowerCase() === 'cpu usage' ||
    (i.rule_name.includes('CPU使用率'))
  )
  
  const memItems = serverItems.filter(i => 
    i.rule_name === '内存使用率' || 
    i.rule_name.toLowerCase() === 'memory usage' ||
    (i.rule_name.includes('内存使用率'))
  )
  
  const diskItems = serverItems.filter(i => 
    i.rule_name === '磁盘使用率' || 
    i.rule_name.toLowerCase() === 'disk usage' ||
    (i.rule_name.includes('磁盘使用率'))
  )
  
  renderTrendCharts(dates, instances, cpuItems, memItems, diskItems)
}

const renderTrendCharts = (
  dates: string[], 
  instances: string[],
  cpuItems: InspectionItem[],
  memItems: InspectionItem[],
  diskItems: InspectionItem[]
) => {
  // CPU趋势图
  if (cpuChart.value) {
    const chart = echarts.init(cpuChart.value)
    const series: any[] = []
    
    instances.forEach((instance, idx) => {
      const item = cpuItems.find(i => stripPort(i.instance) === instance)
      let data: number[] = []
      
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
      
      if (data.length === 0 && item) {
        data = Array(7).fill(item.value)
      }
      
      if (data.length > 0) {
        series.push({
          name: instance,
          type: 'line',
          data: data.length >= 7 ? data.slice(0, 7) : [...data, ...Array(7 - data.length).fill(data[data.length - 1] || 0)],
          smooth: true,
          itemStyle: { color: colorPalette[idx % colorPalette.length] }
        })
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
        data: series.map(s => s.name),
        bottom: 0,
        type: 'scroll'
      },
      grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
  
  // 内存趋势图
  if (memChart.value) {
    const chart = echarts.init(memChart.value)
    const series: any[] = []
    
    instances.forEach((instance, idx) => {
      const item = memItems.find(i => stripPort(i.instance) === instance)
      let data: number[] = []
      
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
      
      if (data.length === 0 && item) {
        data = Array(7).fill(item.value)
      }
      
      if (data.length > 0) {
        series.push({
          name: instance,
          type: 'line',
          data: data.length >= 7 ? data.slice(0, 7) : [...data, ...Array(7 - data.length).fill(data[data.length - 1] || 0)],
          smooth: true,
          itemStyle: { color: colorPalette[idx % colorPalette.length] }
        })
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
      legend: { data: series.map(s => s.name), bottom: 0, type: 'scroll' },
      grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
  
  // 磁盘趋势图
  if (diskChart.value) {
    const chart = echarts.init(diskChart.value)
    const series: any[] = []
    
    instances.forEach((instance, idx) => {
      const item = diskItems.find(i => stripPort(i.instance) === instance)
      let data: number[] = []
      
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
      
      if (data.length === 0 && item) {
        data = Array(7).fill(item.value)
      }
      
      if (data.length > 0) {
        series.push({
          name: instance,
          type: 'line',
          data: data.length >= 7 ? data.slice(0, 7) : [...data, ...Array(7 - data.length).fill(data[data.length - 1] || 0)],
          smooth: true,
          itemStyle: { color: colorPalette[idx % colorPalette.length] }
        })
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
      legend: { data: series.map(s => s.name), bottom: 0, type: 'scroll' },
      grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
      series
    })
  }
}

// ==================== 操作方法 ====================

const handleExport = async () => {
  const element = document.getElementById('report-content')
  if (!element) return
  
  try {
    const canvas = await html2canvas(element, { scale: 2, useCORS: true, backgroundColor: '#fff' })
    const imgWidth = 210
    const pageHeight = 297
    const imgHeight = (canvas.height * imgWidth) / canvas.width
    const pdf = new jsPDF('p', 'mm', 'a4')
    
    let heightLeft = imgHeight
    let position = 0
    
    pdf.addImage(canvas.toDataURL('image/png'), 'PNG', 0, position, imgWidth, imgHeight)
    heightLeft -= pageHeight
    
    while (heightLeft > 0) {
      position = heightLeft - imgHeight
      pdf.addPage()
      pdf.addImage(canvas.toDataURL('image/png'), 'PNG', 0, position, imgWidth, imgHeight)
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
  await inspectionApi.updateSummary(report.value.id, summaryForm.value.summary, summaryForm.value.remark)
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

const goBack = () => router.push('/reports')

onMounted(() => loadReport())
</script>

<style scoped>
.page-container { background: #fff; padding: 20px; border-radius: 4px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; }
.section-card { margin-bottom: 20px; }
.section-title { font-weight: bold; font-size: 16px; }
.overview-cards { display: flex; flex-wrap: wrap; gap: 20px; }
.overview-card { flex: 1; min-width: 200px; padding: 15px; border: 1px solid #ebeef5; border-radius: 4px; background: #fafafa; }
.card-title { font-weight: bold; margin-bottom: 10px; font-size: 14px; }
.card-stats { display: flex; gap: 15px; }
.stat-item { display: flex; flex-direction: column; align-items: center; }
.stat-label { font-size: 12px; color: #999; }
.stat-value { font-size: 20px; font-weight: bold; }
.stat-value.critical { color: #F56C6C; }
.stat-value.warning { color: #E6A23C; }
.stat-value.normal { color: #67C23A; }
.trend-charts { display: flex; flex-wrap: wrap; gap: 20px; }
.chart-container { flex: 1; min-width: 300px; }
.chart { width: 100%; height: 300px; }
.rule-section { margin-bottom: 20px; }
.rule-title { font-weight: bold; margin-bottom: 10px; color: #606266; }
.table-note { margin-top: 10px; font-size: 12px; color: #909399; }
.status-critical { color: #F56C6C; font-weight: bold; }
.status-warning { color: #E6A23C; font-weight: bold; }
.status-normal { color: #67C23A; font-weight: bold; }
</style>

<style>
.cell-critical { background-color: #fef0f0 !important; color: #F56C6C !important; font-weight: bold; }
.cell-warning { background-color: #fdf6ec !important; color: #E6A23C !important; font-weight: bold; }
</style>
