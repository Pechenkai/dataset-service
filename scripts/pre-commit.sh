#!/usr/bin/env sh
set -e

echo "Running pre-commit checks..."

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "golangci-lint is required. Install via: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" >&2
  exit 1
fi

make lint
