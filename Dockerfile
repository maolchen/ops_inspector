# ============================================
# 运维巡检平台 Dockerfile
# 多阶段构建：前端 + 后端
# ============================================

# ============================================
# 阶段1: 构建前端
# ============================================
FROM node:20-alpine AS frontend-builder

# 安装 pnpm
RUN npm install -g pnpm

WORKDIR /app/frontend

# 复制前端依赖文件（利用 Docker 缓存）
COPY frontend/package.json frontend/pnpm-lock.yaml* ./

# 安装依赖
RUN pnpm install --frozen-lockfile || pnpm install

# 复制前端源码
COPY frontend/ ./

# 构建前端
RUN pnpm build

# ============================================
# 阶段2: 构建后端
# ============================================
FROM golang:1.21-alpine AS backend-builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app/backend

# 复制 go.mod 和 go.sum 用于缓存依赖
COPY backend/go.mod backend/go.sum* ./

# 下载依赖
RUN go mod download

# 复制后端源码
COPY backend/ ./

# 构建后端（纯 Go，无需 CGO）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/main.go

# ============================================
# 阶段3: 运行环境
# ============================================
FROM alpine:3.19

LABEL maintainer="Ops Inspection Platform"
LABEL description="轻量级运维巡检平台"

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 设置端口（可通过环境变量覆盖）
ENV PORT=5000

# 创建非 root 用户
RUN adduser -D -u 1000 appuser

WORKDIR /app/backend

# 从构建阶段复制产物
COPY --from=backend-builder /app/server ./
COPY backend/config.yaml ./

# 复制前端静态文件
COPY --from=frontend-builder /app/frontend/dist ../frontend/dist

# 创建数据目录并设置权限
RUN mkdir -p /app/backend/data && \
    chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 5000

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5000/ || exit 1

# 启动服务
CMD ["./server"]
