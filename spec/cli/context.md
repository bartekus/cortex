---
feature: CLI_COMMAND_CONTEXT
version: v1
status: approved
domain: cli
inputs:
  flags:
    - --xray-bin
  args:
    - subcommand
outputs:
  exit_codes:
    0: Success
    1: Failure
---
# CLI Command: Context
## Summary
The `context` command manages the AI context pipeline, including building .ai-context representations and running XRAY scans.

## Surface
- **Command**: `cortex context [subcommand]`
- **Subcommands**:
  - `build`: Build AI context representation.
  - `docs`: Generate AI-Agent documentation.
  - `xray`: Run XRAY scan.

## Flags
- `--xray-bin <path>`: Path to custom xray binary.
- `--output <path>`: (Subcommand `xray scan` only) Output directory for index.

## Behavior
- **Build**: Orchestrates XRAY scan -> Index Read -> Context Builder.
- **XRAY Wrapper**: Proxies commands to the Rust XRAY binary.
- **Docs**: (Deprecated node implementation) Generates markdown documentation.

## References
- `cmd/cortex/commands/context.go`
- `internal/builder`
