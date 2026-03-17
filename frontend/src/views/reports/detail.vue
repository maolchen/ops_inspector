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
          <!-- IP 列（固定首列） -->
          <el-table-column prop="ip" label="IP地址" min-width="130" fixed="left" />
          
          <!-- 动态列（循环生成） -->
          <el-table-column
            v-for="col in tableColumns"
            :key="col.prop"
            :prop="col.prop"
            :label="col.label"
            :min-width="col.width > 0 ? col.width : 100"
          />
        </el-table>
      </el-card>

      <!-- K8S节点就绪状态 -->
      <el-card class="section-card" v-if="k8sNodeTableData.length > 0">
        <template #header>
          <span class="section-title">K8S节点就绪状态</span>
        </template>
        <el-table :data="k8sNodeTableData" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.value === 0 ? 'cell-normal' : 'cell-critical') : ''">
          <el-table-column prop="node" label="节点" min-width="200" />
          <el-table-column prop="statusType" label="状态类型" width="100">
            <template #default>Ready</template>
          </el-table-column>
          <el-table-column prop="value" label="值" width="80" />
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              {{ row.value === 0 ? '正常' : '异常' }}
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
        <el-table :data="k8sPodTableData" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.value === 1 ? 'cell-normal' : 'cell-critical') : ''">
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column prop="pod" label="Pod名" min-width="300" />
          <el-table-column prop="value" label="运行状态" width="100" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              {{ row.value === 1 ? '正常' : '异常' }}
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
        <el-table :data="k8sPVCTableData" stripe border size="small" :cell-class-name="({row, column}) => (column.property === 'usedPercent' || column.property === 'status') ? (row.usedPercent >= 90 ? 'cell-critical' : 'cell-normal') : ''">
          <el-table-column prop="pvc" label="PVC名称" min-width="300" />
          <el-table-column prop="namespace" label="命名空间" width="150" />
          <el-table-column prop="usedPercent" label="PVC使用率" width="120">
            <template #default="{ row }">
              {{ row.usedPercent.toFixed(2) }}%
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              {{ row.usedPercent >= 90 ? '异常' : '正常' }}
            </template>
          </el-table-column>
        </el-table>
        <div class="table-note">说明：值>=90%表示异常，值<90%表示正常</div>
      </el-card>

      <!-- K8S证书状态 - 三列布局 -->
      <div class="triple-column-cards" v-if="Object.keys(k8sCertGroupedData).length > 0">
        <template v-for="(certGroup, certType) in k8sCertGroupedData" :key="certType">
          <el-card class="section-card third-width" v-if="certGroup.length > 0">
            <template #header>
              <span class="section-title">{{ certType }}</span>
            </template>
            <el-table :data="certGroup" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.value >= 30 ? 'cell-normal' : 'cell-warning') : ''">
              <el-table-column prop="node" label="节点" min-width="120" />
              <el-table-column prop="value" label="值" width="80">
                <template #default="{ row }">
                  {{ row.value }}天
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="60">
                <template #default="{ row }">
                  {{ row.value >= 30 ? '正常' : '异常' }}
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </template>
      </div>

      <!-- 磁盘IO表格（双列布局） -->
      <div class="dual-column-cards" v-if="diskWriteTableData.length > 0 || diskReadTableData.length > 0">
        <el-card class="section-card half-width" v-if="diskWriteTableData.length > 0">
          <template #header>
            <span class="section-title">30分钟内磁盘平均写入值</span>
          </template>
          <el-table :data="diskWriteTableData" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.status === 'normal' ? 'cell-normal' : 'cell-warning') : ''">
            <el-table-column prop="instance" label="节点" min-width="120" />
            <el-table-column prop="device" label="设备" min-width="100" />
            <el-table-column prop="valueFormatted" label="值" width="100" />
            <el-table-column prop="status" label="状态" width="70">
              <template #default="{ row }">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="section-card half-width" v-if="diskReadTableData.length > 0">
          <template #header>
            <span class="section-title">30分钟内磁盘平均读取值</span>
          </template>
          <el-table :data="diskReadTableData" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.status === 'normal' ? 'cell-normal' : 'cell-warning') : ''">
            <el-table-column prop="instance" label="节点" min-width="120" />
            <el-table-column prop="device" label="设备" min-width="100" />
            <el-table-column prop="valueFormatted" label="值" width="100" />
            <el-table-column prop="status" label="状态" width="70">
              <template #default="{ row }">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>

      <!-- 网络IO表格（双列布局） -->
      <div class="dual-column-cards" v-if="networkUploadTableData.length > 0 || networkDownloadTableData.length > 0">
        <el-card class="section-card half-width" v-if="networkUploadTableData.length > 0">
          <template #header>
            <span class="section-title">30分钟内上传速率</span>
          </template>
          <el-table :data="networkUploadTableData" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.status === 'normal' ? 'cell-normal' : 'cell-warning') : ''">
            <el-table-column prop="instance" label="节点" min-width="120" />
            <el-table-column prop="device" label="设备" min-width="100" />
            <el-table-column prop="valueFormatted" label="值" width="100" />
            <el-table-column prop="status" label="状态" width="70">
              <template #default="{ row }">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="section-card half-width" v-if="networkDownloadTableData.length > 0">
          <template #header>
            <span class="section-title">30分钟内下载速率</span>
          </template>
          <el-table :data="networkDownloadTableData" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.status === 'normal' ? 'cell-normal' : 'cell-warning') : ''">
            <el-table-column prop="instance" label="节点" min-width="120" />
            <el-table-column prop="device" label="设备" min-width="100" />
            <el-table-column prop="valueFormatted" label="值" width="100" />
            <el-table-column prop="status" label="状态" width="70">
              <template #default="{ row }">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>

      <!-- 进程TOP表格（双列布局） -->
      <div class="dual-column-cards" v-if="processCPUTableData.length > 0 || processMemTableData.length > 0">
        <el-card class="section-card half-width" v-if="processCPUTableData.length > 0">
          <template #header>
            <span class="section-title">进程CPU使用率Top5</span>
          </template>
          <el-table :data="processCPUTableData" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.status === 'normal' ? 'cell-normal' : 'cell-warning') : ''">
            <el-table-column prop="processName" label="进程名" min-width="150" />
            <el-table-column prop="instance" label="所在机器" min-width="120" />
            <el-table-column prop="value" label="值" width="80" />
            <el-table-column prop="status" label="状态" width="70">
              <template #default="{ row }">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="section-card half-width" v-if="processMemTableData.length > 0">
          <template #header>
            <span class="section-title">进程内存使用率Top5</span>
          </template>
          <el-table :data="processMemTableData" stripe border size="small" :cell-class-name="({row, column}) => column.property === 'status' ? (row.status === 'normal' ? 'cell-normal' : 'cell-warning') : ''">
            <el-table-column prop="processName" label="进程名" min-width="150" />
            <el-table-column prop="instance" label="所在机器" min-width="120" />
            <el-table-column prop="value" label="值" width="80" />
            <el-table-column prop="status" label="状态" width="70">
              <template #default="{ row }">
                {{ row.status === 'normal' ? '正常' : '告警' }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>

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
          <el-table :data="ruleDetail.items" stripe size="small" border :cell-class-name="({row, column}) => column.property === 'value' ? (row.status === 'critical' ? 'cell-critical' : row.status === 'warning' ? 'cell-warning' : row.status === 'normal' ? 'cell-normal' : '') : ''">
            <el-table-column
              v-for="col in ruleDetail.columns"
              :key="col.prop"
              :prop="col.prop"
              :label="col.label"
              :width="col.width"
            >
              <template #default="{ row }">
                {{ row[col.prop] }}{{ row.unit ? ` ${row.unit}` : '' }}
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
            <div class="remark-editor-wrapper">
              <div class="editor-toolbar">
                <el-button size="small" @click="insertImage('remark')">插入图片</el-button>
                <span class="editor-hint">支持 Ctrl+V 粘贴图片</span>
              </div>
              <div
                ref="remarkEditor"
                class="rich-editor"
                contenteditable="true"
                @paste="handlePaste"
                @input="syncRemarkContent"
                v-html="summaryForm.remark"
              ></div>
            </div>
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
const remarkEditor = ref<HTMLElement>()

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

// 动态表格列接口
interface TableColumn {
  prop: string           // 列属性名
  label: string          // 列标题
  width: number          // 列宽度
  order: number          // 列顺序
  type: 'value' | 'label' | 'mountpoint' // 列类型：value=规则值, label=从labels提取, mountpoint=挂载点（特殊处理）
  merge: boolean         // 是否参与合并
  unit: string           // 单位
  isDiskRelated: boolean // 是否与磁盘相关（影响按挂载点查询）
  ruleType: boolean      // 规则类型：true=告警（显示背景色），false=展示（不显示背景色）
}

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

// 判断规则名是否与磁盘相关
const isDiskRelatedRule = (ruleName: string): boolean => {
  const lower = ruleName.toLowerCase()
  return lower.includes('磁盘') || 
         lower.includes('disk') || 
         lower.includes('挂载') ||
         lower.includes('mount')
}

// 动态表格列配置
const tableColumns = computed(() => {
  // 只取基础资源分组且 show_in_table 的数据
  const basicItems = items.value.filter(i => isBasicResourceGroup(i.group_name) && i.show_in_table)
  
  // 收集所有唯一的列配置
  const columnMap = new Map<string, TableColumn>()
  let hasDiskRelatedColumn = false
  let minDiskColumnOrder = Infinity
  
  basicItems.forEach(item => {
    const key = `value_${item.rule_name}`
    const isDiskRelated = isDiskRelatedRule(item.rule_name)
    
    if (isDiskRelated) {
      hasDiskRelatedColumn = true
      if (item.table_column_order < minDiskColumnOrder) {
        minDiskColumnOrder = item.table_column_order
      }
    }
    
    if (!columnMap.has(key)) {
      columnMap.set(key, {
        prop: key,
        label: item.rule_name,
        width: item.table_column_width || 100,
        order: item.table_column_order || 0,
        type: 'value',
        merge: item.table_column_merge,
        unit: item.unit,
        isDiskRelated,
        ruleType: item.table_column_rule_type  // 告警规则显示背景色，展示规则不显示
      })
    }
  })
  
  // 如果有磁盘相关列，自动添加挂载点列（在磁盘列之前）
  if (hasDiskRelatedColumn) {
    columnMap.set('mountpoint', {
      prop: 'mountpoint',
      label: '挂载点',
      width: 100,
      order: minDiskColumnOrder - 1, // 在磁盘列之前
      type: 'mountpoint',
      merge: false, // 挂载点不参与合并
      unit: '',
      isDiskRelated: true,
      ruleType: false  // 挂载点不显示背景色
    })
  }
  
  // 按顺序排序
  return Array.from(columnMap.values()).sort((a, b) => a.order - b.order)
})

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

// 创建基础资源行数据（动态版本）
const createBasicRow = (ip: string, data: Record<string, any>, mountpoint: string, isFirst: boolean = true) => {
  const row: Record<string, any> = {
    _ip: ip,
    _isFirst: isFirst,
    ip: isFirst ? ip : ''
  }
  
  // 根据列配置动态填充数据
  tableColumns.value.forEach(col => {
    if (col.type === 'mountpoint') {
      // 挂载点列：直接使用传入的 mountpoint 参数
      row[col.prop] = mountpoint || '-'
    } else if (col.type === 'label') {
      // 其他 label 列：从数据中提取
      const firstItem = Object.values(data)[0] as any
      const labels = firstItem?.labels || {}
      row[col.prop] = labels[col.label] || '-'
    } else {
      // 规则值列
      const ruleName = col.label
      
      if (col.isDiskRelated) {
        // 磁盘相关列：按挂载点查找数据
        const mpKey = `${ruleName}_${mountpoint}`
        const item = data[mpKey]
        
        if (item && item.value !== undefined) {
          row[col.prop] = formatValue(item.value, col.unit)
          row[`${col.prop}_status`] = item.status
        } else {
          row[col.prop] = '-'
          row[`${col.prop}_status`] = ''
        }
      } else {
        // 非磁盘列：按 IP 查找数据（不区分挂载点）
        const item = data[ruleName]
        
        if (item && item.value !== undefined) {
          // 非首行的合并列显示空，首行显示值
          if (!isFirst && col.merge) {
            row[col.prop] = ''
            row[`${col.prop}_status`] = ''
          } else {
            row[col.prop] = formatValue(item.value, col.unit)
            row[`${col.prop}_status`] = item.status
          }
        } else {
          row[col.prop] = isFirst ? '-' : ''
          row[`${col.prop}_status`] = ''
        }
      }
    }
  })
  
  return row
}

// 格式化数值
const formatValue = (value: number, unit: string): string => {
  if (value === undefined || value === null) return '-'
  
  // 字节单位
  if (unit === 'bytes' || unit === 'B') {
    return formatBytesToHuman(value)
  }
  
  // 百分比
  if (unit === '%') {
    return value.toFixed(2) + '%'
  }
  
  // 秒数（运行时间）
  if (unit === 's' || unit === 'seconds') {
    return formatUptime(value)
  }
  
  // 整数
  if (Number.isInteger(value)) {
    return value.toString()
  }
  
  // 默认保留两位小数
  return value.toFixed(2)
}

// 基础资源表格单元格合并方法（动态版本）
const basicResourceSpanMethod = ({ row, column, rowIndex, columnIndex }: { row: any; column: any; rowIndex: number; columnIndex: number }) => {
  // 第一列（IP列）始终参与合并
  if (columnIndex === 0) {
    if (row._rowSpan > 0) {
      return { rowspan: row._rowSpan, colspan: 1 }
    } else {
      return { rowspan: 0, colspan: 0 }
    }
  }
  
  // 动态列：根据列配置的 merge 属性决定是否合并
  const prop = column.property
  const col = tableColumns.value.find(c => c.prop === prop)
  
  if (col && col.merge) {
    if (row._rowSpan > 0) {
      return { rowspan: row._rowSpan, colspan: 1 }
    } else {
      return { rowspan: 0, colspan: 0 }
    }
  }
}

// 基础资源表格单元格样式
const getBasicTableCellClass = ({ row, column }: { row: any; column: any }) => {
  const propName = column.property
  if (!propName) return ''
  
  // 查找对应的列配置
  const colConfig = tableColumns.value.find(col => col.prop === propName)
  
  // 如果是展示类型（ruleType=false），不显示背景色
  if (colConfig && !colConfig.ruleType) {
    return ''
  }
  
  // 状态 key 格式：prop_status（与 createBasicRow 中设置的一致）
  const statusKey = propName + '_status'
  const status = row[statusKey]
  if (status === 'critical') return 'cell-critical'
  if (status === 'warning') return 'cell-warning'
  if (status === 'normal') return 'cell-normal'
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
// 按命名空间排序
const k8sPodTableData = computed(() => {
  const podItems = items.value.filter(i => 
    isK8SGroup(i.group_name) && 
    (i.rule_name.includes('Pod状态') || i.rule_name.includes('Pod运行') || 
     i.rule_name.toLowerCase().includes('pod status'))
  )
  
  const data = podItems.map(item => {
    const labels = parseLabels(item.labels)
    return {
      namespace: labels.namespace || 'default',
      pod: labels.pod || stripPort(item.instance),
      value: item.value
    }
  })
  
  // 按命名空间排序
  return data.sort((a, b) => a.namespace.localeCompare(b.namespace))
})

// K8S PVC使用率表格数据
// 列顺序：PVC名称、命名空间、PVC使用率、状态
// 值>=90%表示异常
// 按命名空间排序
const k8sPVCTableData = computed(() => {
  const pvcItems = items.value.filter(i => 
    isK8SGroup(i.group_name) && 
    (i.rule_name.includes('PVC') || i.rule_name.includes('持久卷') || i.rule_name.toLowerCase().includes('pvc') || i.rule_name.includes('存储卷'))
  )
  
  const data = pvcItems.map(item => {
    const labels = parseLabels(item.labels)
    return {
      pvc: labels.persistentvolumeclaim || 'unknown',
      namespace: labels.namespace || 'default',
      usedPercent: item.value
    }
  })
  
  // 按命名空间排序
  return data.sort((a, b) => a.namespace.localeCompare(b.namespace))
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
    
    // 图例单选：点击某个图例仅展示该线条
    const allSeriesNames = series.map(s => s.name)
    chart.on('legendselectchanged', (params: any) => {
      const selectedName = params.name
      // 构建新的选中状态：只有被点击的为 true
      const newSelected: Record<string, boolean> = {}
      allSeriesNames.forEach(name => {
        newSelected[name] = name === selectedName
      })
      chart.setOption({ legend: { selected: newSelected } })
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
    
    // 图例单选：点击某个图例仅展示该线条
    const allSeriesNames = series.map(s => s.name)
    chart.on('legendselectchanged', (params: any) => {
      const selectedName = params.name
      const newSelected: Record<string, boolean> = {}
      allSeriesNames.forEach(name => {
        newSelected[name] = name === selectedName
      })
      chart.setOption({ legend: { selected: newSelected } })
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
    
    // 图例单选：点击某个图例仅展示该线条
    const allSeriesNames = series.map(s => s.name)
    chart.on('legendselectchanged', (params: any) => {
      const selectedName = params.name
      const newSelected: Record<string, boolean> = {}
      allSeriesNames.forEach(name => {
        newSelected[name] = name === selectedName
      })
      chart.setOption({ legend: { selected: newSelected } })
    })
  }
}

// ==================== 操作方法 ====================

const handleExport = async () => {
  const element = document.getElementById('report-content')
  if (!element) return
  
  try {
    // 优化PDF导出：降低分辨率减小文件大小
    const canvas = await html2canvas(element, { 
      scale: 1.5,  // 降低缩放比例，减小文件大小
      useCORS: true, 
      backgroundColor: '#fff',
      logging: false,
      imageTimeout: 15000
    })
    
    const imgWidth = 210  // A4宽度（mm）
    const pageHeight = 297  // A4高度（mm）
    const imgHeight = (canvas.height * imgWidth) / canvas.width
    
    // 使用JPEG格式并设置质量，大幅减小文件大小
    const imgData = canvas.toDataURL('image/jpeg', 0.7)
    
    const pdf = new jsPDF('p', 'mm', 'a4')
    
    let heightLeft = imgHeight
    let position = 0
    
    pdf.addImage(imgData, 'JPEG', 0, position, imgWidth, imgHeight)
    heightLeft -= pageHeight
    
    while (heightLeft > 0) {
      position = heightLeft - imgHeight
      pdf.addPage()
      pdf.addImage(imgData, 'JPEG', 0, position, imgWidth, imgHeight)
      heightLeft -= pageHeight
    }
    
    pdf.save(`巡检报告_${report.value?.project_name}_${new Date().toISOString().slice(0, 10)}.pdf`)
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('PDF导出失败:', error)
    ElMessage.error('导出失败')
  }
}

const handleSaveSummary = async () => {
  if (!report.value) return
  await inspectionApi.updateSummary(report.value.id, summaryForm.value.summary, summaryForm.value.remark)
  ElMessage.success('保存成功')
}

// 处理粘贴事件（支持粘贴图片）
const handlePaste = (e: ClipboardEvent) => {
  const items = e.clipboardData?.items
  if (!items) return
  
  for (let i = 0; i < items.length; i++) {
    const item = items[i]
    if (item.type.indexOf('image') !== -1) {
      e.preventDefault()
      const file = item.getAsFile()
      if (file) {
        const reader = new FileReader()
        reader.onload = (event) => {
          const imgHtml = `<img src="${event.target?.result}" alt="粘贴的图片" style="max-width:100%;max-height:300px;margin:5px 0;">`
          document.execCommand('insertHTML', false, imgHtml)
          syncRemarkContent()
        }
        reader.readAsDataURL(file)
      }
      break
    }
  }
}

// 同步富文本内容到form
const syncRemarkContent = () => {
  if (remarkEditor.value) {
    summaryForm.value.remark = remarkEditor.value.innerHTML
  }
}

// 插入图片（点击按钮）
const insertImage = (type: string) => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  input.onchange = (e: Event) => {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (event) => {
        const imgHtml = `<img src="${event.target?.result}" alt="插入的图片" style="max-width:100%;max-height:300px;margin:5px 0;">`
        if (remarkEditor.value) {
          remarkEditor.value.focus()
          document.execCommand('insertHTML', false, imgHtml)
          syncRemarkContent()
        }
      }
      reader.readAsDataURL(file)
    }
  }
  input.click()
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

/* 双列卡片布局 */
.dual-column-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  margin-bottom: 20px;
}
.dual-column-cards .section-card.half-width {
  flex: 1 1 calc(50% - 10px);
  min-width: 400px;
  margin-bottom: 0;
}
.dual-column-cards .section-card.half-width :deep(.el-card__body) {
  padding: 15px;
}
.dual-column-cards .section-card.half-width :deep(.el-table) {
  width: 100%;
}
.dual-column-cards .section-card.half-width :deep(.el-table__body-wrapper) {
  width: 100% !important;
}

/* 三列卡片布局 */
.triple-column-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  margin-bottom: 20px;
}
.triple-column-cards .section-card.third-width {
  flex: 1 1 calc(33.33% - 10px);
  min-width: 280px;
  margin-bottom: 0;
}
.triple-column-cards .section-card.third-width :deep(.el-card__body) {
  padding: 12px;
}
.triple-column-cards .section-card.third-width :deep(.el-table) {
  width: 100%;
}
.triple-column-cards .section-card.third-width :deep(.el-table th) {
  padding: 6px 0;
}
.triple-column-cards .section-card.third-width :deep(.el-table td) {
  padding: 6px 0;
}

/* 富文本编辑器样式 */
.remark-editor-wrapper {
  width: 100%;
}
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  padding: 5px;
  background: #f5f7fa;
  border: 1px solid #dcdfe6;
  border-bottom: none;
  border-radius: 4px 4px 0 0;
}
.editor-hint {
  font-size: 12px;
  color: #909399;
}
.rich-editor {
  min-height: 120px;
  padding: 10px;
  border: 1px solid #dcdfe6;
  border-radius: 0 0 4px 4px;
  background: #fff;
  line-height: 1.6;
  overflow-y: auto;
  max-height: 400px;
}
.rich-editor:focus {
  border-color: #409eff;
  outline: none;
}
.rich-editor img {
  max-width: 100%;
  max-height: 300px;
  margin: 5px 0;
  border-radius: 4px;
}
.rich-editor p {
  margin: 5px 0;
}
</style>

<style>
/* 状态单元格背景色样式 */
.cell-normal { 
  background-color: #f0f9eb !important; 
  color: #67C23A !important; 
  font-weight: bold;
}
.cell-warning { 
  background-color: #fdf6ec !important; 
  color: #E6A23C !important; 
  font-weight: bold;
}
.cell-critical { 
  background-color: #fef0f0 !important; 
  color: #F56C6C !important; 
  font-weight: bold;
}
</style>
