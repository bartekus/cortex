# Spec Coverage Map

This document maps all user-facing surfaces to their authoritative specifications.

## Go CLI Commands

| Command | Surface | Spec Contract | Status |
| :--- | :--- | :--- | :--- |
| `cortex (root)` | `cmd/cortex` | `spec/cli/contract.md` | **COVERED** |
| `cortex run` | `cmd/cortex/commands/run.go` | `spec/cli/run.md` | **MISSING** |
| `cortex context` | `cmd/cortex/commands/context.go` | `spec/cli/context.md` | **MISSING** |
| `cortex features` | `cmd/cortex/commands/features.go` | `spec/cli/features.md` | **MISSING** |
| `cortex gov` | `cmd/cortex/commands/gov.go` | `spec/cli/gov.md` | **MISSING** |
| `cortex status` | `cmd/cortex/commands/status.go` | `spec/cli/status.md` | **MISSING** |
| `cortex commit` | `cmd/cortex/commands/commit_*.go` | `spec/cli/commit.md` | **MISSING** |
| `cortex feature` | `cmd/cortex/commands/feature_traceability.go` | `spec/cli/feature.md` | **MISSING** |

## Rust XRAY CLI

| Command | Surface | Spec Contract | Status |
| :--- | :--- | :--- | :--- |
| `xray scan` | `rust/xray` (scan subcommand) | `spec/xray/cli.md` | **MISSING** |
| `xray` (library) | `rust/xray` (core logic) | `spec/xray/scan-policy.md` | **COVERED** |
| `xray` (index) | `rust/xray` (schema) | `spec/xray/index-format.md` | **COVERED** |

## Rust MCP Server

| Component | Surface | Spec Contract | Status |
| :--- | :--- | :--- | :--- |
| Protocol / Router | `rust/mcp` | `spec/mcp/contract.md` | **COVERED** |
| Tool: `resolve_mcp` | `rust/mcp` | `spec/mcp/tools.md` | **MISSING** |
| Tool: `list_mounts` | `rust/mcp` | `spec/mcp/tools.md` | **MISSING** |

## Skills & Reports

| Component | Surface | Spec Contract | Status |
| :--- | :--- | :--- | :--- |
| Registry | `internal/skills` | `spec/skills/registry.md` | **MISSING** |
| Report: Commit Health | `internal/reports/commithealth` | `spec/reports/commit-health.md` | **MISSING** |
| Report: Traceability | `internal/reports/featuretrace` | `spec/reports/traceability.md` | **MISSING** |
