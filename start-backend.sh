#!/bin/bash

# 后端启动脚本（单一入口，避免旧进程/旧二进制问题）
# 用法: ./start-backend.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR"
PORT="${PORT:-8081}"
RUNTIME_DIR="$PROJECT_DIR/.dev-runtime"
PID_FILE="$RUNTIME_DIR/backend.pid"
LOG_FILE="$RUNTIME_DIR/backend.log"
BINARY_FILE="$PROJECT_DIR/bin/server"

# 环境变量（可被外部覆盖）
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-3306}"
export DB_USER="${DB_USER:-root}"
export DB_PASSWORD="${DB_PASSWORD:-silenceopr@2026}"
export DB_NAME="${DB_NAME:-devops_first}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

mkdir -p "$RUNTIME_DIR"

echo -e "${YELLOW}Stopping old backend process (if any)...${NC}"
if [[ -f "$PID_FILE" ]]; then
    old_pid="$(cat "$PID_FILE" || true)"
    if [[ -n "${old_pid}" ]] && kill -0 "$old_pid" 2>/dev/null; then
        kill "$old_pid" 2>/dev/null || true
        sleep 1
        kill -9 "$old_pid" 2>/dev/null || true
    fi
    rm -f "$PID_FILE"
fi

# Double safety: kill any process still listening on target port.
lsof -ti tcp:"$PORT" | xargs kill -9 2>/dev/null || true
sleep 1

echo -e "${YELLOW}Building backend binary...${NC}"
cd "$PROJECT_DIR"
go build -o "$BINARY_FILE" ./cmd/server

echo -e "${YELLOW}Starting backend on :$PORT ...${NC}"
nohup "$BINARY_FILE" > "$LOG_FILE" 2>&1 &
new_pid=$!
echo "$new_pid" > "$PID_FILE"

# Wait a short time for startup.
for _ in 1 2 3 4 5; do
    if lsof -nP -iTCP:"$PORT" -sTCP:LISTEN >/dev/null 2>&1; then
        echo -e "${GREEN}Backend started successfully.${NC}"
        echo "PID: $new_pid"
        echo "Log: $LOG_FILE"
        exit 0
    fi
    sleep 1
done

echo -e "${RED}Backend failed to listen on port $PORT. Last log lines:${NC}"
tail -n 50 "$LOG_FILE" || true
exit 1
