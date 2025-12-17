use anyhow::{anyhow, Result};
use cortex_mcp::io::fs::RealFs;
use cortex_mcp::resolver::order::ResolveEngine;
use cortex_mcp::router::{JsonRpcRequest, Router};
use std::io::{self, BufRead, Read, Write};
use std::path::PathBuf;
use std::sync::Arc;

fn main() -> Result<()> {
    eprintln!("cortex-mcp starting (stdio - MCP framed JSON-RPC)");

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

    // 4. Stdio Loop (MCP framing)
    let stdin = io::stdin();
    let mut input = stdin.lock();
    let mut stdout = io::stdout();

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
                eprintln!("Failed to parse JSON-RPC payload: {}", e);
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
/// when the first non-empty line starts with '{'.
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
        return Ok(Some(trimmed.to_string()));
    }

    // Otherwise, treat it as the start of headers.
    let mut content_length: Option<usize> = None;
    parse_header_line(trimmed, &mut content_length);

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

        parse_header_line(l, &mut content_length);
    }

    let len = content_length.ok_or_else(|| anyhow!("Missing Content-Length header"))?;

    let mut buf = vec![0u8; len];
    r.read_exact(&mut buf)?;

    let s = String::from_utf8(buf)?;
    Ok(Some(s))
}

fn parse_header_line(line: &str, content_length: &mut Option<usize>) {
    // Keep this deterministic and strict.
    // We accept both `Content-Length:` and `content-length:`.
    let lower = line.to_ascii_lowercase();
    if let Some(rest) = lower.strip_prefix("content-length:") {
        let v = rest.trim();
        if let Ok(n) = v.parse::<usize>() {
            *content_length = Some(n);
        }
    }
}

fn write_mcp_message<W: Write>(w: &mut W, payload: &[u8]) -> Result<()> {
    // MCP stdio framing
    write!(w, "Content-Length: {}\r\n\r\n", payload.len())?;
    w.write_all(payload)?;
    w.flush()?;
    Ok(())
}
