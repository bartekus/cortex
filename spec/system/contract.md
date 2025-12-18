# Repository System Contract

**Feature**: `CORE_REPO_CONTRACT`
**Status**: Approved

## Purpose
Defines the foundational contracts for the `cortex` repository, including build interfaces, CI guarantees, and global determinism rules. This contract serves as the base dependency for all other features.

## Interface: Build & Test
The repository root `Makefile` is the canonical entry point for all development operations.

| Information | Contract |
| :--- | :--- |
| **Command** | `make build` |
| **Input** | Source files in `cmd/`, `internal/`, `pkg/`, `rust/` |
| **Output** | `./bin/cortex` (Go binary), `rust/target/release/xray` (Rust binary), `rust/target/release/cortex-mcp` (Rust binary) |
| **Guarantee** | Zero side effects outside `./bin/` and `./rust/target/`. |

| Information | Contract |
| :--- | :--- |
| **Command** | `make test` |
| **Scope** | Unit and Integration tests for both Go (`./...`) and Rust (`cortex-mcp`, `xray`). |
| **Guarantee** | All tests passed. Non-zero exit code on failure. |

## Interface: CI Environment
Active workflows in `.github/workflows/` define the authoritative Continuous Integration gates.

| Workflow | Trigger | Gate Description |
| :--- | :--- | :--- |
| `ci.yml` | `push`, `pull_request` | Must pass `lint` (Go+Rust), `test` (Go+Rust), and `build`. |

## Invariants: Determinism
1.  **Go Builds**: Must use `-trimpath` and `-ldflags "-s -w"` to ensure binary reproducibility.
2.  **Rust Builds**: Must build in `--release` mode for production artifacts.
3.  **Generated Code**: Committed generated code must match local regeneration (checked via `make go-mod-tidy-check` and similar targets).

## Failure Model
- **Build Failure**: Any compiler error or missing dependency yields exit code `!= 0`.
- **Lint Failure**: Standard linters (`golangci-lint`, `clippy`) must report zero issues with default config.

## Compatibility
- **Dev Environment**: Requires Go 1.23+, Rust 1.83+ (stable), Make, Bash.
- **Platform Support**: Linux (amd64/arm64), macOS (amd64/arm64), Windows (limited support).
