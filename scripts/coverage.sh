#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

COVERAGE_FILE="coverage.out"
MIN_COVERAGE=60

go test -coverprofile="$COVERAGE_FILE" ./...

coverage=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')

echo "Total coverage: ${coverage}%"

if (( $(echo "$coverage < $MIN_COVERAGE" | bc -l) )); then
    echo "FAIL: coverage ${coverage}% is below minimum ${MIN_COVERAGE}%"
    exit 1
fi

echo "PASS: coverage ${coverage}% meets minimum ${MIN_COVERAGE}%"
