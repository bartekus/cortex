# CLI Command Interface

**Feature**: `CLI_CONTRACT`
**Status**: Approved

## Purpose
Defines the user-facing command-line interface of the `cortex` binary. This contract guarantees the stability of command flags, exit codes, and output formats.

## Global Interface
### Environment Variables
| Variable | Purpose                                                                |
| :--- |:-----------------------------------------------------------------------|
| `CORTEX_VERSION` | **Legacy Support**. Overrides the version reported by `cortex version`.|

### Global Flags
| Flag | Short | Type | Description |
| :--- | :--- | :--- | :--- |
| `--verbose` | `-v` | Bool | Enable verbose logging to stderr. |

### Exit Codes
| Code | Meaning |
| :--- | :--- |
| `0` | Success. |
| `1` | General error (command failed, check failed, lint failed). |

## Command Tree
The following top-level commands are guaranteed to exist:
- `version`: Print version info.
- `run`: Canonical task runner (wraps internal logic).
- `gov`: Governance and spec utilities.
- `status`: Repository health and status checks.
- `features`: Feature flag and registry management.
- `context`: Context management (build, docs, xray).

*Note: `xray` is aliased under `context` or available as `cortex context xray` depending on exact command wiring, but `cortex context xray` is the canonical path in this contract.*

## Output Policy
- **Machine Output**: Some subcommands may support JSON output (via a scoped `--json` flag). When enabled, output must be minified and schema-compliant.
- **Human Output**: Default stdout is for humans. Structure is not guaranteed stable unless explicitly documented.
- **Stderr**: Used for logs, progress bars, and errors.

## Example: Canonical Help
```text
Cortex provides repository scanning, governance checks, and AI context generation tools.

Usage:
  cortex [command]

Available Commands:
  context     Context management tools
  features    Feature registry management
  gov         Governance utilities
  run         Execute defined tasks
  status      Check repository status
  version     Print the version number of Cortex

Flags:
  -h, --help      help for cortex
  -v, --verbose   enable verbose output
```
