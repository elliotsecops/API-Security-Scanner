#!/usr/bin/env bash
set -euo pipefail

# Simple launcher for API Security Scanner
# - Builds the React GUI (./gui) if not already built
# - Checks and auto-adjusts ports if occupied by generating a temporary config
# - Starts the scanner with or without the dashboard

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/api-security-scanner"
DEFAULT_CONFIG="$SCRIPT_DIR/config.yaml"
USE_DASHBOARD=1
CUSTOM_CONFIG=""
SERVER_PORT_OVERRIDE=""
METRICS_PORT_OVERRIDE=""

usage() {
  cat <<EOF
Usage: $(basename "$0") [--config path] [--no-gui] [--port PORT] [--metrics-port PORT]

Options:
  --config path       Path to YAML config (default: config.yaml)
  --no-gui            Run without dashboard (disables GUI)
  --port PORT         Override server.port for this run
  --metrics-port PORT Override metrics.port (dashboard) for this run
EOF
}

# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --config)
      CUSTOM_CONFIG="$2"; shift 2;;
    --no-gui)
      USE_DASHBOARD=0; shift;;
    --port)
      SERVER_PORT_OVERRIDE="$2"; shift 2;;
    --metrics-port)
      METRICS_PORT_OVERRIDE="$2"; shift 2;;
    -h|--help)
      usage; exit 0;;
    *)
      echo "Unknown option: $1" >&2; usage; exit 1;;
  esac
done

CONFIG_PATH="${CUSTOM_CONFIG:-$DEFAULT_CONFIG}"

if [[ ! -x "$BIN" ]]; then
  echo "Error: backend binary not found at $BIN" >&2
  echo "Please build or place the api-security-scanner binary in the repo root." >&2
  exit 1
fi

if [[ ! -f "$CONFIG_PATH" ]]; then
  echo "Error: config file not found at $CONFIG_PATH" >&2
  exit 1
fi

# Build GUI if dashboard enabled and build is missing
if [[ "$USE_DASHBOARD" -eq 1 ]]; then
  if [[ ! -f "$SCRIPT_DIR/gui/build/index.html" ]]; then
    echo "GUI build not found. Attempting to build the GUI..."
    if command -v npm >/dev/null 2>&1; then
      # Prefer ci if lockfile exists
      if [[ -f "$SCRIPT_DIR/gui/package-lock.json" ]]; then
        npm --prefix "$SCRIPT_DIR/gui" ci || npm --prefix "$SCRIPT_DIR/gui" install
      else
        npm --prefix "$SCRIPT_DIR/gui" install
      fi
      npm --prefix "$SCRIPT_DIR/gui" run build
      echo "GUI build completed."
    else
      echo "npm not found. Proceeding without a prebuilt GUI (legacy dashboard)."
    fi
  fi
fi

# Helpers
is_port_busy() {
  local port="$1"
  ss -ltn 2>/dev/null | grep -q ":$port "
}

pick_free_port() {
  local port="$1"
  local max_tries=50
  local tries=0
  while is_port_busy "$port"; do
    port=$((port+1))
    tries=$((tries+1))
    if (( tries > max_tries )); then
      echo ""; return 1
    fi
  done
  echo "$port"
}

# Extract current ports from config (best-effort)
extract_block_port() {
  local block_name="$1"; shift
  awk -v blk="^"$block_name":$" '
    BEGIN{inblk=0}
    $0 ~ blk {inblk=1; next}
    inblk && $0 ~ /^[^[:space:]]/ {inblk=0}
    inblk && $1 ~ /^port:/ {print $2; exit}
  ' "$CONFIG_PATH" 2>/dev/null || true
}

SERVER_PORT_CUR="$(extract_block_port server)"
METRICS_PORT_CUR="$(extract_block_port metrics)"

SERVER_PORT_RUN="${SERVER_PORT_OVERRIDE:-$SERVER_PORT_CUR}"
METRICS_PORT_RUN="${METRICS_PORT_OVERRIDE:-$METRICS_PORT_CUR}"

# Fallback defaults if not discovered
SERVER_PORT_RUN="${SERVER_PORT_RUN:-8081}"
METRICS_PORT_RUN="${METRICS_PORT_RUN:-8090}"

# Auto-adjust if busy
ADJUSTED=0
if is_port_busy "$SERVER_PORT_RUN"; then
  NEW_SERVER_PORT="$(pick_free_port "$SERVER_PORT_RUN")"
  if [[ -n "$NEW_SERVER_PORT" ]]; then
    SERVER_PORT_RUN="$NEW_SERVER_PORT"; ADJUSTED=1
    echo "Notice: server port busy, using $SERVER_PORT_RUN instead"
  fi
fi

if [[ "$USE_DASHBOARD" -eq 1 ]]; then
  if is_port_busy "$METRICS_PORT_RUN"; then
    NEW_METRICS_PORT="$(pick_free_port "$METRICS_PORT_RUN")"
    if [[ -n "$NEW_METRICS_PORT" ]]; then
      METRICS_PORT_RUN="$NEW_METRICS_PORT"; ADJUSTED=1
      echo "Notice: dashboard/metrics port busy, using $METRICS_PORT_RUN instead"
    fi
  fi
fi

RUN_CONFIG="$CONFIG_PATH"
TMP_CFG=""
if [[ "$ADJUSTED" -eq 1 || -n "$SERVER_PORT_OVERRIDE" || -n "$METRICS_PORT_OVERRIDE" ]]; then
  TMP_CFG="$(mktemp)"
  awk -v mport="$METRICS_PORT_RUN" -v sport="$SERVER_PORT_RUN" '
    BEGIN{in_metrics=0; in_server=0; m_done=0; s_done=0}
    /^metrics:/ {in_metrics=1; in_server=0}
    /^server:/ {in_server=1; in_metrics=0}
    in_metrics && !m_done && $1 ~ /^port:/ { sub(/:[[:space:]]*[0-9]+/, ": " mport); m_done=1 }
    in_server && !s_done && $1 ~ /^port:/ { sub(/:[[:space:]]*[0-9]+/, ": " sport); s_done=1 }
    {print}
  ' "$CONFIG_PATH" > "$TMP_CFG"
  RUN_CONFIG="$TMP_CFG"
fi

CLEANUP() {
  if [[ -n "$TMP_CFG" && -f "$TMP_CFG" ]]; then rm -f "$TMP_CFG"; fi
}
trap CLEANUP EXIT

CMD=("$BIN" -config "$RUN_CONFIG")
if [[ "$USE_DASHBOARD" -eq 1 ]]; then
  CMD+=( -dashboard )
fi

echo "Starting API Security Scanner..."
echo " - Config: $RUN_CONFIG"
if [[ "$USE_DASHBOARD" -eq 1 ]]; then
  echo " - Dashboard URL: http://localhost:$METRICS_PORT_RUN"
fi

exec "${CMD[@]}"
