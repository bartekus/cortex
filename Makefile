# SPDX-License-Identifier: AGPL-3.0-or-later
# Cortex Makefile

SHELL := /bin/bash

.PHONY: all build test lint fmt-check go-build go-test go-lint rust-build rust-test rust-lint rust-fmt-check

all: lint test build

# --- Top-level targets (canonical) ---

build: go-build rust-build

test: go-test rust-test

lint: go-lint rust-lint

fmt-check: rust-fmt-check
	@echo "fmt-check: (go formatting checks are enforced via lint/tools; add gofmt/goimports checks here if desired)"

# --- Go (repo root) ---

go-build:
	go build -o ./bin/cortex ./cmd/cortex

go-test:
	go test ./...

go-lint:
	golangci-lint run ./...

# --- Rust (rust/ root) ---

rust-fmt-check:
	cd rust && cargo fmt --check

rust-lint:
	cd rust && cargo clippy -p cortex-mcp -p xray --all-targets -- -D warnings

rust-test:
	cd rust && cargo test -p cortex-mcp -p xray

rust-build:
	cd rust && cargo build -p cortex-mcp -p xray --release
