<template>
  <div class="page-container">
    <div class="page-header">
      <h2>规则配置</h2>
      <div>
        <el-button type="primary" @click="handleAddGroup">
          <el-icon><Plus /></el-icon>
          新增规则组
        </el-button>
        <el-button type="success" @click="handleAddRule">
          <el-icon><Plus /></el-icon>
          新增规则
        </el-button>
      </div>
    </div>

    <el-collapse v-model="activeGroups" v-loading="loading">
      <el-collapse-item
        v-for="group in groups"
        :key="group.id"
        :name="group.id"
      >
        <template #title>
          <div class="group-header">
            <span class="group-name">{{ group.name }}</span>
            <span class="group-desc">{{ group.description }}</span>
            <el-button type="primary" link @click.stop="handleEditGroup(group)">编辑组</el-button>
          </div>
        </template>

        <el-table :data="getGroupRules(group.id)" stripe size="small">
          <el-table-column prop="name" label="规则名称" width="200" />
          <el-table-column prop="type" label="类型" width="80">
            <template #default="{ row }">
              <el-tag :type="row.type ? 'warning' : 'info'" size="small">
                {{ row.type ? '告警' : '展示' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="show_in_table" label="表格展示" width="80">
            <template #default="{ row }">
              {{ row.show_in_table ? '是' : '否' }}
            </template>
          </el-table-column>
          <el-table-column prop="threshold" label="阈值" width="100">
            <template #default="{ row }">
              <span v-if="row.threshold !== null && row.threshold !== undefined">{{ row.threshold }} {{ row.unit }}</span>
              <span v-else style="color: #999">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="enabled" label="启用" width="80">
            <template #default="{ row }">
              <el-switch
                v-model="row.enabled"
                @change="handleToggleRule(row)"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button type="primary" link @click="handleEditRule(row)">编辑</el-button>
              <el-button type="danger" link @click="handleDeleteRule(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>
    </el-collapse>

    <!-- 规则组对话框 -->
    <el-dialog v-model="groupDialogVisible" :title="groupEditMode ? '编辑规则组' : '新增规则组'" width="500px">
      <el-form :model="groupForm" label-width="100px" :rules="groupRules" ref="groupFormRef">
        <el-form-item label="组名称" prop="name">
          <el-input v-model="groupForm.name" placeholder="请输入规则组名称" />
        </el-form-item>
        <el-form-item label="组标识" prop="code">
          <el-input v-model="groupForm.code" placeholder="如：basic_resources" :disabled="groupEditMode" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="groupForm.description" type="textarea" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="groupForm.sort_order" :min="1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="groupDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitGroup" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 规则对话框 -->
    <el-dialog v-model="ruleDialogVisible" :title="ruleEditMode ? '编辑规则' : '新增规则'" width="700px">
      <el-form :model="ruleForm" label-width="120px" :rules="ruleRules" ref="ruleFormRef">
        <el-form-item label="所属规则组" prop="group_id">
          <el-select v-model="ruleForm.group_id" placeholder="选择规则组">
            <el-option
              v-for="g in groups"
              :key="g.id"
              :label="g.name"
              :value="g.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="ruleForm.name" placeholder="如：CPU使用率" />
        </el-form-item>
        <el-form-item label="规则类型">
          <el-radio-group v-model="ruleForm.type">
            <el-radio :value="true">告警规则</el-radio>
            <el-radio :value="false">仅展示</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="表格展示">
          <el-switch v-model="ruleForm.show_in_table" />
        </el-form-item>
        
        <!-- 表格列配置（仅当表格展示开启时显示） -->
        <template v-if="ruleForm.show_in_table">
          <el-divider content-position="left">表格列配置</el-divider>
          <el-form-item label="列顺序">
            <el-input-number v-model="ruleForm.table_column_order" :min="0" :max="999" />
            <div class="form-tip" style="color: #909399; font-size: 12px; margin-top: 4px;">
              数字越小越靠前，相同顺序按名称排序。磁盘相关列会自动在其前面添加"挂载点"列
            </div>
          </el-form-item>
          <el-form-item label="列宽度">
            <el-input-number v-model="ruleForm.table_column_width" :min="0" :max="500" />
            <span style="margin-left: 8px; color: #909399; font-size: 12px;">像素，0 表示自适应</span>
          </el-form-item>
          <el-form-item label="合并单元格">
            <el-switch v-model="ruleForm.table_column_merge" />
            <div class="form-tip" style="color: #909399; font-size: 12px; margin-top: 4px;">
              同一 IP 多行数据时是否合并显示。CPU、内存等非磁盘指标建议开启，磁盘相关指标建议关闭
            </div>
          </el-form-item>
        </template>
        
        <el-form-item label="指标描述">
          <el-input v-model="ruleForm.description" type="textarea" />
        </el-form-item>
        <el-form-item label="即时查询" prop="query">
          <el-input v-model="ruleForm.query" type="textarea" :rows="3" placeholder="PromQL 查询语句" />
        </el-form-item>
        <el-form-item label="趋势查询">
          <el-input v-model="ruleForm.trend_query" type="textarea" :rows="2" placeholder="可选，用于图表展示" />
        </el-form-item>
        <el-form-item label="阈值">
          <el-input-number v-model="ruleForm.threshold" :precision="2" :step="0.1" />
          <el-select v-model="ruleForm.threshold_type" style="width: 150px; margin-left: 10px">
            <el-option label="大于阈值告警" value="greater" />
            <el-option label="大于等于告警" value="greater_equal" />
            <el-option label="小于阈值正常" value="less" />
            <el-option label="小于等于正常" value="less_equal" />
            <el-option label="等于阈值正常" value="equal" />
            <el-option label="至少达到阈值" value="at_least" />
          </el-select>
        </el-form-item>
        <el-form-item label="单位">
          <el-input v-model="ruleForm.unit" placeholder="如：%, GB, 核" />
        </el-form-item>
        <el-form-item label="标签别名">
          <el-input v-model="ruleForm.labels" type="textarea" :rows="2" placeholder='JSON格式: {"instance": "节点IP"}' />
        </el-form-item>
        <el-form-item label="适用项目">
          <el-input v-model="ruleForm.project_scope" placeholder="* 表示所有项目，或输入项目名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitRule" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { ruleGroupApi, ruleApi, type RuleGroup, type Rule } from '../../api/rule'

// 判断是否为磁盘相关规则
const isDiskRelatedRule = (name: string): boolean => {
  const lower = name.toLowerCase()
  return lower.includes('磁盘') || lower.includes('disk') || lower.includes('挂载') || lower.includes('mount')
}

const loading = ref(false)
const submitting = ref(false)
const groups = ref<RuleGroup[]>([])
const rules = ref<Rule[]>([])
const activeGroups = ref<number[]>([])

const groupDialogVisible = ref(false)
const groupEditMode = ref(false)
const groupFormRef = ref()
const groupForm = ref<Partial<RuleGroup>>({
  name: '',
  code: '',
  description: '',
  sort_order: 1
})

const ruleDialogVisible = ref(false)
const ruleEditMode = ref(false)
const ruleFormRef = ref()
const ruleForm = ref<Partial<Rule>>({
  group_id: undefined,
  name: '',
  type: true,
  show_in_table: false,
  description: '',
  query: '',
  trend_query: '',
  threshold: null,
  unit: '',
  labels: '',
  threshold_type: 'greater',
  project_scope: '*',
  enabled: true,
  // 表格列配置
  table_column_order: 0,
  table_column_width: 100,
  table_column_type: 'value',
  table_column_label: '',
  table_column_merge: true
})

const groupRules = {
  name: [{ required: true, message: '请输入规则组名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入规则组标识', trigger: 'blur' }]
}

const ruleRules = {
  group_id: [{ required: true, message: '请选择规则组', trigger: 'change' }],
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  query: [{ required: true, message: '请输入 PromQL 查询', trigger: 'blur' }]
}

const getGroupRules = (groupId: number) => {
  return rules.value.filter(r => r.group_id === groupId)
}

const loadData = async () => {
  loading.value = true
  try {
    const [groupRes, ruleRes] = await Promise.all([
      ruleGroupApi.list(),
      ruleApi.list()
    ])
    groups.value = groupRes.data
    rules.value = ruleRes.data
    activeGroups.value = groups.value.map(g => g.id)
  } finally {
    loading.value = false
  }
}

// 规则组操作
const handleAddGroup = () => {
  groupEditMode.value = false
  groupForm.value = { name: '', code: '', description: '', sort_order: 1 }
  groupDialogVisible.value = true
}

const handleEditGroup = (row: RuleGroup) => {
  groupEditMode.value = true
  groupForm.value = { ...row }
  groupDialogVisible.value = true
}

const handleSubmitGroup = async () => {
  await groupFormRef.value.validate()
  submitting.value = true
  try {
    if (groupEditMode.value) {
      await ruleGroupApi.update(groupForm.value.id!, groupForm.value)
      ElMessage.success('更新成功')
    } else {
      await ruleGroupApi.create(groupForm.value)
      ElMessage.success('创建成功')
    }
    groupDialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

// 规则操作
const handleAddRule = () => {
  ruleEditMode.value = false
  ruleForm.value = {
    group_id: groups.value[0]?.id,
    name: '',
    type: true,
    show_in_table: false,
    description: '',
    query: '',
    trend_query: '',
    threshold: null,
    unit: '',
    labels: '',
    threshold_type: 'greater',
    project_scope: '*',
    enabled: true,
    // 表格列配置
    table_column_order: 0,
    table_column_width: 100,
    table_column_type: 'value',
    table_column_label: '',
    table_column_merge: true
  }
  ruleDialogVisible.value = true
}

const handleEditRule = (row: Rule) => {
  ruleEditMode.value = true
  ruleForm.value = { ...row }
  ruleDialogVisible.value = true
}

const handleDeleteRule = async (row: Rule) => {
  await ElMessageBox.confirm('确定要删除该规则吗？', '提示', { type: 'warning' })
  await ruleApi.delete(row.id)
  ElMessage.success('删除成功')
  loadData()
}

const handleToggleRule = async (row: Rule) => {
  await ruleApi.toggle(row.id)
  ElMessage.success('状态已更新')
}

const handleSubmitRule = async () => {
  await ruleFormRef.value.validate()
  submitting.value = true
  try {
    if (ruleEditMode.value) {
      await ruleApi.update(ruleForm.value.id!, ruleForm.value)
      ElMessage.success('更新成功')
    } else {
      await ruleApi.create(ruleForm.value)
      ElMessage.success('创建成功')
    }
    ruleDialogVisible.value = false
    loadData()
  } finally {
    submitting.value = false
  }
}

// 监听规则名称变化，自动设置合并单元格默认值
watch(() => ruleForm.value.name, (newName) => {
  // 只在新增模式且开启表格展示时自动设置
  if (!ruleEditMode.value && ruleForm.value.show_in_table && newName) {
    // 磁盘相关规则默认不合并
    ruleForm.value.table_column_merge = !isDiskRelatedRule(newName)
  }
})

// 监听表格展示开关，自动设置合并单元格默认值
watch(() => ruleForm.value.show_in_table, (showInTable) => {
  // 只在新增模式时自动设置
  if (!ruleEditMode.value && showInTable && ruleForm.value.name) {
    ruleForm.value.table_column_merge = !isDiskRelatedRule(ruleForm.value.name)
  }
})

onMounted(() => {
  loadData()
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

.group-header {
  display: flex;
  align-items: center;
  gap: 20px;
}

.group-name {
  font-weight: bold;
  font-size: 15px;
}

.group-desc {
  color: #666;
  font-size: 13px;
}
</style>
