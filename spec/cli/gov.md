---
feature: CLI_COMMAND_GOV
version: v1
status: approved
domain: cli
inputs:
  flags:
    - name: --format
  args:
    - name: subcommand
outputs:
  exit_codes:
    0: 0 # Success
    1: 1 # Failure
    2: 2 # Internal Error
---
# CLI Command: Gov
## Summary
The `gov` command suite performs governance checks to ensure spec compliance, registry integrity, and code alignment.

## Surface
- **Command**: `cortex gov [subcommand]`
- **Subcommands**:
  - `feature-mapping`: Validate feature/spec/code/test mapping.
  - `spec-validate`: Validate specification format and frontmatter.
  - `cli-dump-json`: Dump the CLI command tree (commands + flags) to JSON for spec-vs-cli
  - `spec-vs-cli`: Validate spec contracts against CLI implementation.
  - `validate`: Run general functional validation.
  - `drift`: Check for drift between generated artifacts and code.

## Flags
- `--format <text|json>`: Output format for reports (supported by some subcommands).

## Behavior
- **Strictness**: Governance checks are strict and intended for CI use.
- **Exit Codes**: Distinguishes between validation failures (1) and internal errors (2).

## References
- `cmd/cortex/commands/gov.go`
- `cmd/cortex/commands/gov_cli_dump_json.go`
- `cmd/cortex/commands/gov_drift.go`
- `cmd/cortex/commands/gov_spec_validate.go`
- `cmd/cortex/commands/gov_spec_vs_cli.go`
- `cmd/cortex/commands/gov_validate.go`
- `internal/governance`
