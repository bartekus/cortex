use crate::io::Fs;
use anyhow::Result;
use std::path::Path;

/// Read remote origin URL from `.git/config` if present.
pub fn read_origin_url<F: Fs>(fs: &F, repo_dir: &Path) -> Result<Option<String>> {
    let cfg_path = repo_dir.join(".git").join("config");
    if !fs.exists(&cfg_path) {
        return Ok(None);
    }
    let cfg = fs.read_to_string(&cfg_path)?;
    Ok(parse_origin_url(&cfg))
}

fn parse_origin_url(cfg: &str) -> Option<String> {
    // Small INI-like parse. Deterministic.
    let mut in_origin = false;
    for line in cfg.lines() {
        let l = line.trim();
        if l.starts_with('[') && l.ends_with(']') {
            in_origin = l == "[remote \"origin\"]";
            continue;
        }
        if !in_origin {
            continue;
        }
        if let Some(rest) = l.strip_prefix("url") {
            let rest = rest.trim_start();
            if let Some(v) = rest.strip_prefix('=') {
                return Some(v.trim().to_string());
            }
        }
    }
    None
}

/// Returns true if a remote URL corresponds to the requested name.
/// Supports common GitHub URL forms:
/// - https://github.com/OWNER/REPO(.git)
/// - git@github.com:OWNER/REPO(.git)
pub fn remote_matches(name: &str, remote_url: &str) -> bool {
    let requested = name.trim();
    if requested.is_empty() {
        return false;
    }

    let (req_owner, req_repo) = split_owner_repo(requested);
    let (rem_owner, rem_repo) = extract_owner_repo(remote_url);

    match (req_owner, req_repo, rem_owner, rem_repo) {
        (Some(ro), rr, Some(so), sr) => ro == so && rr == sr,
        (None, rr, _, sr) => rr == sr,
        _ => false,
    }
}

fn split_owner_repo(name: &str) -> (Option<String>, String) {
    if let Some((o, r)) = name.split_once('/') {
        return (Some(o.to_string()), r.to_string());
    }
    (None, name.to_string())
}

fn extract_owner_repo(remote_url: &str) -> (Option<String>, String) {
    if let Some(after) = remote_url.split("github.com:").nth(1) {
        return normalize_owner_repo(after.trim());
    }
    if let Some(after) = remote_url.split("github.com/").nth(1) {
        return normalize_owner_repo(after.trim());
    }

    let repo = remote_url
        .rsplit('/')
        .next()
        .unwrap_or(remote_url)
        .trim()
        .trim_end_matches(".git")
        .to_string();
    (None, repo)
}

fn normalize_owner_repo(s: &str) -> (Option<String>, String) {
    let cleaned = s.trim().trim_end_matches(".git");
    if let Some((o, r)) = cleaned.split_once('/') {
        return (Some(o.to_string()), r.to_string());
    }
    (None, cleaned.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn remote_matches_owner_repo_https() {
        assert!(remote_matches(
            "bartekus/cortex",
            "https://github.com/bartekus/cortex.git"
        ));
    }

    #[test]
    fn remote_matches_owner_repo_ssh() {
        assert!(remote_matches(
            "bartekus/cortex",
            "git@github.com:bartekus/cortex.git"
        ));
    }

    #[test]
    fn remote_matches_repo_only() {
        assert!(remote_matches(
            "cortex",
            "https://github.com/bartekus/cortex.git"
        ));
    }

    #[test]
    fn remote_mismatch() {
        assert!(!remote_matches(
            "bartekus/cortex",
            "https://github.com/other/cortex.git"
        ));
    }
}
