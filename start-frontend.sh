#!/bin/bash

# 前端启动脚本
# 用法: ./start-frontend.sh

PROJECT_DIR="/Users/moryzang/GoProjects/DevOps-first/web"
PORT=5173

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}📂 切换到项目目录: $PROJECT_DIR${NC}"
cd "$PROJECT_DIR" || { echo -e "${RED}❌ 无法切换到项目目录${NC}"; exit 1; }

echo -e "${YELLOW}🚀 启动前端开发服务器...${NC}"
echo -e "${YELLOW}   监听: http://127.0.0.1:$PORT${NC}"
echo -e "${YELLOW}   代理 /api → http://localhost:8081${NC}"
echo ""

# 启动vite dev服务器
npm run dev -- --host 127.0.0.1 --port $PORT
