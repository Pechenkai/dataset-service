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

mkdir -p "$(dirname "$LOG_FILE")"

CAPTURE_FILE=""
TMP_CAPTURE=""
CAPTURE_MODE="ascii"

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
  CAP_CMD=("sudo" "${CAP_CMD[@]}")
fi

stop_capture() {
  if [[ -n "${CAP_PID:-}" ]]; then
    if kill -0 "$CAP_PID" &>/dev/null; then
      kill "$CAP_PID" || true
      wait "$CAP_PID" 2>/dev/null || true
    fi
  fi
}
finish() {
  stop_capture
  if [[ -n "$TMP_CAPTURE" && -f "$TMP_CAPTURE" ]]; then
    rm -f "$TMP_CAPTURE"
  fi
}
trap finish EXIT

echo "=> Seeding database via cmd/e2e_seed"
SEED_OUTPUT=$(GOCACHE=$(mktemp -d) go run ./cmd/e2e_seed)
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

stop_capture
trap - EXIT

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
