---
id: SPEC_AI_CORTEX
title: Cortex (Go) Specification
status: approved
type: spec
group: ai
---

# Cortex (Go) Specification

## 1. Overview

**Cortex** is the canonical, systems-level orchestration CLI for Stagecraft's AI context engine. It is the **only** supported entry point for generating AI context, running governance checks, and managing repository insights.

### Strategic Role
- **Orchestrator**: Cortex manages the workflow. It does not reinvent the wheel; it delegates compute-heavy scanning to **XRAY**.
- **Canonical Entry Point**: Human and CI workflows must invoke `cortex` (via `scripts/run-cortex.sh`), never `xray` directly (unless debugging) and **never** the deprecated `tools/context-compiler`.
- **Determinism Enforcer**: Cortex ensures all outputs are byte-for-byte identical across environments (local vs CI) by enforcing strict input processing and sorting.

### Architecture
- **Language**: Go (Standard Library + `spf13/cobra`).
- **Integration**: Cortex shells out to the `xray` binary (Rust) for high-performance file indexing and static analysis.
- **Location**: External Repository (`github.com/bartekus/cortex`).
- **Entry Point**: `scripts/run-cortex.sh`.

### Ownership Rules
- Cortex is the sole writer of `.ai-context/`.
- XRAY MUST NOT write to `.ai-context/`.
- XRAY writes only to `.xraycache/`.

## 2. Command Set

Cortex provides the following authoritative commands:

### `cortex context build`
Builds the standard `.ai-context/` bundle for the current repository.

- **Outcome**: Populates `.ai-context/` with:
  - `meta.json`: Timestamps NOT allowed. Deterministic metadata.
  - `files/manifest.json`: Full file list with hashes.
  - `files/chunks.ndjson`: Token-aware chunks of codebase content.
- **Mechanism**:
  1. Invokes `xray scan` to update `.xraycache/`.
  2. Reads `.xraycache/` index.
  3. Processes files into chunks (deterministically sorted).
  4. Writes atomic output to `.ai-context/`.

### `cortex context docs`
Generates human-readable documentation projections from the AI context.

- **Input**: `.ai-context/` schemas.
- **Output**: `docs/generated/` (Markdown files).
- **Behavior**: Pure projection. No external API calls.

### `cortex xray [scan|docs|all]`
Direct wrapper around the XRAY binary for convenience and debugging.

- **`scan`**: Forces a fresh scan of the repository into `.xraycache`.
- **`docs`**: Runs XRAY's documentation generation subsystem (if applicable).
- **`all`**: Runs the full XRAY suite.

### Global Flags
- `--repo-slug`: Overrides the repository identifier (default: auto-detected from git).
- `--target`: Sets the scan target directory (default: current repository root).
- `--output`: Overrides output directory (default: `.ai-context` or `.xraycache` depending on command).

## 3. Deprecation & Migration

> [!WARNING]
> **DEPRECATION NOTICE**: The Node.js-based `tools/context-compiler` is **STRICTLY FORBIDDEN** in the new architecture.

- **Status**: Deprecated.
- **Removal**: All logic must be ported to Cortex (Go) or XRAY (Rust).
- **Prohibition**: No new specs, docs, or code may reference `tools/context-compiler`, `npm run context:*`, or `tsx`.
- **Migration Path**:
  - Existing `context:build` (TS) -> `cortex context build` (Go).
  - Existing `xray:scan` (TS) -> `xray scan` (Rust) wrapped by `cortex xray scan`.
