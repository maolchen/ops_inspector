# 动态表格设计方案

## 1. 需求背景

### 1.1 当前问题
- K8S集群状态组、进程指标组、其他指标组的表格展示是**写死**的
- 基础资源组中 `show_in_table=false` 的规则也需要独立展示
- 新增规则无法自动展示，需要修改代码

### 1.2 目标
- **一条规则一张表**：每个规则独立展示为一张表格
- **动态表头**：基于规则配置的标签别名自动生成表头
- **自动增删**：新增规则自动展示，删除规则自动隐藏

## 2. 数据模型

### 2.1 规则配置示例
```json
{
  "rule_name": "节点就绪状态",
  "group_name": "K8S集群状态",
  "type": true,  // 告警类型，显示背景色
  "labels": "{\"node\":\"节点\",\"condition\":\"状态类型\"}",
  "threshold": 0,
  "threshold_type": "equal",
  "unit": "",
  "show_in_table": false
}
```

### 2.2 标签别名映射
- `labels` 字段是一个 JSON 对象
- **key**：Prometheus 返回的标签名（如 `node`）
- **value**：表头显示名称（如 `节点`）

示例：
| labels配置 | 表头列 |
|-----------|--------|
| `{"node":"节点","condition":"状态类型"}` | 节点、状态类型 |
| `{"namespace":"命名空间","pod":"Pod名称"}` | 命名空间、Pod名称 |
| `{"instance":"节点"}` | 节点 |

### 2.3 巡检项数据结构
```typescript
interface InspectionItem {
  id: number
  rule_id: number
  rule_name: string
  group_name: string
  instance: string
  value: number
  status: string  // normal, warning, critical
  labels: string  // JSON字符串
  unit: string
  type: boolean   // 规则类型：true=告警，false=展示
  show_in_table: boolean
  // ...
}
```

## 3. 表格结构设计

### 3.1 表头组成
每张动态表格包含以下列：

1. **标签列（动态）**：从 `labels` 配置解析
   - 按 labels 中 key 的顺序显示
   - 值从 Prometheus 返回数据的对应标签中提取

2. **值列（固定）**：
   - 列名：固定为"值"
   - 值：Prometheus 查询返回的 value
   - 单位：附加在值后面（如 `80%`、`100MB/s`）

3. **状态列（固定）**：
   - 列名：固定为"状态"
   - 值：根据阈值比较结果，显示"正常"/"异常"

### 3.2 表格示例

**规则：节点就绪状态**
```
labels: {"node":"节点","condition":"状态类型"}
threshold: 0, threshold_type: "equal"
```

| 节点 | 状态类型 | 值 | 状态 |
|-----|---------|---|------|
| node-01 | Ready | 0 | 正常 |
| node-02 | Ready | 1 | 异常 |

**规则：PVC使用率**
```
labels: {"namespace":"命名空间","persistentvolumeclaim":"PVC名称"}
threshold: 90, threshold_type: "greater"
unit: "%"
```

| 命名空间 | PVC名称 | 值 | 状态 |
|---------|--------|---|------|
| default | data-pvc | 85% | 正常 |
| prod | log-pvc | 95% | 异常 |

**规则：CPU使用率（基础资源组，show_in_table=false）**
```
labels: {"instance":"节点"}
threshold: 80, threshold_type: "greater"
unit: "%"
```

| 节点 | 值 | 状态 |
|-----|---|------|
| 192.168.1.1 | 45% | 正常 |
| 192.168.1.2 | 85% | 异常 |

## 4. 页面布局设计

### 4.1 分组展示逻辑

```
巡检报告详情页
├── 巡检总览（保持不变）
├── 资源趋势（保持不变）
├── 基础资源详情表（show_in_table=true 的规则，保持不变）
├── 其他规则表格组（动态生成）
│   ├── 基础资源组（show_in_table=false 的规则）
│   │   ├── 规则A表格
│   │   ├── 规则B表格
│   │   └── ...
│   ├── K8S集群状态组
│   │   ├── 节点就绪状态表格
│   │   ├── 证书状态表格
│   │   └── ...
│   ├── 进程指标组
│   │   └── ...
│   └── 其他指标组
│       └── ...
```

### 4.2 多列布局策略

**规则**：
- 每张表格的列数 ≤ 3 时：使用三列布局（一行显示3张表）
- 每张表格的列数 > 3 时：使用全宽布局（一行显示1张表）

**实现方式**：
```html
<div class="dynamic-tables-container">
  <!-- 三列布局组 -->
  <div class="triple-column-layout" v-if="tripleColumnRules.length > 0">
    <el-card v-for="rule in tripleColumnRules" class="third-width">
      <!-- 表格内容 -->
    </el-card>
  </div>
  
  <!-- 全宽布局组 -->
  <el-card v-for="rule in fullWidthRules" class="full-width">
    <!-- 表格内容 -->
  </el-card>
</div>
```

**CSS**：
```css
.triple-column-layout {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}
.triple-column-layout .el-card {
  flex: 0 0 calc(33.33% - 11px);
  min-width: 300px; /* 最小宽度保证可读性 */
}
```

## 5. 核心算法

