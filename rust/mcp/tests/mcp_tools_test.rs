use cortex_mcp::io::fs::RealFs;
use cortex_mcp::resolver::order::ResolveEngine;
use cortex_mcp::router::mounts::MountRegistry;
use cortex_mcp::router::{JsonRpcRequest, Router};
use cortex_mcp::snapshot::{lease::LeaseStore, store::BlobStore, tools::SnapshotTools};
use cortex_mcp::workspace::WorkspaceTools;
use serde_json::json;
use std::path::PathBuf;
use std::sync::Arc;

// Feature: MCP_TOOLS
// Spec: spec/mcp/tools.md

#[test]
fn test_mcp_tools_list() {
    let fs = RealFs;
    let resolver = Arc::new(ResolveEngine::new(fs, Vec::<PathBuf>::new()));
    let mounts = MountRegistry::new();

    let blob_store = Arc::new(BlobStore::new());
    let lease_store = Arc::new(LeaseStore::new());

    let snapshot_tools = Arc::new(SnapshotTools::new(lease_store.clone(), blob_store.clone()));
    let workspace_tools = Arc::new(WorkspaceTools::new(lease_store.clone(), blob_store.clone()));

    let router = Router::new(resolver, mounts, snapshot_tools, workspace_tools);

    // Test tools/list
    let req = JsonRpcRequest {
        jsonrpc: "2.0".to_string(),
        method: "tools/list".to_string(),
        params: None,
        id: Some(json!(1)),
    };
    let resp = router.handle_request(&req);
    assert!(resp.result.is_some());
    let res = resp.result.unwrap();

    let tools = res["tools"].as_array().expect("tools should be an array");

    // Check for required tools
    let required_tools = vec!["resolve_mcp", "list_mounts"];
    for req_tool in required_tools {
        let found = tools.iter().any(|t| t["name"] == req_tool);
        assert!(found, "Tool {} not found", req_tool);
    }
}

#[test]
fn test_mcp_tools_call_validation() {
    let fs = RealFs;
    let resolver = Arc::new(ResolveEngine::new(fs, Vec::<PathBuf>::new()));
    let mounts = MountRegistry::new();

    let blob_store = Arc::new(BlobStore::new());
    let lease_store = Arc::new(LeaseStore::new());

    let snapshot_tools = Arc::new(SnapshotTools::new(lease_store.clone(), blob_store.clone()));
    let workspace_tools = Arc::new(WorkspaceTools::new(lease_store.clone(), blob_store.clone()));

    let router = Router::new(resolver, mounts, snapshot_tools, workspace_tools);

    // Call resolve_mcp without name -> Error
    let req = JsonRpcRequest {
        jsonrpc: "2.0".to_string(),
        method: "tools/call".to_string(),
        params: Some(json!({
            "name": "resolve_mcp",
            "arguments": {}
        })),
        id: Some(json!(2)),
    };
    let resp = router.handle_request(&req);
    assert!(resp.error.is_some());
    // Expect error about missing argument
}
