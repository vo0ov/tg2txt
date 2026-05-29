BINARY_NAME ?= tg2txt
DIST_DIR ?= dist
MAIN_PACKAGE ?= ./cmd
MODULE := $(shell go list -m)

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT) -X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)

TARGETS ?= linux/amd64 linux/arm64 linux/arm/7 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 windows/386
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || echo "go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2")
GOSEC ?= $(shell command -v gosec 2>/dev/null || echo "go run github.com/securego/gosec/v2/cmd/gosec@v2.25.0")

.DEFAULT_GOAL := help

.PHONY: help run ci test race vet fmt lint sec compile build release clean prepare

help:
	@echo "✨ tg2txt Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make run       Run tg2txt from source"
	@echo "  make ci        Run CI checks: race, lint, sec, compile"
	@echo "  make test      Run tests"
	@echo "  make race      Run tests with the race detector"
	@echo "  make vet       Run go vet"
	@echo "  make fmt       Format Go code"
	@echo "  make lint      Run golangci-lint"
	@echo "  make sec       Run gosec"
	@echo "  make compile   Compile all packages"
	@echo "  make build     Build dist/$(BINARY_NAME) for this OS"
	@echo "  make release   Build release archives for Linux, macOS, and Windows"
	@echo "  make clean     Remove dist/"

run:
	@go run $(MAIN_PACKAGE)

ci:
	@echo "🚦 Running local CI..."
	@$(MAKE) race
	@$(MAKE) lint
	@$(MAKE) sec
	@$(MAKE) compile
	@echo "✅ CI passed!"

test:
	@echo "🧪 Running tests..."
	@go test -v ./...

race:
	@echo "🧪 Running race detector tests..."
	@CGO_ENABLED=1 go test -v -race ./...

vet:
	@echo "🔎 Running go vet..."
	@go vet ./...

fmt:
	@echo "🎨 Formatting Go code..."
	@gofmt -w $$(find cmd internal -name '*.go')

lint:
	@echo "🔎 Running golangci-lint..."
	@if [ -f .golangci.yml ]; then \
		$(GOLANGCI_LINT) run --timeout=5m --config=.golangci.yml; \
	else \
		$(GOLANGCI_LINT) run --timeout=5m; \
	fi

sec:
	@echo "🔒 Running security checks..."
	@$(GOSEC) ./...

compile:
	@echo "🏗️  Compiling packages..."
	@CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" ./...

build: prepare
	@echo "🏗️  Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✅ Built $(DIST_DIR)/$(BINARY_NAME)"

release: clean prepare
	@echo "🚀 Building release artifacts for $(VERSION)..."
	@set -e; \
	for target in $(TARGETS); do \
		os=$${target%%/*}; \
		rest=$${target#*/}; \
		arch=$${rest%%/*}; \
		arm=""; \
		variant=""; \
		if [ "$$rest" != "$$arch" ]; then \
			arm=$${rest#*/}; \
			variant="_armv$${arm}"; \
		fi; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		asset="$(BINARY_NAME)_$(VERSION)_$${os}_$${arch}$${variant}"; \
		out_dir="$(DIST_DIR)/$$asset"; \
		mkdir -p "$$out_dir"; \
		echo "  → $$asset"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch GOARM=$$arm go build -trimpath -ldflags="$(LDFLAGS)" -o "$$out_dir/$(BINARY_NAME)$$ext" $(MAIN_PACKAGE); \
		cp README.md LICENSE "$$out_dir/"; \
		if [ "$$os" = "windows" ]; then \
			(cd $(DIST_DIR) && zip -qr "$$asset.zip" "$$asset"); \
		else \
			(cd $(DIST_DIR) && tar -czf "$$asset.tar.gz" "$$asset"); \
		fi; \
		rm -rf "$$out_dir"; \
	done
	@cd $(DIST_DIR) && if command -v sha256sum >/dev/null 2>&1; then sha256sum * > SHA256SUMS.txt; else shasum -a 256 * > SHA256SUMS.txt; fi
	@echo "✅ Release artifacts are ready in $(DIST_DIR)/"

clean:
	@rm -rf $(DIST_DIR)
	@echo "✅ Cleaned $(DIST_DIR)/"

prepare:
	@mkdir -p $(DIST_DIR)
