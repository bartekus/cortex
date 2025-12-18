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

pub struct Router {
    resolver: Arc<ResolveEngine<RealFs>>,
    mounts: MountRegistry,
}

impl Router {
    pub fn new(resolver: Arc<ResolveEngine<RealFs>>, mounts: MountRegistry) -> Self {
        Self { resolver, mounts }
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
                                "required": ["name"],
                                "additionalProperties": false
                            }
                        },
                        {
                            "name": "list_mounts",
                            "description": "List currently resolved/mounted servers",
                            "inputSchema": {
                                "type": "object",
                                "properties": {},
                                "additionalProperties": false
                            }
                        },
                        {
                            "name": "describe_skills",
                            "description": "List available skills",
                            "inputSchema": {
                                "type": "object",
                                "properties": {},
                                "additionalProperties": false
                            }
                        },
                        {
                            "name": "get_capabilities",
                            "description": "Get server capabilities",
                            "inputSchema": {
                                "type": "object",
                                "properties": {},
                                "additionalProperties": false
                            }
                        }
                    ]
                }),
            ),

            "tools/call" => {
                let params_result = req
                    .params
                    .as_ref()
                    .ok_or(())
                    .and_then(|p| p.as_object().ok_or(()));
                if params_result.is_err() {
                    return json_rpc_error(
                        req.id.clone(),
                        -32602,
                        "Invalid params: must be an object",
                    );
                }
                let params = params_result.unwrap();

                let name = params.get("name").and_then(|n| n.as_str());
                if name.is_none() {
                    return json_rpc_error(
                        req.id.clone(),
                        -32602,
                        "Invalid params: missing 'name'",
                    );
                }
                let name = name.unwrap();

                let args_value = params.get("arguments");
                if let Some(v) = args_value {
                    if !v.is_object() {
                        return json_rpc_error(
                            req.id.clone(),
                            -32602,
                            "Invalid params: 'arguments' must be an object",
                        );
                    }
                }
                let args = args_value.and_then(|a| a.as_object());

                let args_is_empty = match args {
                    Some(m) => m.is_empty(),
                    None => true,
                };

                match name {
                    "resolve_mcp" => {
                        // Strict validation: resolve_mcp requires 'name' inside arguments
                        let target = args.and_then(|a| a.get("name")).and_then(|n| n.as_str());
                        if target.is_none() {
                            return json_rpc_error(
                                req.id.clone(),
                                -32602,
                                "Invalid params: arguments.name is required",
                            );
                        }
                        let target = target.unwrap();

                        match self.resolver.resolve(target) {
                            Ok(resp) => {
                                // Side effect: Register mount if resolved
                                if resp.status == crate::protocol::types::ResolveStatus::Resolved {
                                    if let (Some(root), Some(rid)) = (&resp.root, &resp.resolved_id)
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

                                // Return structured JSON + Text (JSON FIRST)
                                let json_val = serde_json::to_value(&resp).unwrap();
                                let text_val = serde_json::to_string(&resp).unwrap();

                                // Direct result object
                                json_rpc_ok(
                                    req.id.clone(),
                                    json!({
                                        "content": [
                                            { "type": "json", "json": json_val },
                                            { "type": "text", "text": text_val }
                                        ]
                                    }),
                                )
                            }
                            Err(e) => json_rpc_error(
                                req.id.clone(),
                                -32603,
                                &format!("Resolution failed: {}", e),
                            ),
                        }
                    }
                    "list_mounts" => {
                        if !args_is_empty {
                            return json_rpc_error(
                                req.id.clone(),
                                -32602,
                                "Invalid params: arguments must be an empty object",
                            );
                        }
                        let list = self.mounts.list();
                        json_rpc_ok(
                            req.id.clone(),
                            json!({
                                "content": [{ "type": "json", "json": list }]
                            }),
                        )
                    }
                    "describe_skills" => {
                        if !args_is_empty {
                            return json_rpc_error(
                                req.id.clone(),
                                -32602,
                                "Invalid params: arguments must be an empty object",
                            );
                        }
                        let skills = json!({
                            "repo.read": { "methods": ["resolve_mcp"], "notes": "Resolve repo names to local paths" },
                            "format:gofumpt": { "notes": "Planned" },
                            "governance.audit": { "notes": "Planned" }
                        });
                        json_rpc_ok(
                            req.id.clone(),
                            json!({
                                "content": [{ "type": "json", "json": skills }]
                            }),
                        )
                    }
                    "get_capabilities" => {
                        if !args_is_empty {
                            return json_rpc_error(
                                req.id.clone(),
                                -32602,
                                "Invalid params: arguments must be an empty object",
                            );
                        }
                        // Return the same capabilities object we use for initialize
                        // plus any extended metadata
                        let caps = json!({
                            "name": "cortex-mcp",
                            "server_capabilities": get_server_capabilities(),
                            "protocol": { "type": "mcp-router", "supports_dynamic_mounts": true, "supports_aliases": true },
                            "resolution": { "default_order": ["alias_map", "git_remote_match", "folder_name_match"] }
                        });
                        json_rpc_ok(
                            req.id.clone(),
                            json!({
                                "content": [{ "type": "json", "json": caps }]
                            }),
                        )
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
