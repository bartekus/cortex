---
feature: CLI_COMMAND_STATUS
version: v1
status: approved
domain: cli
inputs:
  flags:
    - name: --features
    - name: --output
  args:
    - name: subcommand
outputs:
  exit_codes:
    0: 0
    1: 1
---
# CLI Command: Status
## Summary
The `status` command provides high-level repository status and roadmap analysis.

## Surface
- **Command**: `cortex status [subcommand]`
- **Subcommands**:
  - `roadmap`: Generate feature completion analysis.

## Flags
- `--features <path>`: Path to `features.yaml` (default: `spec/features.yaml`).
- `--output <path>`: Output path for markdown report (default: `docs/__generated__/feature-completion-analysis.md`).

## Behavior
- **Roadmap**: Analyzes feature status (approved, draft, etc.) and groups them by phases to generate a completion report.

## References
- `cmd/cortex/commands/status.go`
- `internal/roadmap`
