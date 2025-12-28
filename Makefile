# SPDX-License-Identifier: AGPL-3.0-or-later
# Cortex Makefile

# Feature: CORE_REPO_CONTRACT
# Spec: spec/system/contract.md

SHELL := /bin/bash

# Tools versions (overridable)
GOFUMPT_VERSION := v0.6.0
GOIMPORTS_VERSION := v0.27.0
GOLANGCI_LINT_VERSION := v1.63.4
ADDLICENSE_VERSION := v1.1.1

.PHONY: all build test lint fmt-check go-build go-test go-lint go-mod-tidy-check go-fmt-check tools-install rust-build rust-test rust-lint rust-fmt-check gov-onboard

install:
	@echo "Installing gofumpt@$(GOFUMPT_VERSION)"
	@go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	@echo "Installing goimports@$(GOIMPORTS_VERSION)"
	@go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	@echo "Installing golangci-lint@$(GOLANGCI_LINT_VERSION)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@echo "Installing addlicense@$(ADDLICENSE_VERSION)"
	@go install github.com/google/addlicense@$(ADDLICENSE_VERSION)
	@echo "installing node dependencies for tests harness"
	@cd tests && npm install

all: clean fmt-check lint test build context docs reports gov

check: validate-and-build;

clean:
	@rm -rf .cortex docs/__generated__ tests/fixtures/run/_cortex_data

docs: build
	@echo " "
	@echo "Validate and visualize the feature DAG."
	@./bin/cortex features graph
	@echo "Analyze downstream impact of a feature: CORE_REPO_CONTRACT."
	@./bin/cortex features impact CORE_REPO_CONTRACT
	@echo " "
	@echo "Generate feature overview documentation."
	@./bin/cortex features overview

# --- Top-level targets (canonical) ---
context: build
	@echo " "
	@echo "Generating AI context representation."
	@./bin/cortex context build
	@echo "Generate AI-Agent documentation."
	@./bin/cortex context docs
	@echo " "

gov: build
	@echo " "
	@echo "Running governance checks..."
	@./bin/cortex gov drift help
	@echo " "
	@echo "Detect drift between implementation and fixtures..."
	@./bin/cortex gov drift xray
	@echo " "
	@echo "Validate feature/spec/code/test mapping..."
	@./bin/cortex gov feature-mapping
	@echo " "
	@echo "Validate spec file frontmatter..."
	@./bin/cortex gov spec-validate
	@echo " "
	@echo "Dump the CLI command tree (commands + flags) to JSON for spec-vs-cli"
	@./bin/cortex gov cli-dump-json --out .cortex/data/cli.json
	@echo " "
	@echo "Validate alignment between CLI help output and Spec flags"
	@./bin/cortex gov spec-vs-cli --binary-json .cortex/data/cli.json
	@echo " "
	@echo "Validate the feature registry and spec integrity..."
	@./bin/cortex gov validate
	@echo " "

reports: build
	@echo " "
	@echo "Generates a commit health report analyzing commit message discipline."
	@./bin/cortex reports commit-report
	@echo "Saved as ./.cortex/reports/commit-health.json"
	@echo " "
	@echo "Generates a feature traceability report analyzing feature presence across spec, implementation, tests, and commits."
	@./bin/cortex reports feature-traceability
	@echo "Saved as ./.cortex/reports/feature-traceability.json"
	@echo " "
	@echo "Reads commit health and feature traceability reports and generates actionable suggestions for improving commit discipline."
	@./bin/cortex reports commit-suggest
	@echo " "
	@echo " "
	@echo "Generate phase-level feature completion analysis from spec/features.yaml."
	@./bin/cortex reports status-roadmap
	@echo "Saved as docs/__generated__/feature-completion-analysis.md"

validate-and-build: validate-and-build-rust validate-and-build-go

build: go-build rust-build

validate-and-build-go: go-fmt-check go-lint go-test go-build

validate-and-build-rust: rust-fmt-check rust-lint rust-test rust-build

test: go-test rust-test node-test

lint: go-lint rust-lint

fmt-check: go-fmt-check rust-fmt-check

# --- Go (repo root) ---

go-build:
	@echo "Building Go..."
	@go build -trimpath -ldflags="-s -w" -o ./bin/cortex ./cmd/cortex
	@echo " "

go-test:
	@echo "Testing Go..."
	@go test ./...
	@echo " "

go-lint: go-mod-tidy-check
	@echo "Linting Go..."
	@golangci-lint run ./...

go-mod-tidy-check:
	@echo "Checking go.mod/go.sum tidiness..."
	@cp go.mod go.mod.bak
	@cp go.sum go.sum.bak
	@go mod tidy
	@if ! diff go.mod go.mod.bak > /dev/null; then \
		echo "go.mod is not tidy"; \
		mv go.mod.bak go.mod; \
		mv go.sum.bak go.sum; \
		exit 1; \
	fi
	@if ! diff go.sum go.sum.bak > /dev/null; then \
		echo "go.sum is not tidy"; \
		mv go.mod.bak go.mod; \
		mv go.sum.bak go.sum; \
		exit 1; \
	fi
	@rm go.mod.bak go.sum.bak

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

# --- Rust (rust/ root) ---

rust-fmt-check:
	@echo "Checking Rust formatting..."
	@cd rust && cargo fmt --check

rust-lint:
	@echo "Linting Rust..."
	@cd rust && cargo clippy -p cortex-mcp -p xray --all-targets -- -D warnings

rust-test:
	@echo "Testing Rust..."
	@cd rust && cargo test -p cortex-mcp -p xray && cd ../tests && npm test

rust-build:
	@echo "Building Rust..."
	@cd rust && cargo build -p cortex-mcp -p xray --release && cp ./target/release/xray ../bin/xray && cp ./target/release/cortex-mcp ../bin/cortex-mcp

# --- NodeJS (tests/ root) ---
node-test:
	@echo "Testing Rust..."
	@cd tests && npm test
