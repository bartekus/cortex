use anyhow::{anyhow, Result};
use cortex_mcp::io::fs::RealFs;
use cortex_mcp::resolver::order::ResolveEngine;
use cortex_mcp::router::{JsonRpcRequest, Router};
use env_logger::Target;
use std::io::{self, BufRead, Read, Write};
use std::path::PathBuf;
use std::sync::Arc;

// Feature: MCP_ROUTER_CONTRACT
// Spec: spec/mcp/contract.md

// POLICY: stdout is RESERVED for protocol messages.
// All logs, panics, and diagnostics MUST write to stderr.
fn main() -> Result<()> {
    // 0. Setup Logging & Panic Safety
    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or("info"))
        .target(Target::Stderr)
        .format_timestamp(None) // Stable tests
        .init();

    std::panic::set_hook(Box::new(|info| {
        log::error!("Panic: {}", info);
    }));

    log::info!("cortex-mcp starting (stdio - MCP framed JSON-RPC)");

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

    // 2. Setup Stores & Tools
    // Persistent store
    let config = cortex_mcp::config::StorageConfig::default();
    log::info!(
        "[cortex-mcp] storage data_dir: {}",
        config.data_dir.display()
    );
    let store = Arc::new(cortex_mcp::snapshot::store::Store::new(config)?);

    // LeaseStore (currently in-memory, Option A)
    let lease_store = Arc::new(cortex_mcp::snapshot::lease::LeaseStore::new());

    let snapshot_tools = Arc::new(cortex_mcp::snapshot::tools::SnapshotTools::new(
        lease_store.clone(),
        store.clone(),
    ));

    let workspace_tools = Arc::new(cortex_mcp::workspace::WorkspaceTools::new(
        lease_store.clone(),
        store.clone(),
    ));

    // 3. Setup MountRegistry
    let mounts = cortex_mcp::router::mounts::MountRegistry::new();

    // 4. Setup Router
    let router = Router::new(resolver, mounts, snapshot_tools, workspace_tools);

    // 4. Stdio Loop (MCP framing)
    let stdin = io::stdin();
    let mut input = stdin.lock();
    let stdout = io::stdout();
    let mut stdout = stdout.lock();

    loop {
        let maybe_payload = read_mcp_message(&mut input)?;
        let Some(payload) = maybe_payload else {
            break; // EOF
        };

        match serde_json::from_str::<JsonRpcRequest>(&payload) {
            Ok(req) => {
                let response = router.handle_request(&req);
                let resp_str = serde_json::to_string(&response)?;
                write_mcp_message(&mut stdout, resp_str.as_bytes())?;
            }
            Err(e) => {
                // IMPORTANT: Some clients will send other traffic; log but don't crash.
                log::error!("Failed to parse JSON-RPC payload: {}", e);
            }
        }
    }

    Ok(())
}

/// Reads a single MCP stdio framed message.
///
/// MCP clients typically speak:
///   Content-Length: <n>\r\n
///   \r\n
///   <n bytes of JSON>
///
/// For local diagnostics, we also accept a single-line JSON payload (line-delimited)
/// **IF AND ONLY IF** `CORTEX_MCP_ALLOW_LINE_JSON` is set.
fn read_mcp_message<R: BufRead + Read>(r: &mut R) -> Result<Option<String>> {
    let mut first_line = String::new();

    // Read until we find a non-empty line or EOF.
    loop {
        first_line.clear();
        let n = r.read_line(&mut first_line)?;
        if n == 0 {
            return Ok(None);
        }
        if !first_line.trim().is_empty() {
            break;
        }
    }

    let trimmed = first_line.trim_end_matches(['\r', '\n']);

    // Line-delimited JSON fallback for dev/testing.
    if trimmed.starts_with('{') {
        if std::env::var("CORTEX_MCP_ALLOW_LINE_JSON").is_ok() {
            return Ok(Some(trimmed.to_string()));
        }
        // If strict, we fall through and likely fail parsing headers, which is correct.
        // Or we could return an error here if strictness implies NO fallback.
        // Protocol specifies Content-Length. If it starts with {, it's likely not a header.
        // We will try to parse "{" as a header key/val and fail.
    }

    // Otherwise, treat it as the start of headers.
    let mut content_length: Option<usize> = None;
    parse_header_line(trimmed, &mut content_length)?;

    // Read remaining headers until blank line.
    loop {
        let mut line = String::new();
        let n = r.read_line(&mut line)?;
        if n == 0 {
            return Err(anyhow!("EOF while reading MCP headers"));
        }

        let l = line.trim_end_matches(['\r', '\n']);
        if l.is_empty() {
            break;
        }

        parse_header_line(l, &mut content_length)?;
    }

    let len = content_length.ok_or_else(|| anyhow!("Missing Content-Length header"))?;

    let mut buf = vec![0u8; len];
    r.read_exact(&mut buf)?;

    let s = String::from_utf8(buf)?;
    Ok(Some(s))
}

fn parse_header_line(line: &str, content_length: &mut Option<usize>) -> Result<()> {
    // Keep this deterministic and strict.
    // We accept both `Content-Length:` and `content-length:`.
    let lower = line.to_ascii_lowercase();
    if let Some(rest) = lower.strip_prefix("content-length:") {
        let v = rest.trim();
        if let Ok(n) = v.parse::<usize>() {
            *content_length = Some(n);
        } else {
            return Err(anyhow!("Invalid Content-Length value: {}", v));
        }
    }
    Ok(())
}

fn write_mcp_message<W: Write>(w: &mut W, payload: &[u8]) -> Result<()> {
    // MCP stdio framing
    write!(w, "Content-Length: {}\r\n\r\n", payload.len())?;
    w.write_all(payload)?;
    w.flush()?;
    Ok(())
}
