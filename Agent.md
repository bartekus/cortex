# Cortex Agent Protocol

This document defines the contract for AI Agents and automated systems interacting with this repository.

## Operational Constraints

1.  **Determinism**: All outputs must be reproducible. Avoid timestamps, random ordering, or flaky logic in generated artifacts.
2.  **No Side Effects**: Commands labeled as "read-only" or "inspection" must never modify the filesystem.
3.  **Parity**: Assume the `Makefile` is the single source of truth for build verification. Do not invent new workflows.

## Conventions

- **Golden Tests**: CLI output changes must be verified against `cmd/cortex/testdata/*.golden`.
- **Rust Integration**: Rust code resides in `rust/` but is built/tested via root `Makefile` targets.
- **Hygiene**: `test-state/` and `bin/` are gitignored. Do not commit runtime artifacts.

## Critical Paths

- **Context Pipeline**: Used to generate AI context. Must remain hermetic.
- **XRAY**: The Rust-based repository scanner. Must be built via `make build` (calls `cargo build --release`).
