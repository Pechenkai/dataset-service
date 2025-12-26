#!/usr/bin/env bash

set -euo pipefail

API_BASE=${API_BASE:-http://localhost:8080}
INTERFACE=${INTERFACE:-lo}
API_PORT=${API_PORT:-8080}
USER_ID=${USER_ID:-}
DATASET_ID=${DATASET_ID:-}
CATEGORY_ID=${CATEGORY_ID:-}
LOG_FILE=${LOG_FILE:-logs/e2e_capture_example.txt}
PCAP_FILE=${PCAP_FILE:-}
CAPTURE_TOOL=${CAPTURE_TOOL:-tcpdump}
USE_SUDO=${USE_SUDO:-1}
SEED_BIN=${SEED_BIN:-build/bin/e2e_seed}
SEED_FORCE_REBUILD=${SEED_FORCE_REBUILD:-0}

mkdir -p "$(dirname "$LOG_FILE")"

CAPTURE_NEEDS_SUDO=0
CAPTURE_FILE=""
TMP_CAPTURE=""
CAPTURE_MODE="ascii"
FORCE_SHUTDOWN=0

ensure_dir() {
  local dir="$1"
  if [[ -n "$dir" && ! -d "$dir" ]]; then
    mkdir -p "$dir"
  fi
}

if [[ -n "$PCAP_FILE" ]]; then
  mkdir -p "$(dirname "$PCAP_FILE")"
  CAPTURE_FILE="$PCAP_FILE"
  CAPTURE_MODE="pcap"
  CAP_CMD=("$CAPTURE_TOOL" "-i" "$INTERFACE" "-s" "0" "-w" "$CAPTURE_FILE" "tcp port $API_PORT")
else
  TMP_CAPTURE="$(mktemp)"
  CAPTURE_FILE="$TMP_CAPTURE"
  CAP_CMD=("$CAPTURE_TOOL" "-i" "$INTERFACE" "-s" "0" "-A" "-nn" "tcp port $API_PORT")
fi

if [[ $USE_SUDO -eq 1 && $EUID -ne 0 ]]; then
  CAPTURE_NEEDS_SUDO=1
  echo "=> Elevating privileges for capture tool (sudo -v)"
  sudo -v
  CAP_CMD=("sudo" "${CAP_CMD[@]}")
fi

cleanup_existing_capture() {
  local pattern="$CAPTURE_TOOL -i $INTERFACE -s 0"
  if ! pgrep -f "$pattern" >/dev/null 2>&1; then
    return
  fi
  echo "=> Cleaning previous capture processes ($pattern)"
  local pkill_cmd=(pkill -f "$pattern")
  if [[ $CAPTURE_NEEDS_SUDO -eq 1 ]]; then
    pkill_cmd=(sudo "${pkill_cmd[@]}")
  fi
  "${pkill_cmd[@]}" &>/dev/null || true
  sleep 1
  if pgrep -f "$pattern" >/dev/null 2>&1; then
    echo "!! Previous capture processes are still running. Please terminate them manually: $pattern" >&2
  fi
}

ensure_seed_binary() {
  if [[ $SEED_FORCE_REBUILD -eq 0 && -x "$SEED_BIN" ]]; then
    return
  fi
  echo "=> Building e2e seed binary ($SEED_BIN)"
  ensure_dir "$(dirname "$SEED_BIN")"
  go build -o "$SEED_BIN" ./cmd/e2e_seed
}

collect_children() {
  local pid="$1"
  local children_file="/proc/${pid}/task/${pid}/children"
  if [[ ! -r "$children_file" ]]; then
    return
  fi
  local child
  while read -r child; do
    [[ -z "$child" ]] && continue
    echo "$child"
    collect_children "$child"
  done < <(tr ' ' '\n' <"$children_file")
}

stop_capture() {
  if [[ -z "${CAP_PID:-}" ]]; then
    return
  fi
  local signal_cmd=(kill)
  if [[ $CAPTURE_NEEDS_SUDO -eq 1 ]]; then
    signal_cmd=(sudo kill)
  fi
  local -a descendants=()
  if [[ -r "/proc/${CAP_PID}/task/${CAP_PID}/children" ]]; then
    mapfile -t descendants < <(collect_children "$CAP_PID" 2>/dev/null || true)
  fi
  # Kill children first to avoid orphaned sudo/tcpdump processes.
  for ((idx=${#descendants[@]}-1; idx>=0; idx--)); do
    local child_pid="${descendants[idx]}"
    [[ -z "$child_pid" ]] && continue
    if "${signal_cmd[@]}" -0 "$child_pid" &>/dev/null; then
      "${signal_cmd[@]}" "$child_pid" &>/dev/null || true
    fi
  done
  if "${signal_cmd[@]}" -0 "$CAP_PID" &>/dev/null; then
    "${signal_cmd[@]}" "$CAP_PID" &>/dev/null || true
    wait "$CAP_PID" 2>/dev/null || true
  fi
  sleep 1
  local pattern="$CAPTURE_TOOL -i $INTERFACE -s 0"
  if pgrep -f "$pattern" >/dev/null 2>&1; then
    echo "=> Forcing capture tool shutdown for pattern: $pattern"
    FORCE_SHUTDOWN=1
    local pkill_cmd=(pkill -f "$pattern")
    if [[ $CAPTURE_NEEDS_SUDO -eq 1 ]]; then
      pkill_cmd=(sudo "${pkill_cmd[@]}")
    fi
    "${pkill_cmd[@]}" &>/dev/null || true
  fi
  unset CAP_PID
}
finish() {
  stop_capture
  if [[ -n "$TMP_CAPTURE" && -f "$TMP_CAPTURE" ]]; then
    rm -f "$TMP_CAPTURE"
  fi
}
trap finish EXIT

ensure_seed_binary

echo "=> Seeding database via ${SEED_BIN}"
SEED_OUTPUT=$("$SEED_BIN")
if [[ -z "${SEED_OUTPUT}" ]]; then
  echo "Seeding command returned no data" >&2
  exit 1
fi
eval "${SEED_OUTPUT}"
if [[ -z "${USER_ID}" || -z "${DATASET_ID}" ]]; then
  echo "Seed command did not return USER_ID or DATASET_ID" >&2
  exit 1
fi
echo "   -> USER_ID=${USER_ID} DATASET_ID=${DATASET_ID} CATEGORY_ID=${CATEGORY_ID}"

cleanup_existing_capture

echo "=> Starting traffic capture (${CAP_CMD[*]})"
if [[ "$CAPTURE_MODE" == "pcap" ]]; then
  "${CAP_CMD[@]}" &
else
  "${CAP_CMD[@]}" >"$TMP_CAPTURE" &
fi
CAP_PID=$!
sleep 1

api_get() {
  local path="$1"
  echo "=> GET ${API_BASE}${path}"
  curl -fsS -X GET \
    -H "Accept: application/json" \
    "${API_BASE}${path}" >/dev/null
}

api_post() {
  local path="$1"
  local payload="$2"
  echo "=> POST ${API_BASE}${path}"
  curl -fsS -X POST \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "${API_BASE}${path}" >/dev/null
}

api_get "/categories"
api_post "/reviews" "{\"user_id\":${USER_ID},\"dataset_id\":${DATASET_ID},\"rating\":5,\"text\":\"Great dataset for prototyping models\"}"
api_get "/datasets?public=true"
api_get "/datasets/${DATASET_ID}/versions"
api_post "/subscriptions" "{\"user_id\":${USER_ID},\"dataset_id\":${DATASET_ID}}"

echo "=> Stopping capture"
stop_capture
echo "=> Capture stopped"
trap - EXIT

if [[ $FORCE_SHUTDOWN -eq 1 ]]; then
  echo "Capture tool required forceful shutdown; PCAP may be incomplete. Please rerun once tcpdump is fully stopped." >&2
  exit 1
fi

if [[ "$CAPTURE_MODE" == "pcap" ]]; then
  tmp_ascii="$(mktemp)"
  tcpdump -nn -A -r "$CAPTURE_FILE" >"$tmp_ascii"
  {
    echo "# Capture toolchain: ${CAPTURE_TOOL} -> ${CAPTURE_FILE}"
    echo "# Scenario: public dataset journey (automated $(date -Iseconds))"
    echo
    cat "$tmp_ascii"
  } >"$LOG_FILE"
  rm -f "$tmp_ascii"
  echo "=> Capture saved to ${LOG_FILE} and PCAP ${CAPTURE_FILE}"
else
  {
    echo "# Capture toolchain: ${CAPTURE_TOOL}"
    echo "# Scenario: public dataset journey (automated $(date -Iseconds))"
    echo
    cat "$CAPTURE_FILE"
  } >"$LOG_FILE"
  rm -f "$CAPTURE_FILE"
  echo "=> Capture saved to ${LOG_FILE}"
fi
