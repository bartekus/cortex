---
feature: MCP_TOOLS
version: v1
status: approved
domain: mcp
inputs:
  args:
    - name (tool name)
    - arguments (json object)
outputs:
  result: JSON object
---
# MCP Tools
## Summary
Standard tools exposed by the Cortex MCP server.

## Surface
- **Method**: `tools/call`

## Tools

### `resolve_mcp`
- **Purpose**: Resolve MCP server configuration for a named service.
- **Inputs**:
  - `name`: String (Service name).
- **Outputs**:
  - `config`: JSON object containing server configuration (command, args, env).

### `list_mounts`
- **Purpose**: List currently active file system mounts.
- **Inputs**: None.
- **Outputs**:
  - `mounts`: Array of mount objects (host path, mount point).

## References
- `rust/mcp/src/main.rs`
- `rust/mcp/src/router/`
