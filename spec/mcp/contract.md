---
feature: MCP_ROUTER_CONTRACT
version: v1
status: approved
domain: mcp
inputs:
  transport:
    type: stdio
    protocol: json-rpc-2.0
outputs:
  exit_codes:
    0: 0
    1: 1
---

> TODO: Extend frontmatter:
```
capabilities:
  tools:
    - resolve_mcp
    - list_mounts
```

# MCP Router Contract

**Feature**: `MCP_ROUTER_CONTRACT`
**Status**: Approved

## Purpose
Defines the contract for the Cortex Model Context Protocol (MCP) server, which exposes repository capabilities to AI agents via a standardized protocol.

## Transport & Protocol
- **Transport**: Standard Input/Output (stdio).
- **Format**: JSON-RPC 2.0.
- **Framing**: Line-delimited JSON messages.

## Security Boundaries
1.  **Path Traversal**: All file access requests MUST be validated to be within the allowed `root` path(s).
2.  **Read-Only**: Unless explicitly authorized (e.g. specialized tools), default tools should be read-only or strictly scoped.

## Capability: Tool Discovery
The server implements `tools/list` to advertise capabilities.

| Tool | Purpose | Contract |
| :--- | :--- | :--- |
| `resolve_mcp` | Resolves MCP server configs. | Input: `name` (string). Output: Server config object. |
| `list_mounts` | Lists active file mounts. | Input: None. Output: List of mounts. |

## Error Model
- **JSON-RPC Errors**: Standard codes (`ParseError`, `InvalidRequest`, `MethodNotFound`).
- **Application Errors**:
    - `ToolNotFound`: If `tools/call` requests unknown tool.
    - `InvalidArgs`: If arguments do not match schema.
    - `SecurityViolation`: If path is outside allowed root.

## Example: Tool List
```json
{
  "jsonrpc": "2.0",
  "result": {
    "tools": [
      {
        "name": "list_mounts",
        "description": "List all active mounts",
        "inputSchema": { "type": "object", "properties": {} }
      }
    ]
  },
  "id": 1
}
```
