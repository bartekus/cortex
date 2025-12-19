---
feature: CLI_COMMAND_COMMIT
version: v1
status: approved
domain: cli
inputs:
  flags:
    - --from
    - --to
    - --format
    - --severity
    - --max-suggestions
  args:
    - subcommand
outputs:
  exit_codes:
    0: Success
    1: Failure
---
# CLI Command: Commit
## Summary
The `commit` command suite analyzes git commit history for discipline and health.

## Surface
- **Command**: `cortex commit [subcommand]`
- **Subcommands**:
  - `report`: Generate commit health report.
  - `suggest`: Generate commit discipline suggestions.

## Flags
- `--from <ref>`: Start of commit range (default: origin/main).
- `--to <ref>`: End of commit range (default: HEAD).
- `--format <text|json>`: Output format (default: text).
- `--severity <info|warning|error>`: Minimum severity filter.
- `--max-suggestions <int>`: Cap usage suggestions.

## Behavior
- **Report**: Analyzes commits against conventional commit standards and feature references.
- **Suggest**: Consumes reports to suggest improvements (e.g., "Add feature tag to commit X").

## References
- `cmd/cortex/commands/commit_report.go`
- `cmd/cortex/commands/commit_suggest.go`
- `internal/reports/commithealth`
