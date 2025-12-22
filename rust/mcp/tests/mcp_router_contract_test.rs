use cortex_mcp::router::{Router, JsonRpcRequest};
use cortex_mcp::resolver::order::ResolveEngine;
use cortex_mcp::router::mounts::MountRegistry;
use cortex_mcp::io::fs::RealFs;
use std::sync::Arc;
use serde_json::json;

// Feature: MCP_ROUTER_CONTRACT
// Spec: spec/mcp/contract.md

#[test]
fn test_router_contract_routing() {
    let fs = Arc::new(RealFs);
    // Initialize without config for now
    let resolver = Arc::new(ResolveEngine::new(fs, None));
    let mounts = MountRegistry::new();
    let router = Router::new(resolver, mounts);

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
