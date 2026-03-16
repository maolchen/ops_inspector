# 运维巡检平台

一个轻量级的运维巡检平台，用于自动化收集 Prometheus 监控数据并生成可视化的巡检报告。

## 技术栈

### 后端
- **语言**: Go 1.21+
- **框架**: Gin (轻量级 Web 框架)
- **数据库**: SQLite (轻量级嵌入式数据库)
- **监控集成**: Prometheus Client SDK

### 前端
- **框架**: Vue 3
- **UI 组件**: Element Plus
- **图表**: ECharts 5
- **路由**: Vue Router 4
- **状态管理**: Pinia
- **构建工具**: Vite

## 功能特性

### 1. 项目管理
- 支持多项目管理
- 配置项目名称、Prometheus API 地址、认证 Token
- Token 支持脱敏展示 (**********)

### 2. 规则组管理
- 预置规则组：基础资源组、K8S容器组、进程资源组、其他分组
- 支持自定义规则组
- 支持规则组排序

### 3. 巡检规则配置
- **规则字段**:
  - 名称、类型（告警/展示）、是否在表格展示
  - PromQL 即时查询、趋势查询
  - 阈值、单位、标签别名
  - 阈值比较方式：greater、greater_equal、less、less_equal、equal、at_least
  - 适用项目范围

### 4. 巡检执行
- 选择项目执行巡检
- 实时采集 Prometheus 数据
- 自动阈值判断和状态标记

### 5. 巡检报告展示
- **巡检总览**: 按分组展示统计卡片
- **资源趋势图**: CPU、内存、磁盘使用率趋势（最近7天）
- **基础资源详情表**: 服务器基础资源概览
- **分组详情**: 按规则组分类展示详细数据
- **巡检总结**: 支持编辑巡检详情和备注

### 6. 报告导出
- 支持导出为 PDF 格式
- 所见即所得的导出效果

## 目录结构

```
ops-inspection/
├── backend/                    # 后端项目
│   ├── cmd/
│   │   └── main.go            # 程序入口
│   ├── internal/
│   │   ├── config/            # 配置管理
│   │   ├── model/             # 数据模型
│   │   ├── repository/        # 数据访问层
│   │   ├── service/           # 业务逻辑层
│   │   ├── handler/           # HTTP 处理器
│   │   ├── middleware/        # 中间件
│   │   └── router/            # 路由配置
│   ├── pkg/
│   │   └── prometheus/        # Prometheus 客户端
│   ├── config.yaml            # 配置文件
│   └── go.mod
│
├── frontend/                   # 前端项目
│   ├── src/
│   │   ├── api/               # API 接口
│   │   ├── views/             # 页面组件
│   │   ├── router/            # 路由配置
│   │   ├── App.vue            # 根组件
│   │   └── main.ts            # 入口文件
│   ├── package.json
│   └── vite.config.ts
│
├── data/                       # 数据目录
│   └── inspection.db          # SQLite 数据库
│
└── README.md
```

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+
- pnpm 8+

### 后端启动

```bash
# 进入后端目录
cd backend

# 下载依赖
go mod download

# 运行服务
go run cmd/main.go
```

后端服务默认运行在 http://localhost:5001

### 前端启动

```bash
# 进入前端目录
cd frontend

# 安装依赖
pnpm install

# 开发模式运行
pnpm dev
```

前端服务默认运行在 http://localhost:5000

### 生产构建

```bash
# 后端构建
cd backend
go build -o inspection cmd/main.go

# 前端构建
cd frontend
pnpm build
```

## 配置说明

### 后端配置 (backend/config.yaml)

```yaml
server:
  port: 5001           # 服务端口
  mode: debug          # 运行模式: debug, release

database:
  type: sqlite         # 数据库类型
  path: ./data/inspection.db  # 数据库路径

prometheus:
  timeout: 30          # 查询超时时间（秒）
  max_retries: 3       # 最大重试次数

log:
  level: info          # 日志级别
```

## API 接口

### 项目管理
- `GET /api/projects` - 获取项目列表
- `GET /api/projects/:id` - 获取项目详情
- `POST /api/projects` - 创建项目
- `PUT /api/projects/:id` - 更新项目
- `DELETE /api/projects/:id` - 删除项目

### 规则组管理
- `GET /api/rule-groups` - 获取规则组列表
- `POST /api/rule-groups` - 创建规则组
- `PUT /api/rule-groups/:id` - 更新规则组
- `DELETE /api/rule-groups/:id` - 删除规则组

### 规则管理
- `GET /api/rules` - 获取规则列表
- `GET /api/rules/:id` - 获取规则详情
- `POST /api/rules` - 创建规则
- `PUT /api/rules/:id` - 更新规则
- `DELETE /api/rules/:id` - 删除规则
- `POST /api/rules/:id/toggle` - 切换规则启用状态

### 巡检管理
- `POST /api/inspections/start` - 启动巡检
- `GET /api/inspections` - 获取巡检历史
- `GET /api/inspections/:id` - 获取巡检报告详情
- `PUT /api/inspections/:id/summary` - 更新巡检总结

## 预置规则示例

### 基础资源组
- CPU 使用率（告警阈值 > 80%）
- CPU 核心数（仅展示）
- 内存使用率（告警阈值 > 85%）
- 磁盘使用率（告警阈值 > 90%）
- TCP 连接数

### K8S 容器组
- K8S 节点状态（Ready 检测）
- Pod 运行状态
- PVC 使用情况

### 进程资源组
- 进程 CPU 使用 Top10
- 进程内存使用 Top10

## 阈值比较方式说明

| 类型 | 说明 |
|------|------|
| greater | 值 > 阈值 = 严重 (critical) |
| greater_equal | 值 >= 阈值 = 严重 (critical) |
| less | 值 < 阈值 = 正常 (normal) |
| less_equal | 值 <= 阈值 = 正常 (normal) |
| equal | 值 == 阈值 = 正常 (normal) |
| at_least | 值 >= 阈值 = 正常 (normal) |

**告警规则**: 当值超过阈值 90% 时，状态为警告 (warning)

## 开发指南

### 后端开发
- 遵循分层架构：Handler → Service → Repository → Model
- 一个功能一个模块，一个模块可以有多个 .go 文件
- 使用 GORM 进行数据库操作

### 前端开发
- 使用 Vue 3 Composition API
- 遵循组件化开发原则
- 使用 TypeScript 增强类型安全

## 许可证

MIT License
