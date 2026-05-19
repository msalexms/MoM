#!/bin/bash
set -euo pipefail

echo "=== Smoke Test ==="

echo "[1/4] Running go vet..."
go vet ./...
echo "  PASS"

echo "[2/4] Running go build..."
go build ./...
echo "  PASS"

echo "[3/4] Running go test..."
go test ./... -count=1
echo "  PASS"

echo "[4/4] Checking binary version..."
make build > /dev/null 2>&1
OUTPUT=$(./bin/mom --version)
if echo "${OUTPUT}" | grep -q "mom v"; then
    echo "  PASS: ${OUTPUT}"
else
    echo "  FAIL: unexpected output: ${OUTPUT}"
    exit 1
fi

echo ""
echo "=== All smoke tests passed ==="
