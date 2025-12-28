use cortex_mcp::io::fs::RealFs;
use cortex_mcp::resolver::order::ResolveEngine;
use cortex_mcp::router::mounts::MountRegistry;
use cortex_mcp::router::{JsonRpcRequest, Router};
use cortex_mcp::snapshot::{lease::LeaseStore, tools::SnapshotTools};
use cortex_mcp::workspace::WorkspaceTools;
use serde_json::json;
use std::path::PathBuf;
use std::process::Command;
use std::sync::Arc;
use tempfile::TempDir;

fn setup_repo() -> TempDir {
    let dir = tempfile::tempdir().unwrap();
    let root = dir.path();

    // Ignore output to avoid clogging test logs
    let _ = Command::new("git")
        .arg("init")
        .current_dir(root)
        .output()
        .unwrap();
    let _ = Command::new("git")
        .arg("config")
        .arg("user.email")
        .arg("test@example.com")
        .current_dir(root)
        .output()
        .unwrap();
    let _ = Command::new("git")
        .arg("config")
        .arg("user.name")
        .arg("Test")
        .current_dir(root)
        .output()
        .unwrap();

    // Create initial commit
    std::fs::write(root.join("file.txt"), "initial").unwrap();
    // Use proper git commands
    let _ = Command::new("git")
        .args(["add", "."])
        .current_dir(root)
        .output()
        .unwrap();
    let _ = Command::new("git")
        .args(["commit", "-m", "initial"])
        .current_dir(root)
        .output()
        .unwrap();

    dir
}

#[test]
fn test_stale_lease_error_structure() {
    let repo = setup_repo();
    let repo_path = repo.path().to_str().unwrap();

    let fs = RealFs;
    let resolver = Arc::new(ResolveEngine::new(fs, Vec::<PathBuf>::new()));
    let mounts = MountRegistry::new();

    let db_dir = tempfile::tempdir().unwrap();
    let config = cortex_mcp::config::StorageConfig {
        data_dir: db_dir.path().to_path_buf(),
        blob_backend: cortex_mcp::config::BlobBackend::Fs,
        compression: cortex_mcp::config::Compression::None,
    };
    let store = Arc::new(cortex_mcp::snapshot::store::Store::new(config).unwrap());
    let lease_store = Arc::new(LeaseStore::new());

    let snapshot_tools = Arc::new(SnapshotTools::new(lease_store.clone(), store.clone()));
    let workspace_tools = Arc::new(WorkspaceTools::new(lease_store.clone(), store.clone()));

    let router = Router::new(resolver, mounts, snapshot_tools, workspace_tools);

    // 1. Get a lease via snapshot.list (worktree mode)
    let req = JsonRpcRequest {
        jsonrpc: "2.0".to_string(),
        method: "tools/call".to_string(),
        params: Some(json!({
            "name": "snapshot.list",
            "arguments": {
                "repo_root": repo_path,
                "path": ".",
                "mode": "worktree"
            }
        })),
        id: Some(json!(1)),
    };
    let resp = router.handle_request(&req);

    // Debug output if fails
    if let Some(err) = &resp.error {
        println!("Initial request failed: {:?}", err);
    }

    assert!(resp.result.is_some(), "Initial request failed");
    let res = resp.result.unwrap();
    let content = &res["content"][0]["json"];

    // Worktree mode returns lease_id at top level of result?
    // Check snapshot.list response schema.
    // It returns { "entries": [...], "lease_id": "...", "fingerprint": ... }

    let lease_id_val = content.get("lease_id");
    if lease_id_val.is_none() {
        println!("Response content: {:?}", content);
        panic!("lease_id missing from response");
    }
    let lease_id = lease_id_val
        .unwrap()
        .as_str()
        .expect("lease_id strings")
        .to_string();

    // 2. Modify repo (commit a change to change HEAD oid)
    std::fs::write(repo.path().join("new_file.txt"), "modified").unwrap();
    let _ = Command::new("git")
        .args(["add", "."])
        .current_dir(repo.path())
        .output()
        .unwrap();
    let _ = Command::new("git")
        .args(["commit", "-m", "update"])
        .current_dir(repo.path())
        .output()
        .unwrap();

    // 3. Call snapshot.list with old lease
    let req2 = JsonRpcRequest {
        jsonrpc: "2.0".to_string(),
        method: "tools/call".to_string(),
        params: Some(json!({
            "name": "snapshot.list",
            "arguments": {
                "repo_root": repo_path,
                "path": ".",
                "mode": "worktree",
                "lease_id": lease_id
            }
        })),
        id: Some(json!(2)),
    };
    let resp2 = router.handle_request(&req2);

    // 4. Expect Error
    let err = resp2.error.expect("Should return error for stale lease");
    println!("Error details: {:?}", err);

    assert_eq!(err["code"], "STALE_LEASE");
    assert_eq!(err["message"], "Lease is stale (repo changed)");
    let data = &err["data"];
    assert!(data["current_fingerprint"].is_object());
    assert!(data["current_fingerprint"]["head_oid"].is_string());
    assert_eq!(data["lease_id"], lease_id);
}
