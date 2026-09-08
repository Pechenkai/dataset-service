#!/usr/bin/env bash
set -euo pipefail

TARGET="${TARGET:-http://localhost:8088}"
CONCURRENCY="${CONCURRENCY:-40}"
REQUESTS="${REQUESTS:-400}"
OUT_DIR="${OUT_DIR:-logs/ab}"
AB_IMAGE="${AB_IMAGE:-jordi/ab}"
PATHS_DEFAULT="/api/v1/categories /mirror/api/v1/categories"

IFS=' ' read -r -a PATH_LIST <<< "${PATHS:-${PATHS_DEFAULT}}"
mkdir -p "${OUT_DIR}"

if command -v ab >/dev/null 2>&1; then
  AB_CMD=(ab)
  echo "=> Using local 'ab' binary"
else
  AB_CMD=(docker run --rm --net=host "${AB_IMAGE}")
  echo "=> Local 'ab' not found, using Docker image ${AB_IMAGE}"
fi

run_ab() {
  local path="$1"
  local name="$2"
  local url="${TARGET}${path}"
  echo "=> Hitting ${url} with ${REQUESTS} requests (${CONCURRENCY} concurrent)"
  "${AB_CMD[@]}" -k -c "${CONCURRENCY}" -n "${REQUESTS}" "${url}" | tee "${OUT_DIR}/${name}.txt"
}

for path in "${PATH_LIST[@]}"; do
  clean_name=$(echo "${path}" | sed 's|^/||; s|/|_|g')
  run_ab "${path}" "${clean_name}"
  echo
done

echo "Done. Results saved under ${OUT_DIR}/"
