---
feature: CLI_COMMAND_CONTEXT
version: v1
status: approved
domain: cli
inputs:
  flags:
    - name: --xray-bin
  args:
    - name: subcommand
outputs:
  exit_codes:
    0: 0
    1: 1
---
# CLI Command: Context

## Summary

The `context` command manages the AI context pipeline, including building `.cortex/` representations and running XRAY scans.

## Surface

- **Command**: `cortex context [subcommand]`
- **Subcommands**:
  - `build`: Build AI context representation.
  - `docs`: Generate deterministic documentation from XRAY index.
  - `xray`: Run XRAY scan.

## Flags

- `--xray-bin <path>`: Path to custom xray binary.
- `--output <path>`: (Subcommand `xray scan` only) Output directory for index.

## Behavior

- **Build**: Orchestrates XRAY scan -> Index read -> Context builder.
- **XRAY Wrapper**: Proxies commands to the Rust XRAY binary.
- **Docs**: Projects XRAY index into deterministic Markdown documentation.

## Subcommand: `docs`

### Usage

```bash
cortex context docs
```

### Contract

#### Inputs

- **XRAY Index**: `.cortex/data/index.json` (Required)
- Must conform to `spec/xray/index-format.md`.

#### Outputs

- **Directory**: `docs/__generated__/context/`
- **Files**:
  - `index.md`: Repository overview (stats, languages, top dirs).
  - `files.md`: Flat list of files with metadata.
  - `modules.md`:  List of module configuration files (as reported by XRAY).

#### Determinism

> The generator MUST produce deterministic output for a given input index.

- **Maps**: Keys for `languages` and `topDirs` MUST be sorted lexicographically before rendering.
- **Lists**: `files` and `moduleFiles` MUST be rendered in the order provided by the XRAY index (which guarantees sortedness).
- **Paths**:  Absolute paths MUST NOT be included in the output; paths MUST be repo-relative.
- **Timestamps**: No generation timestamps allowed.

## References

	•	cmd/cortex/commands/context.go
	•	internal/builder
	•	internal/contextdocs
	•	internal/projection
