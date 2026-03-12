#!/usr/bin/env bash
set -uo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUNTIME_DIR="$ROOT_DIR/.dev-runtime"
BACKEND_PID_FILE="$RUNTIME_DIR/backend.pid"
FRONTEND_PID_FILE="$RUNTIME_DIR/frontend.pid"

kill_by_pid_file() {
  local name="$1"
  local pid_file="$2"

  if [[ ! -f "$pid_file" ]]; then
    echo "[dev-stop] $name pid file not found: $pid_file"
    return 0
  fi

  local pid
  pid="$(cat "$pid_file" 2>/dev/null || true)"

  if [[ -z "$pid" ]]; then
    echo "[dev-stop] $name pid file empty: $pid_file"
    rm -f "$pid_file"
    return 0
  fi

  if kill -0 "$pid" 2>/dev/null; then
    kill "$pid" 2>/dev/null || true
    sleep 1
    if kill -0 "$pid" 2>/dev/null; then
      kill -9 "$pid" 2>/dev/null || true
    fi
    echo "[dev-stop] stopped $name pid=$pid"
  else
    echo "[dev-stop] $name pid=$pid already not running"
  fi

  rm -f "$pid_file"
}

# Fallback cleanup by ports in case PID files are stale or missing.
kill_by_port() {
  local name="$1"
  local port="$2"

  local pids
  pids="$(lsof -ti tcp:"$port" 2>/dev/null || true)"
  if [[ -n "$pids" ]]; then
    # shellcheck disable=SC2086
    kill $pids 2>/dev/null || true
    sleep 1
    pids="$(lsof -ti tcp:"$port" 2>/dev/null || true)"
    if [[ -n "$pids" ]]; then
      # shellcheck disable=SC2086
      kill -9 $pids 2>/dev/null || true
    fi
    echo "[dev-stop] cleaned $name listeners on port $port"
  fi
}

kill_by_pid_file "backend" "$BACKEND_PID_FILE"
kill_by_pid_file "frontend" "$FRONTEND_PID_FILE"
kill_by_port "backend" 8081
kill_by_port "frontend" 5173

echo "[dev-stop] done"
