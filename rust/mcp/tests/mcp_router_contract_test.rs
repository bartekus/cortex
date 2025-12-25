use cortex_mcp::io::fs::RealFs;
use cortex_mcp::resolver::order::ResolveEngine;
use cortex_mcp::router::mounts::MountRegistry;
use cortex_mcp::router::{JsonRpcRequest, Router};
use serde_json::json;
use std::path::PathBuf;
use std::sync::Arc;

// Feature: MCP_ROUTER_CONTRACT
// Spec: spec/mcp/contract.md

#[test]
fn test_router_contract_routing() {
    let fs = RealFs;
    // Initialize without config for now
    // Initialize without config for now
    let resolver = Arc::new(ResolveEngine::new(fs, Vec::<PathBuf>::new()));
    let mounts = MountRegistry::new();

    // Tools
    use cortex_mcp::snapshot::{lease::LeaseStore, tools::SnapshotTools};
    use cortex_mcp::workspace::WorkspaceTools;

    // db_path removed
    // let db_path = std::env::temp_dir().join("cortex_test_db_router");
    let config = cortex_mcp::config::StorageConfig::default();
    let store = Arc::new(cortex_mcp::snapshot::store::Store::new(config).unwrap());
    let lease_store = Arc::new(LeaseStore::new());

    let snapshot_tools = Arc::new(SnapshotTools::new(lease_store.clone(), store.clone()));
    let workspace_tools = Arc::new(WorkspaceTools::new(lease_store.clone(), store.clone()));

    let router = Router::new(resolver, mounts, snapshot_tools, workspace_tools);

    // 1. Unknown Method -> Error -32601
    let req = JsonRpcRequest {
        jsonrpc: "2.0".to_string(),
        method: "unknown/method".to_string(),
        params: None,
        id: Some(json!(1)),
    };
    let resp = router.handle_request(&req);
    assert!(resp.error.is_some());
    let err = resp.error.unwrap();
    assert_eq!(err["code"], -32601);

    // 2. initialize -> OK
    let req = JsonRpcRequest {
        jsonrpc: "2.0".to_string(),
        method: "initialize".to_string(),
        params: Some(json!({})),
        id: Some(json!(2)),
    };
    let resp = router.handle_request(&req);
    assert!(resp.result.is_some());
    let res = resp.result.unwrap();
    assert!(res["capabilities"].is_object());
    assert!(res["serverInfo"].is_object());
}
