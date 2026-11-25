BINARY  := mom
BIN_DIR := bin
PKG     := github.com/msalexms/MoM/internal/version
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(PKG).Version=$(VERSION) -X $(PKG).Commit=$(COMMIT) -X $(PKG).Date=$(DATE)

.PHONY: build build-all test lint fmt run clean release

build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) ./cmd/mom

build-all:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-linux-amd64 ./cmd/mom
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-linux-arm64 ./cmd/mom
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY)-linux-armv7 ./cmd/mom
	cd $(BIN_DIR) && sha256sum $(BINARY)-linux-* > checksums.txt

test:
	go test ./... -race -count=1

lint:
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@2025.1.1 ./...

fmt:
	gofmt -w .

run:
	go run ./cmd/mom

clean:
	rm -rf $(BIN_DIR)

release:
	goreleaser release --clean
