use anyhow::Result;
use cortex_mcp::io::fs::RealFs;
use cortex_mcp::resolver::order::ResolveEngine;
use cortex_mcp::router::{Router, JsonRpcRequest};
use std::io::{self, BufRead, Write};
use std::path::PathBuf;
use std::sync::Arc;

fn main() -> Result<()> {
    eprintln!("cortex-mcp starting (stdio - JSON-RPC fallback)");

    // 1. Setup Resolver
    let dirs = match std::env::var("CORTEX_WORKSPACE_ROOTS") {
        Ok(val) => val.split(':').map(PathBuf::from).collect(),
        Err(_) => {
            // Default roots
            let mut roots = Vec::new();
            if let Some(home) = dirs::home_dir() {
                roots.push(home.join("Dev"));
                roots.push(home.join("src"));
            }
            roots
        }
    };
    
    let fs = RealFs;
    let resolver = Arc::new(ResolveEngine::new(fs, dirs));

    // 2. Setup MountRegistry
    let mounts = cortex_mcp::router::mounts::MountRegistry::new();

    // 3. Setup Router
    let router = Router::new(resolver, mounts);

    // 4. Stdio Loop
    let stdin = io::stdin();
    let lock = stdin.lock();
    let mut stdout = io::stdout();

    for line_res in lock.lines() {
        let line = line_res?;
        if line.trim().is_empty() { continue; }

        match serde_json::from_str::<JsonRpcRequest>(&line) {
            Ok(req) => {
                let response = router.handle_request(&req);
                let resp_str = serde_json::to_string(&response)?;
                writeln!(stdout, "{}", resp_str)?;
                stdout.flush()?;
            }
            Err(e) => {
                eprintln!("Failed to parse JSON-RPC: {}", e);
            }
        }
    }

    Ok(())
}
