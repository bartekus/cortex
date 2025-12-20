---
feature: CLI_COMMAND_FEATURE
version: v1
status: approved
domain: cli
inputs:
  flags: []
  args: []
outputs:
  exit_codes:
    0: 0
    1: 1
---
# CLI Command: Feature (Singular)
## Summary
The `feature` command generates a traceability report for features, linking them to commits and code.

## Surface
- **Command**: `cortex feature` (No subcommands)

## Behavior
- Generates `feature-traceability.json` report.
- Scans codebase for `Feature:` annotations and correlates with `features.yaml`.

## References
- `cmd/cortex/commands/feature_traceability.go`
- `internal/reports/featuretrace`