### 5.1 数据分组
```typescript
// 获取需要动态展示表格的规则
const dynamicTableRules = computed(() => {
  // 条件：show_in_table=false 或者 不是基础资源组
  return items.value.filter(item => 
    !item.show_in_table || !isBasicResourceGroup(item.group_name)
  )
})

// 按规则分组
const rulesGrouped = computed(() => {
  const groupMap = new Map<string, InspectionItem[]>()
  
  dynamicTableRules.value.forEach(item => {
    const key = `${item.group_name}__${item.rule_name}`
    if (!groupMap.has(key)) {
      groupMap.set(key, [])
    }
    groupMap.get(key)!.push(item)
  })
  
  return groupMap
})
```

### 5.2 表头解析
```typescript
interface DynamicColumn {
  prop: string      // 数据字段名（标签key）
  label: string     // 表头显示名（标签别名）
  isLabel: boolean  // 是否为标签列
}

interface TableConfig {
  ruleName: string
  groupName: string
  columns: DynamicColumn[]
  data: any[]
  ruleType: boolean  // 是否告警类型
}

function parseTableConfig(ruleName: string, groupName: string, items: InspectionItem[]): TableConfig {
  if (items.length === 0) return null
  
  const firstItem = items[0]
  const labels = JSON.parse(firstItem.labels || '{}')
  
  // 解析标签列为动态列
  const columns: DynamicColumn[] = Object.entries(labels).map(([key, alias]) => ({
    prop: key,
    label: alias as string,
    isLabel: true
  }))
  
  // 添加值列和状态列
  columns.push(
    { prop: 'value', label: '值', isLabel: false },
    { prop: 'status', label: '状态', isLabel: false }
  )
  
  // 构建表格数据
  const data = items.map(item => {
    const itemLabels = JSON.parse(item.labels || '{}')
    return {
      ...itemLabels,  // 展开标签
      value: item.value,
      status: item.status,
      _raw: item  // 保留原始数据
    }
  })
  
  return {
    ruleName,
    groupName,
    columns,
    data,
    ruleType: firstItem.type
  }
}
```

### 5.3 状态背景色逻辑
```typescript
function getCellClass(row: any, column: DynamicColumn, ruleType: boolean) {
  // 只对状态列显示背景色
  if (column.prop !== 'status') return ''
  
  // 如果是展示类型（非告警），不显示背景色
  if (!ruleType) return ''
  
  // 根据状态返回对应样式
  if (row.status === 'critical') return 'cell-critical'
  if (row.status === 'warning') return 'cell-warning'
  return 'cell-normal'
}
```

## 6. UI 设计

### 6.1 表格卡片结构
```
┌─────────────────────────────────────────┐
│ 规则名称                    [所属规则组] │
├─────────────────────────────────────────┤
│ 标签列1 | 标签列2 | ... | 值 | 状态     │
│ --------------------------------------- │
│ 数据行...                               │
├─────────────────────────────────────────┤
│ 说明：阈值比较规则描述                   │
└─────────────────────────────────────────┘
```

### 6.2 分组标题
- 同一规则组的表格归为一组
- 分组标题显示规则组名称
- 表格按规则名称排序

## 7. 边界情况处理

### 7.1 空数据处理
- 如果规则查询结果为空，不显示该表格

### 7.2 标签缺失处理
- 如果某条数据的标签缺失，显示 `-`

### 7.3 单位显示
- 值列附加单位显示（如 `80%`、`100MB/s`）
- 如果单位为空，仅显示数值

### 7.4 阈值说明
- 在表格底部显示说明文字
- 格式：`说明：值 {比较符} {阈值} 表示异常`

## 8. 实现步骤

### 8.1 后端修改
1. ✅ 已完成：InspectionItem 包含 `table_column_rule_type` 字段

### 8.2 前端修改
1. 移除写死的 K8S、进程、其他指标表格组件
2. 实现动态表格配置解析函数
3. 实现动态表格组件
4. 实现多列布局逻辑
5. 实现单元格样式计算

### 8.3 测试验证
1. 新增规则后自动展示表格
2. 删除规则后自动隐藏表格
3. 背景色正确显示
4. 多列布局正确响应

## 9. 示例效果

### 三列布局效果
```
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ 节点就绪状态 │ │ Kubelet证书 │ │ Kubeproxy   │
├─────────────┤ ├─────────────┤ ├─────────────┤
│ ...表格...  │ │ ...表格...  │ │ ...表格...  │
└─────────────┘ └─────────────┘ └─────────────┘
```

### 全宽布局效果
```
┌─────────────────────────────────────────────────┐
│ Pod运行状态                         [K8S集群状态] │
├─────────────────────────────────────────────────┤
│ 命名空间 | Pod名称 | 值 | 状态                   │
│ default  | web-pod  | 1  | 正常                  │
│ prod     | api-pod  | 0  | 异常                  │
├─────────────────────────────────────────────────┤
│ 说明：值=1表示正常，值≠1表示异常                  │
└─────────────────────────────────────────────────┘
```

---

## 确认事项

请确认以下设计是否符合预期：

1. **表头设计**：标签列(动态) + 值列(固定) + 状态列(固定) ✅/❌
2. **布局策略**：列数≤3 三列布局，列数>3 全宽布局 ✅/❌
3. **背景色规则**：仅告警类型显示背景色，展示类型不显示 ✅/❌
4. **分组展示**：按规则组分组，组内按规则名称排序 ✅/❌
5. **说明文字**：表格底部显示阈值比较规则 ✅/❌

如有调整，请告知。
