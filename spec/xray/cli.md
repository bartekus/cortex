---
feature: XRAY_CLI
version: v1
status: approved
domain: xray
inputs:
  flags:
    - name: --output
  args:
    - name: command
    - name: target
outputs:
  exit_codes:
    0: 0
    1: 1
---
# XRAY CLI
## Summary
The XRAY CLI is the Rust-based engine for high-performance, deterministic repository scanning.

## Surface
- **Binary**: `xray`
- **Command**: `xray <scan> [target] [flags]`

## Flags
- `--output <dir>`: Directory to write the XRAY index (default: `.cortex/data`).

## Behavior
- **Scan**: Traverses the target directory, respecting `.xrayignore` (or equivalent policy).
- **Index**: Produces a `index.json` adhering to `XRAY_INDEX_FORMAT`.
- **Determinism**: Output is strictly sorted and content-addressed (digest) to ensure reproducibility.

## References
- `rust/xray/src/main.rs`
- `spec/xray/scan-policy.md`
