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
- `internal/governance`
