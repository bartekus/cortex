---
feature: CLI_COMMAND_RUN
version: v1
status: approved
domain: cli
inputs:
  flags:
    - --json
    - --state-dir
    - --fail-on-warning
    - --files0
  args:
    - command (subcommand or skill_id)
outputs:
  exit_codes:
    0: Success
    1: Failure
---
# CLI Command: Run
## Summary
The `run` command is the canonical task runner for Cortex. It orchestrates skills, tests, and governance checks in a deterministic way, maintaining state to allow resuming failures.

## Surface
- **Command**: `cortex run <command|skill> [flags]`
- **Subcommands**:
  - `list`
  - `all`
  - `resume`
  - `reset`
  - `report`

## Flags
- `--json`: Output results in JSON format.
- `--state-dir`: Directory to store run state (default: `.cortex/run`).
- `--fail-on-warning`: Fail if warnings occur.
- `--files0`: Read NULL-delimited file list from stdin (for partial runs).

## Behavior
- **Skill Execution**: If the argument is not a subcommand, it is treated as a skill ID.
- **State Management**: Persists run results (pass/fail) to `state-dir`.
- **Determinism**: 
  - Execution order of skills is stable (lexicographic or dependency-based).
  - JSON output is sorted.

## References
- `cmd/cortex/commands/run.go`
- `internal/runner`
