# Cortex MCP Router

The self-healing substrate for Cortex agents.

## Build
```bash
cargo build -p cortex-mcp --release
```

## Usage
Run as an MCP server (stdio):
```bash
./target/release/cortex-mcp
```

## Configuration
- `CORTEX_WORKSPACE_ROOTS`: Colon-separated list of roots to scan (e.g., `/Users/bart/Dev:/Users/bart/src`). Defaults to `~/Dev` and `~/src`.
- `.cortex/mcp-aliases.json`: Alias map (coming soon).

## Tools
- `resolve_mcp(name)`: Resolves `owner/repo` or aliases to local paths.
- `list_mounts()`: Shows active mounts.
- `describe_skills()`: Shows registry.
