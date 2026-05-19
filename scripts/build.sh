#!/bin/bash
set -euo pipefail

BINARY_NAME="mom"
BUILD_DIR="bin"
VERSION="${VERSION:-0.1.0}"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

mkdir -p "${BUILD_DIR}"

echo "Building ${BINARY_NAME} v${VERSION}..."

for ARCH in amd64 arm64; do
    echo "  -> linux/${ARCH}"
    GOOS=linux GOARCH="${ARCH}" go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/${BINARY_NAME}-linux-${ARCH}" ./cmd/mom
done

echo "  -> linux/arm (v7)"
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="${LDFLAGS}" -o "${BUILD_DIR}/${BINARY_NAME}-linux-armv7" ./cmd/mom

echo ""
echo "Generating checksums..."
cd "${BUILD_DIR}"
sha256sum mom-* > checksums.txt
cat checksums.txt
echo ""
echo "Build complete!"
