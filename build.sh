#!/bin/bash
set -e

# 构建前端
cd frontend
pnpm install
pnpm build

# 编译后端
cd ../backend
go build -o ../server ./cmd/main.go
