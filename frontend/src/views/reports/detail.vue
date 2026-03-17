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

      <!-- 动态表格：一条规则一张表 -->
      <template v-for="[groupName, tables] in dynamicTablesByGroup" :key="groupName">
        <!-- 分组标题 -->
        <div class="group-section-header">{{ groupName }}</div>
        
        <!-- 三列布局（列数≤3） -->
        <div class="triple-column-cards" v-if="tables.filter(t => getTableLayout(t.columnCount) === 'triple').length > 0">
          <el-card 
            v-for="table in tables.filter(t => getTableLayout(t.columnCount) === 'triple')" 
            :key="table.ruleName"
            class="section-card third-width"
          >
            <template #header>
              <span class="section-title">{{ table.ruleName }}</span>
            </template>
            <el-table 
              :data="table.data" 
              stripe 
              border 
              size="small"
              :cell-class-name="({row, column}) => getDynamicTableCellClass(row, table.columns.find(c => c.prop === column.property)!, table.ruleType)"
            >
              <el-table-column
                v-for="col in table.columns"
                :key="col.prop"
                :prop="col.prop"
                :label="col.label"
                :min-width="col.width || 100"
              >
                <template #default="{ row }">
                  <template v-if="col.prop === '_value'">
                    {{ row._valueFormatted }}
                  </template>
                  <template v-else-if="col.prop === '_status'">
                    {{ row._status === 'normal' ? '正常' : row._status === 'warning' ? '告警' : row._status === 'critical' ? '异常' : '-' }}
                  </template>
                  <template v-else>
                    {{ row[col.prop] || '-' }}
                  </template>
                </template>
              </el-table-column>
            </el-table>
            <div class="table-note" v-if="getThresholdNote(table)">{{ getThresholdNote(table) }}</div>
          </el-card>
        </div>
        
        <!-- 全宽布局（列数>3） -->
        <el-card 
          v-for="table in tables.filter(t => getTableLayout(t.columnCount) === 'full')" 
          :key="table.ruleName"
          class="section-card"
        >
          <template #header>
            <span class="section-title">{{ table.ruleName }}</span>
          </template>
          <el-table 
            :data="table.data" 
            stripe 
            border 
            size="small"
            :cell-class-name="({row, column}) => getDynamicTableCellClass(row, table.columns.find(c => c.prop === column.property)!, table.ruleType)"
          >
            <el-table-column
              v-for="col in table.columns"
              :key="col.prop"
              :prop="col.prop"
              :label="col.label"
              :min-width="col.width || 120"
            >
              <template #default="{ row }">
                <template v-if="col.prop === '_value'">
                  {{ row._valueFormatted }}
                </template>
                <template v-else-if="col.prop === '_status'">
                  {{ row._status === 'normal' ? '正常' : row._status === 'warning' ? '告警' : row._status === 'critical' ? '异常' : '-' }}
                </template>
                <template v-else>
                  {{ row[col.prop] || '-' }}
                </template>
              </template>
            </el-table-column>
          </el-table>
          <div class="table-note" v-if="getThresholdNote(table)">{{ getThresholdNote(table) }}</div>
        </el-card>
      </template>

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

// ==================== 动态表格（一条规则一张表） ====================

// 动态表格列接口
interface DynamicTableColumn {
  prop: string      // 数据字段名（标签key）
  label: string     // 表头显示名（标签别名）
  isLabel: boolean  // 是否为标签列
  width?: number    // 列宽度
}

// 动态表格配置接口
interface DynamicTableConfig {
  ruleId: number
  ruleName: string
  groupName: string
  columns: DynamicTableColumn[]
  data: any[]
  ruleType: boolean      // 是否告警类型（决定是否显示背景色）
  threshold: number | null
  thresholdType: string
  unit: string
  columnCount: number    // 列数量，用于判断布局
}

// 获取需要动态展示表格的数据
// 条件：show_in_table=false 或者 不是基础资源分组
const dynamicTableItems = computed(() => {
  return items.value.filter(item => 
    !item.show_in_table || !isBasicResourceGroup(item.group_name)
  )
})

