// Feature: MCP_TOOLS
// Spec: spec/mcp/tools.md

// Router module
pub mod cache;
pub mod mounts;

use crate::io::fs::RealFs;
use crate::resolver::order::ResolveEngine;
use crate::router::mounts::MountRegistry;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::sync::Arc;

#[derive(Serialize, Deserialize, Debug)]
pub struct JsonRpcRequest {
    pub jsonrpc: String,
    pub method: String,
    pub params: Option<Value>,
    pub id: Option<Value>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct JsonRpcResponse {
    pub jsonrpc: String,
    pub result: Option<Value>,
    pub error: Option<Value>,
    pub id: Option<Value>,
}

use crate::snapshot::tools::SnapshotTools;
use crate::workspace::WorkspaceTools;

pub struct Router {
    resolver: Arc<ResolveEngine<RealFs>>,
    mounts: MountRegistry,
    snapshot_tools: Arc<SnapshotTools>,
    workspace_tools: Arc<WorkspaceTools>,
}

impl Router {
    pub fn new(
        resolver: Arc<ResolveEngine<RealFs>>,
        mounts: MountRegistry,
        snapshot_tools: Arc<SnapshotTools>,
        workspace_tools: Arc<WorkspaceTools>,
    ) -> Self {
        Self {
            resolver,
            mounts,
            snapshot_tools,
            workspace_tools,
        }
    }

