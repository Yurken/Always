#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

AI_DIR="$ROOT_DIR/services/ai-py"
CORE_DIR="$ROOT_DIR/services/core-go"
DESKTOP_DIR="$ROOT_DIR/apps/desktop"

if [ ! -d "$AI_DIR/.venv" ]; then
  python3 -m venv "$AI_DIR/.venv"
fi

source "$AI_DIR/.venv/bin/activate"
pip install -r "$AI_DIR/requirements.txt" >/dev/null

deactivate

if [ ! -d "$DESKTOP_DIR/node_modules" ]; then
  (cd "$DESKTOP_DIR" && npm install)
fi

export CORE_PORT=${CORE_PORT:-8081}
export AI_URL=${AI_URL:-http://127.0.0.1:8788}

(
  cd "$AI_DIR"
  source .venv/bin/activate
  uvicorn main:app --host 127.0.0.1 --port 8788
) &
AI_PID=$!

(
  cd "$CORE_DIR"
  go run main.go
) &
CORE_PID=$!

(
  cd "$DESKTOP_DIR"
  npm run dev
) &
DESKTOP_PID=$!

cleanup() {
  kill "$AI_PID" "$CORE_PID" "$DESKTOP_PID" 2>/dev/null || true
}
trap cleanup EXIT

wait