// 按规则分组生成表格配置
const dynamicTablesConfig = computed(() => {
  const ruleMap = new Map<string, InspectionItem[]>()
  
  // 按规则分组
  dynamicTableItems.value.forEach(item => {
    const key = `${item.group_name}__${item.rule_name}`
    if (!ruleMap.has(key)) {
      ruleMap.set(key, [])
    }
    ruleMap.get(key)!.push(item)
  })
  
  // 生成表格配置
  const tables: DynamicTableConfig[] = []
  
  ruleMap.forEach((ruleItems, key) => {
    if (ruleItems.length === 0) return
    
    const firstItem = ruleItems[0]
    
    // 优先使用 rule_labels（规则定义的标签别名映射），否则降级使用 labels 中的键名
    let columnLabels: Record<string, string>
    if (firstItem.rule_labels) {
      // 使用规则定义的标签别名映射
      columnLabels = parseLabels(firstItem.rule_labels)
    } else {
      // 降级：从 Prometheus 返回的 labels 中提取键名作为列名
      const promLabels = parseLabels(firstItem.labels || '{}')
      columnLabels = {}
      Object.keys(promLabels).forEach(key => {
        // 将键名转换为更友好的显示名称
        const displayNameMap: Record<string, string> = {
          'instance': '节点',
          'node': '节点',
          'device': '设备',
          'namespace': '命名空间',
          'pod': 'Pod名称',
          'persistentvolumeclaim': 'PVC名称',
          'mountpoint': '挂载点',
          'condition': '状态类型',
          'groupname': '进程名',
          'target': '域名'
        }
        columnLabels[key] = displayNameMap[key] || key
      })
    }
    
    // 解析标签列为动态列
    const columns: DynamicTableColumn[] = Object.entries(columnLabels).map(([labelKey, labelAlias]) => ({
      prop: labelKey,
      label: labelAlias as string,
      isLabel: true,
      width: getLabelColumnWidth(labelKey)
    }))
    
    // 添加值列和状态列
    columns.push(
      { prop: '_value', label: '值', isLabel: false, width: 100 },
      { prop: '_status', label: '状态', isLabel: false, width: 80 }
    )
    
    // 构建表格数据
    const data = ruleItems.map(item => {
      const itemLabels = parseLabels(item.labels || '{}')
      return {
        ...itemLabels,  // 展开标签
        _value: item.value,
        _status: item.status,
        _raw: item,  // 保留原始数据
        _valueFormatted: formatValue(item.value, item.unit)
      }
    })
    
    tables.push({
      ruleId: firstItem.rule_id,
      ruleName: firstItem.rule_name,
      groupName: firstItem.group_name,
      columns,
      data,
      ruleType: firstItem.table_column_rule_type,
      threshold: firstItem.rule_id ? ruleItems[0].value : null, // TODO: 需要从规则获取阈值
      thresholdType: '',
      unit: firstItem.unit,
      columnCount: columns.length
    })
  })
  
  // 按规则组名称排序，同组内按规则名称排序
  return tables.sort((a, b) => {
    if (a.groupName !== b.groupName) {
      return a.groupName.localeCompare(b.groupName, 'zh-CN')
    }
    return a.ruleName.localeCompare(b.ruleName, 'zh-CN')
  })
})

// 按规则组分组
const dynamicTablesByGroup = computed(() => {
  const groupMap = new Map<string, DynamicTableConfig[]>()
  
  dynamicTablesConfig.value.forEach(table => {
    if (!groupMap.has(table.groupName)) {
      groupMap.set(table.groupName, [])
    }
    groupMap.get(table.groupName)!.push(table)
  })
  
  return groupMap
})

// 根据列数判断布局类型
const getTableLayout = (columnCount: number): 'triple' | 'full' => {
  return columnCount <= 3 ? 'triple' : 'full'
}

// 获取标签列宽度
const getLabelColumnWidth = (labelKey: string): number => {
  const widthMap: Record<string, number> = {
    'node': 150,
    'instance': 130,
    'namespace': 120,
    'pod': 200,
    'persistentvolumeclaim': 200,
    'pvc': 200,
    'mountpoint': 100,
    'device': 100,
    'groupname': 150
  }
  return widthMap[labelKey] || 120
}

// 动态表格单元格样式
const getDynamicTableCellClass = (row: any, column: DynamicTableColumn, ruleType: boolean) => {
  // 只对状态列显示背景色
  if (column.prop !== '_status') return ''
  
  // 如果是展示类型（非告警），不显示背景色
  if (!ruleType) return ''
  
  // 根据状态返回对应样式
  const status = row._status
  if (status === 'critical') return 'cell-critical'
  if (status === 'warning') return 'cell-warning'
  if (status === 'normal') return 'cell-normal'
  return ''
}

// 获取阈值说明文字
const getThresholdNote = (table: DynamicTableConfig): string => {
  const rawItem = table.data[0]?._raw
  if (!rawItem) return ''
  
  // 从原始数据中获取阈值信息
  // 注意：当前 InspectionItem 没有阈值字段，这里简化处理
  // 可以根据规则类型返回默认说明
  if (table.ruleName.includes('证书')) {
    return '说明：值≥30天表示正常，值<30天表示异常'
  }
  if (table.ruleName.includes('Pod') || table.ruleName.includes('pod')) {
    return '说明：值=1表示正常，值≠1表示异常'
  }
  if (table.ruleName.includes('节点就绪') || table.ruleName.includes('节点状态')) {
    return '说明：值=0表示正常，值≠0表示异常'
  }
  if (table.ruleName.includes('PVC') || table.ruleName.includes('存储')) {
    return '说明：值≥90%表示异常，值<90%表示正常'
  }
  
  return ''
}

// ==================== K8S相关表格 ====================

// 判断是否为K8S分组
const isK8SGroup = (groupName: string): boolean => {
  const lower = groupName.toLowerCase()
  return lower.includes('k8s') || lower.includes('kubernetes') || lower.includes('容器') || lower.includes('pod') || lower.includes('container')
}

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

/* 分组标题 */
.group-section-header {
  font-size: 16px;
  font-weight: bold;
  color: #303133;
  margin: 20px 0 15px 0;
  padding-left: 10px;
  border-left: 4px solid #409EFF;
}

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
