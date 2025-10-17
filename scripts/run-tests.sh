#!/usr/bin/env bash
set -euo pipefail

PHASE="${1:-unit}"

export GOCACHE="${GOCACHE:-$(pwd)/.gocache}"
mkdir -p "${GOCACHE}"

run_unit() {
  go test ./...
}

run_integration() {
  go test -tags=integration ./internal/tests/access_tests/... ./internal/tests/integration_tests
}

run_e2e() {
  go test -tags=e2e ./internal/tests/e2e
}

run_allure() {
  make test-dataset-allure
}

case "${PHASE}" in
  unit)
    run_unit
    ;;
  integration)
    run_integration
    ;;
  e2e)
    run_e2e
    ;;
  allure)
    run_allure
    ;;
  *)
    echo "Unknown test phase: ${PHASE}" >&2
    exit 1
    ;;
esac
