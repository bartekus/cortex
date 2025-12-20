---
feature: CLI_COMMAND_FEATURES
version: v1
status: approved
domain: cli
inputs:
  flags: []
  args:
    - name: subcommand
outputs:
  exit_codes:
    0: 0
    1: 1
---
# CLI Command: Features
## Summary
The `features` command provides tools for visualizing, analyzing, and documenting the feature graph defined in `spec/features.yaml`.

## Surface
- **Command**: `cortex features [subcommand]`
- **Subcommands**:
  - `graph`: Visualize feature dependency graph.
  - `impact`: Analyze feature impact.
  - `overview`: Show feature overview.

## Behavior
- **Graph**: Generates DOT or Mermaid graphs of feature dependencies.
- **Impact**: Calculates transitive impact of changes to a feature.
- **Overview**: Summarizes feature status and counts.

## References
- `cmd/cortex/commands/features.go`
- `internal/featureindex`
