#!/usr/bin/env bash

set -euo pipefail

PHASE="${1:-unit}"

echo "=> Selected test phase: ${PHASE}"

case "${PHASE}" in
  unit)
    make test TEST_PHASE=unit
    ;;
  integration)
    make test TEST_PHASE=integration
    ;;
  e2e)
    make test TEST_PHASE=e2e
    ;;
  allure)
    make test TEST_PHASE=allure
    ;;
  *)
    echo "Unknown test phase: ${PHASE}" >&2
    exit 1
    ;;
esac
