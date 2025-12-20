# Cortex Command Inventory

This document lists all user-facing surfaces of the Cortex system, including the Go-based `cortex` CLI, the Rust-based `xray` engine, and the Rust-based `mcp` server.

## 1. Go CLI (`cortex`)

**Binary**: `cmd/cortex`
**Entry Point**: `cmd/cortex/main.go`

### Root Command
- **Usage**: `cortex [command]`
- **Flags**:
  - `-v, --verbose`: Enable verbose output (Global)
  - `-h, --help`: Help for cortex

### Subcommands

#### `version`
- **Output**: `Cortex version <version>`

#### `run`
- **Usage**: `cortex run <command|skill> [flags]`
- **Sources**: `cmd/cortex/commands/run.go`
- **Subcommands**:
  - `list`: List available skills.
    - Flags: `--json` (Output JSON)
  - `all`: Run all skills.
  - `resume`: Resume from last failure.
  - `reset`: Clear run state.
  - `report`: Show last run status.
    - Flags: `--json` (Output JSON)
- **Flags**:
  - `--json`: Output results in JSON.
  - `--state-dir`: Directory to store run state (default: `.cortex/run`).
  - `--fail-on-warning`: Fail if warnings occur.
  - `--files0`: Read NULL-delimited file list from stdin.

#### `context`
- **Usage**: `cortex context [subcommand]`
- **Sources**: `cmd/cortex/commands/context.go`
- **Flags**:
  - `--xray-bin`: Path to xray binary.
- **Subcommands**:
  - `build`: Build AI context representation.
  - `docs`: Generate AI-Agent documentation.
  - `xray`: Run XRAY scan.
    - `scan [target]`: Run XRAY scan against target.
      - Flags: `--output` (Output directory).
    - `docs`: Run XRAY docs (not implemented).
    - `all`: Run XRAY all (not implemented).

#### `features`
- **Usage**: `cortex features [subcommand]`
- **Sources**: `cmd/cortex/commands/features.go`
- **Subcommands**:
  - `graph`: Visualize feature dependency graph.
  - `impact`: Analyze feature impact.
  - `overview`: Show feature overview.

#### `gov`
- **Usage**: `cortex gov [subcommand]`
- **Sources**: `cmd/cortex/commands/gov.go`, `*_validate.go`, `*_drift.go`
- **Subcommands**:
  - `feature-mapping`: Validate feature/spec/code/test mapping.
    - Flags: `--format` (text|json).
  - `spec-validate`: Validate specifications.
  - `spec-vs-cli`: Validate spec vs CLI implementation.
  - `validate`: Run general governance validation.
  - `drift`: Check for governance drift.

#### `status`
- **Usage**: `cortex status [subcommand]`
- **Sources**: `cmd/cortex/commands/status.go`
- **Subcommands**:
  - `roadmap`: Generate feature completion analysis.
    - Flags: `--features` (path), `--output` (path).

#### `commit`
- **Usage**: `cortex commit [subcommand]`
- **Sources**: `cmd/cortex/commands/commit_report.go`, `commit_suggest.go`
- **Subcommands**:
  - `report`: Generate commit health report.
    - Flags: `--from`, `--to`.
  - `suggest`: Generate commit discipline suggestions.
    - Flags: `--format`, `--severity`, `--max-suggestions`.

#### `feature` (Singular)
- **Usage**: `cortex feature` (Note: Distinct from `features`)
- **Sources**: `cmd/cortex/commands/feature_traceability.go`
- **Description**: Generate feature traceability report.

## 2. Rust XRAY CLI (`xray`)

**Binary**: `rust/xray`
**Entry Point**: `rust/xray/src/main.rs`

### Root Command
- **Usage**: `xray <COMMAND>`
- **Flags**: `-h, --help`, `-V, --version`

### Subcommands
- `scan [target]`: Scans the repository and updates `.cortex`.
  - Args: `target` (default: `.`)
  - Flags: `--output <dir>` (Override output directory)
- `docs`: Generate documentation (Not implemented).
- `all`: Run all steps (Not implemented).

## 3. Rust MCP Server (`mcp`)

**Binary**: `rust/mcp`
**Entry Point**: `rust/mcp/src/main.rs`

### Protocol
- **Transport**: Stdio (JSON-RPC 2.0 with MCP framing).
- **Framing**: `Content-Length: <n>\r\n\r\n<payload>`

### Capabilities (`tools/list`)
- `resolve_mcp`: Resolves MCP server configurations.
  - Args: `name` (string)
- `list_mounts`: Lists active file mounts.
  - Args: None

## 4. Skills Registry
**Source**: `internal/skills/`
**Invoked via**: `cortex run <skill_id>`

- `docs:doc-patterns`
- `docs:feature-integrity`
- `docs:header-comments`
- `docs:orphan-docs`
- `docs:orphan-specs`
- `docs:policy`
- `docs:provider-governance`
- `docs:validate-spec`
- `docs:yaml`
- `format:gofumpt`
- `lint:gofumpt`
- `lint:golangci`
- `purity`
- `test:basic`
- `test:coverage`

## 5. Reports
**Source**: `internal/reports/`
**Invoked via**: CLI commands

- **Commit Health**: `cortex commit report`
- **Feature Traceability**: `cortex feature`
- **Suggestions**: `cortex commit suggest`
- **Feature Completion**: `cortex status roadmap`
