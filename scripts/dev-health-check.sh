#!/usr/bin/env bash
set -uo pipefail

# DevOps-first local service health checker.
# Default: check-only mode.
# --ensure: start missing services and wait until healthy.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$ROOT_DIR/web"
RUNTIME_DIR="$ROOT_DIR/.dev-runtime"

BACKEND_HOST="${BACKEND_HOST:-127.0.0.1}"
BACKEND_PORT="${BACKEND_PORT:-8081}"
FRONTEND_HOST="${FRONTEND_HOST:-127.0.0.1}"
FRONTEND_PORT="${FRONTEND_PORT:-5173}"

BACKEND_URL="http://${BACKEND_HOST}:${BACKEND_PORT}"
FRONTEND_URL="http://${FRONTEND_HOST}:${FRONTEND_PORT}"

BACKEND_LOG="$RUNTIME_DIR/backend.log"
FRONTEND_LOG="$RUNTIME_DIR/frontend.log"
BACKEND_PID_FILE="$RUNTIME_DIR/backend.pid"
FRONTEND_PID_FILE="$RUNTIME_DIR/frontend.pid"

MODE="check"
if [[ "${1:-}" == "--ensure" ]]; then
  MODE="ensure"
fi

mkdir -p "$RUNTIME_DIR"

is_backend_healthy() {
  curl -fsS --max-time 2 "$BACKEND_URL/healthz" >/dev/null 2>&1
}

is_frontend_healthy() {
  curl -fsS --max-time 2 "$FRONTEND_URL" >/dev/null 2>&1
}

wait_until_healthy() {
  local target="$1"
  local retries="$2"
  local sleep_secs="$3"

  local i=1
  while [[ $i -le $retries ]]; do
    if [[ "$target" == "backend" ]] && is_backend_healthy; then
      return 0
    fi
    if [[ "$target" == "frontend" ]] && is_frontend_healthy; then
      return 0
    fi
    sleep "$sleep_secs"
    i=$((i + 1))
  done
  return 1
}

start_backend() {
  if is_backend_healthy; then
    return 0
  fi

  (
    cd "$ROOT_DIR" || exit 1
    nohup go run ./cmd/server >"$BACKEND_LOG" 2>&1 &
    echo $! >"$BACKEND_PID_FILE"
  )
}

start_frontend() {
  if is_frontend_healthy; then
    return 0
  fi

  (
    cd "$WEB_DIR" || exit 1
    nohup npm run dev -- --host "$FRONTEND_HOST" --port "$FRONTEND_PORT" >"$FRONTEND_LOG" 2>&1 &
    echo $! >"$FRONTEND_PID_FILE"
  )
}

print_status() {
  local backend_status="DOWN"
  local frontend_status="DOWN"

  if is_backend_healthy; then
    backend_status="UP"
  fi
  if is_frontend_healthy; then
    frontend_status="UP"
  fi

  echo "[dev-health] backend:  $backend_status  ($BACKEND_URL/healthz)"
  echo "[dev-health] frontend: $frontend_status ($FRONTEND_URL)"

  if [[ -f "$BACKEND_PID_FILE" ]]; then
    echo "[dev-health] backend pid:  $(cat "$BACKEND_PID_FILE")"
  fi
  if [[ -f "$FRONTEND_PID_FILE" ]]; then
    echo "[dev-health] frontend pid: $(cat "$FRONTEND_PID_FILE")"
  fi
  echo "[dev-health] logs: $RUNTIME_DIR"
}

if [[ "$MODE" == "ensure" ]]; then
  start_backend
  start_frontend

  wait_until_healthy backend 20 1 || true
  wait_until_healthy frontend 20 1 || true
fi

print_status

if is_backend_healthy && is_frontend_healthy; then
  exit 0
fi

exit 1
