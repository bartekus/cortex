# SPDX-License-Identifier: AGPL-3.0-or-later
# Cortex Makefile

SHELL := /bin/bash

# Tools versions (overridable)
GOFUMPT_VERSION := v0.6.0
GOIMPORTS_VERSION := v0.27.0
GOLANGCI_LINT_VERSION := v1.63.4
GOLANGCI_LINT_VERSION := v1.63.4
ADDLICENSE_VERSION := v1.1.1

.PHONY: all build test lint fmt-check go-build go-test go-lint go-fmt-check rust-build rust-test rust-lint rust-fmt-check

all: lint test build

# --- Top-level targets (canonical) ---

build: go-build rust-build

test: go-test rust-test

lint: go-lint rust-lint

fmt-check: go-fmt-check rust-fmt-check

# --- Go (repo root) ---

go-build:
	go build -o ./bin/cortex ./cmd/cortex

go-test:
	go test ./...

go-lint:
	golangci-lint run ./...

go-fmt-check:
	@echo "Checking Go formatting..."
	@if [ -n "$$(gofumpt -l .)" ]; then \
		echo "Go code is not formatted (gofumpt). Run 'gofumpt -w .'"; \
		exit 1; \
	fi
	@if [ -n "$$(goimports -l .)" ]; then \
		echo "Go code has missing/unordered imports. Run 'goimports -w .'"; \
		exit 1; \
	fi

tools-install:
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install github.com/google/addlicense@$(ADDLICENSE_VERSION)

# --- Rust (rust/ root) ---

rust-fmt-check:
	cd rust && cargo fmt --check

rust-lint:
	cd rust && cargo clippy -p cortex-mcp -p xray --all-targets -- -D warnings

rust-test:
	cd rust && cargo test -p cortex-mcp -p xray

rust-build:
	cd rust && cargo build -p cortex-mcp -p xray --release
