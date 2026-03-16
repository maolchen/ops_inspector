#!/bin/bash
set -e

# 设置 Go 环境变量
export GOROOT=/workspace/go
export PATH=$GOROOT/bin:$PATH

echo "Go version: $(go version)"

# 构建前端
cd frontend
pnpm install
pnpm build

# 编译后端
cd ../backend
go build -o ../server ./cmd/main.go

echo "Build completed successfully!"