    pub fn handle_request(&self, req: &JsonRpcRequest) -> JsonRpcResponse {
        match req.method.as_str() {
            "initialize" => json_rpc_ok(
                req.id.clone(),
                json!({
                    "protocolVersion": "2024-11-05",
                    "capabilities": get_server_capabilities(),
                    "serverInfo": { "name": "cortex-mcp", "version": "0.1.0" }
                }),
            ),
            "tools/list" => json_rpc_ok(
                req.id.clone(),
                json!({
                    "tools": [
                        {
                            "name": "resolve_mcp",
                            "description": "Resolve an MCP server name to a local path or alias",
                            "inputSchema": {
                                "type": "object",
                                "properties": { "name": { "type": "string" } },
                                "required": ["name"]
                            }
                        },
                        {
                            "name": "list_mounts",
                            "description": "List currently resolved/mounted servers",
                            "inputSchema": { "type": "object", "properties": {} }
                        },
                        {
                            "name": "snapshot.create",
                            "description": "Create a new snapshot",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "repo_root": { "type": "string" },
                                    "lease_id": { "type": "string" },
                                    "paths": { "type": "array", "items": { "type": "string" } }
                                },
                                "required": ["repo_root"]
                            }
                        },
                        {
                            "name": "snapshot.list",
                            "description": "List files in a snapshot or worktree",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "repo_root": { "type": "string" },
                                    "path": { "type": "string" },
                                    "mode": { "type": "string", "enum": ["worktree", "snapshot"] },
                                    "lease_id": { "type": "string" },
                                    "snapshot_id": { "type": "string" }
                                },
                                "required": ["repo_root", "path", "mode"]
                            }
                        },
                        {
                            "name": "snapshot.file",
                            "description": "Read a file from snapshot or worktree",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "repo_root": { "type": "string" },
                                    "path": { "type": "string" },
                                    "mode": { "type": "string", "enum": ["worktree", "snapshot"] },
                                    "lease_id": { "type": "string" },
                                    "snapshot_id": { "type": "string" }
                                },
                                "required": ["repo_root", "path", "mode"]
                            }
                        },
                        {
                            "name": "snapshot.grep",
                            "description": "Search for a pattern",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "repo_root": { "type": "string" },
                                    "pattern": { "type": "string" },
                                    "path": { "type": "string" },
                                    "mode": { "type": "string", "enum": ["worktree", "snapshot"] },
                                    "lease_id": { "type": "string" },
                                    "snapshot_id": { "type": "string" },
                                    "case_insensitive": { "type": "boolean" }
                                },
                                "required": ["repo_root", "pattern", "mode"]
                            }
                        },
                        {
                            "name": "workspace.apply_patch",
                            "description": "Apply a patch",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "repo_root": { "type": "string" },
                                    "patch": { "type": "string" },
                                    "mode": { "type": "string", "enum": ["worktree", "snapshot"] },
                                    "strip": { "type": "integer" },
                                    "reject_on_conflict": { "type": "boolean" },
                                    "dry_run": { "type": "boolean" },
                                    "lease_id": { "type": "string" },
                                    "snapshot_id": { "type": "string" }
                                },
                                "required": ["repo_root", "patch", "mode"]
                            }
                        },
                        {
                            "name": "workspace.write_file",
                            "description": "Write a file",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "repo_root": { "type": "string" },
                                    "path": { "type": "string" },
                                    "content": { "type": "string" },
                                    "lease_id": { "type": "string" },
                                    "create_dirs": { "type": "boolean" },
                                    "dry_run": { "type": "boolean" }
                                },
                                "required": ["repo_root", "path", "content", "lease_id"]
                            }
                        },
                        {
                            "name": "workspace.delete",
                            "description": "Delete a file",
                            "inputSchema": {
                                "type": "object",
                                "properties": {
                                    "repo_root": { "type": "string" },
                                    "path": { "type": "string" },
                                    "lease_id": { "type": "string" },
                                    "dry_run": { "type": "boolean" }
                                },
                                "required": ["repo_root", "path", "lease_id"]
                            }
                        }
                    ]
                }),
            ),

            "tools/call" => {
                let params = match req.params.as_ref().and_then(|p| p.as_object()) {
                    Some(p) => p,
                    None => return json_rpc_error(req.id.clone(), -32602, "Invalid params"),
                };
                let name = match params.get("name").and_then(|n| n.as_str()) {
                    Some(n) => n,
                    None => return json_rpc_error(req.id.clone(), -32602, "Missing tool name"),
                };
                let args = params.get("arguments").and_then(|a| a.as_object());
                let args = match args {
                    Some(a) => a,
                    None => return json_rpc_error(req.id.clone(), -32602, "Missing arguments"),
                };

                match name {
                    "resolve_mcp" => {
                        let target = args.get("name").and_then(|n| n.as_str());
                        if let Some(target) = target {
                            match self.resolver.resolve(target) {
                                Ok(resp) => {
                                    if resp.status
                                        == crate::protocol::types::ResolveStatus::Resolved
                                    {
                                        if let (Some(root), Some(rid)) =
                                            (&resp.root, &resp.resolved_id)
                                        {
                                            self.mounts.register(crate::router::mounts::Mount {
                                                name: target.to_string(),
                                                root: root.clone(),
                                                resolved_id: Some(rid.clone()),
                                                kind: resp.kind.clone(),
                                                capabilities: resp.capabilities.clone(),
                                            });
                                        }
                                    }
                                    let content = json!([{ "type": "json", "json": resp }]);
                                    json_rpc_ok(req.id.clone(), json!({ "content": content }))
                                }
                                Err(e) => json_rpc_error(
                                    req.id.clone(),
                                    -32603,
                                    &format!("Resolution failed: {}", e),
                                ),
                            }
                        } else {
                            json_rpc_error(req.id.clone(), -32602, "Missing name argument")
                        }
                    }
                    "list_mounts" => {
                        let list = self.mounts.list();
                        json_rpc_ok(
                            req.id.clone(),
                            json!({ "content": [{ "type": "json", "json": list }] }),
                        )
                    }
                    "get_capabilities" => {
                        let caps = json!({
                            "name": "cortex-mcp",
                            "server_capabilities": get_server_capabilities(),
                        });
                        json_rpc_ok(
                            req.id.clone(),
                            json!({ "content": [{ "type": "json", "json": caps }] }),
                        )
                    }
                    "snapshot.create" => {
                        let repo_root = args.get("repo_root").and_then(|s| s.as_str());
                        let lease_id = args.get("lease_id").and_then(|s| s.as_str());
                        let paths = args.get("paths").and_then(|v| v.as_array()).map(|arr| {
                            arr.iter()
                                .filter_map(|v| v.as_str().map(|s| s.to_string()))
                                .collect()
                        });

                        if let Some(root) = repo_root {
                            match self.snapshot_tools.create_snapshot(
                                std::path::Path::new(root),
                                lease_id.map(|s| s.to_string()),
                                paths,
                            ) {
                                Ok(sid) => json_rpc_ok(
                                    req.id.clone(),
                                    json!({ "content": [{ "type": "json", "json": { "snapshot_id": sid } }] }),
                                ),
                                Err(e) => json_rpc_error(req.id.clone(), -32603, &e.to_string()),
                            }
                        } else {
                            json_rpc_error(req.id.clone(), -32602, "Missing repo_root")
                        }
                    }
                    "snapshot.list" => {
                        let repo_root = args.get("repo_root").and_then(|s| s.as_str());
                        let path = args.get("path").and_then(|s| s.as_str());
                        let mode = args.get("mode").and_then(|s| s.as_str());
                        let lease_id = args.get("lease_id").and_then(|s| s.as_str());
                        let snapshot_id = args.get("snapshot_id").and_then(|s| s.as_str());

                        if let (Some(root), Some(p), Some(m)) = (repo_root, path, mode) {
                            match self.snapshot_tools.list_snapshot(
                                std::path::Path::new(root),
                                p,
                                m,
                                lease_id.map(|s| s.to_string()),
                                snapshot_id.map(|s| s.to_string()),
                            ) {
                                Ok(res) => json_rpc_ok(
                                    req.id.clone(),
                                    json!({ "content": [{ "type": "json", "json": res }] }),
                                ),
                                Err(e) => json_rpc_error(req.id.clone(), -32603, &e.to_string()),
                            }
                        } else {
                            json_rpc_error(req.id.clone(), -32602, "Missing required arguments")
                        }
                    }
                    "snapshot.file" => {
                        let repo_root = args.get("repo_root").and_then(|s| s.as_str());
                        let path = args.get("path").and_then(|s| s.as_str());
                        let mode = args.get("mode").and_then(|s| s.as_str());
                        let lease_id = args.get("lease_id").and_then(|s| s.as_str());
                        let snapshot_id = args.get("snapshot_id").and_then(|s| s.as_str());

                        if let (Some(root), Some(p), Some(m)) = (repo_root, path, mode) {
                            match self.snapshot_tools.read_file(
                                std::path::Path::new(root),
                                p,
                                m,
                                lease_id.map(|s| s.to_string()),
                                snapshot_id.map(|s| s.to_string()),
                            ) {
                                Ok(bytes) => {
                                    // Encode base64
                                    use base64::{engine::general_purpose, Engine as _};
                                    let b64 = general_purpose::STANDARD.encode(bytes);
                                    json_rpc_ok(
                                        req.id.clone(),
                                        json!({ "content": [{ "type": "json", "json": { "content": b64 } }] }),
                                    )
                                }
                                Err(e) => json_rpc_error(req.id.clone(), -32603, &e.to_string()),
                            }
                        } else {
                            json_rpc_error(req.id.clone(), -32602, "Missing required arguments")
                        }
                    }
                    "snapshot.grep" => {
                        let repo_root = args.get("repo_root").and_then(|s| s.as_str());
                        let pattern = args.get("pattern").and_then(|s| s.as_str());
                        let path = args.get("path").and_then(|s| s.as_str()).unwrap_or(".");
                        let mode = args.get("mode").and_then(|s| s.as_str());
                        let lease_id = args.get("lease_id").and_then(|s| s.as_str());
                        let snapshot_id = args.get("snapshot_id").and_then(|s| s.as_str());
                        let case_insensitive = args
                            .get("case_insensitive")
                            .and_then(|b| b.as_bool())
                            .unwrap_or(false);

                        if let (Some(root), Some(pat), Some(m)) = (repo_root, pattern, mode) {
                            match self.snapshot_tools.grep_snapshot(
                                std::path::Path::new(root),
                                pat,
                                path,
                                m,
                                lease_id.map(|s| s.to_string()),
                                snapshot_id.map(|s| s.to_string()),
                                case_insensitive,
                            ) {
                                Ok(res) => {
                                    // Helper: convert (path, line, content) to objects
                                    let matches: Vec<Value> = res.into_iter().map(|(p, l, c)| {
                                         json!({ "file_path": p, "line_number": l, "content": c })
                                     }).collect();
                                    json_rpc_ok(
                                        req.id.clone(),
                                        json!({ "content": [{ "type": "json", "json": { "matches": matches } }] }),
                                    )
                                }
                                Err(e) => json_rpc_error(req.id.clone(), -32603, &e.to_string()),
                            }
                        } else {
                            json_rpc_error(req.id.clone(), -32602, "Missing required arguments")
                        }
                    }
                    "workspace.apply_patch" => {
                        let repo_root = args.get("repo_root").and_then(|s| s.as_str());
                        let patch = args.get("patch").and_then(|s| s.as_str());
                        let mode = args.get("mode").and_then(|s| s.as_str());
                        let strip = args
                            .get("strip")
                            .and_then(|v| v.as_u64())
                            .map(|u| u as usize);
                        let reject_on_conflict = args
                            .get("reject_on_conflict")
                            .and_then(|b| b.as_bool())
                            .unwrap_or(true);
                        let dry_run = args
                            .get("dry_run")
                            .and_then(|b| b.as_bool())
                            .unwrap_or(false);
                        let lease_id = args.get("lease_id").and_then(|s| s.as_str());
                        let snapshot_id = args.get("snapshot_id").and_then(|s| s.as_str());

                        if let (Some(root), Some(p), Some(m)) = (repo_root, patch, mode) {
                            match self.workspace_tools.apply_patch(
                                std::path::Path::new(root),
                                p,
                                m,
                                lease_id.map(|s| s.to_string()),
                                snapshot_id.map(|s| s.to_string()),
                                strip,
                                reject_on_conflict,
                                dry_run,
                            ) {
                                Ok(val) => json_rpc_ok(
                                    req.id.clone(),
                                    json!({ "content": [{ "type": "json", "json": val }] }),
                                ),
                                Err(e) => json_rpc_error(req.id.clone(), -32603, &e.to_string()),
                            }
                        } else {
                            json_rpc_error(req.id.clone(), -32602, "Missing required arguments")
                        }
                    }
                    "workspace.write_file" => {
                        let repo_root = args.get("repo_root").and_then(|s| s.as_str());
                        let path = args.get("path").and_then(|s| s.as_str());
                        let content = args.get("content").and_then(|s| s.as_str());
                        let lease_id = args.get("lease_id").and_then(|s| s.as_str());
                        let create_dirs = args
                            .get("create_dirs")
                            .and_then(|b| b.as_bool())
                            .unwrap_or(false);
                        let dry_run = args
                            .get("dry_run")
                            .and_then(|b| b.as_bool())
                            .unwrap_or(false);

                        if let (Some(root), Some(p), Some(c), Some(lid)) =
                            (repo_root, path, content, lease_id)
                        {
                            match self.workspace_tools.write_file(
                                std::path::Path::new(root),
                                p,
                                c,
                                Some(lid.to_string()),
                                create_dirs,
                                dry_run,
                            ) {
                                Ok(res) => json_rpc_ok(
                                    req.id.clone(),
                                    json!({ "content": [{ "type": "json", "json": { "written": res } }] }),
                                ),
                                Err(e) => json_rpc_error(req.id.clone(), -32603, &e.to_string()),
                            }
                        } else {
                            json_rpc_error(req.id.clone(), -32602, "Missing required arguments")
                        }
                    }
                    "workspace.delete" => {
                        let repo_root = args.get("repo_root").and_then(|s| s.as_str());
                        let path = args.get("path").and_then(|s| s.as_str());
                        let lease_id = args.get("lease_id").and_then(|s| s.as_str());
                        let dry_run = args
                            .get("dry_run")
                            .and_then(|b| b.as_bool())
                            .unwrap_or(false);

                        if let (Some(root), Some(p), Some(lid)) = (repo_root, path, lease_id) {
                            match self.workspace_tools.delete(
                                std::path::Path::new(root),
                                p,
                                Some(lid.to_string()),
                                dry_run,
                            ) {
                                Ok(res) => json_rpc_ok(
                                    req.id.clone(),
                                    json!({ "content": [{ "type": "json", "json": { "deleted": res } }] }),
                                ),
                                Err(e) => json_rpc_error(req.id.clone(), -32603, &e.to_string()),
                            }
                        } else {
                            json_rpc_error(req.id.clone(), -32602, "Missing required arguments")
                        }
                    }
                    _ => json_rpc_error(req.id.clone(), -32601, "Tool not found"),
                }
            }
            "notifications/initialized" => JsonRpcResponse {
                jsonrpc: "2.0".into(),
                result: None,
                error: None,
                id: None,
            },
            _ => json_rpc_error(
                req.id.clone(),
                -32601,
                &format!("Method not found: {}", req.method),
            ),
        }
    }
}

fn get_server_capabilities() -> Value {
    json!({
        "tools": { "listChanged": true },
        // Add other server capabilities here if needed (e.g. logging, resources)
    })
}

fn json_rpc_ok(id: Option<Value>, result: Value) -> JsonRpcResponse {
    JsonRpcResponse {
        jsonrpc: "2.0".to_string(),
        result: Some(result),
        error: None,
        id,
    }
}

fn json_rpc_error(id: Option<Value>, code: i32, message: &str) -> JsonRpcResponse {
    JsonRpcResponse {
        jsonrpc: "2.0".to_string(),
        result: None,
        error: Some(json!({ "code": code, "message": message })),
        id,
    }
}
