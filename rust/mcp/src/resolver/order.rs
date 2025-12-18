use crate::io::Fs;
use crate::protocol::types::{ResolveResponse, ResolveStatus};
use crate::resolver::{alias_map, git_remote, workspace};
use anyhow::Result;
use std::path::PathBuf;

pub struct ResolveEngine<F: Fs> {
    pub fs: F,
    pub workspace_roots: Vec<PathBuf>,
}

impl<F: Fs> ResolveEngine<F> {
    pub fn new(fs: F, workspace_roots: Vec<PathBuf>) -> Self {
        Self {
            fs,
            workspace_roots,
        }
    }

    pub fn resolve(&self, name: &str) -> Result<ResolveResponse> {
        let mut tried: Vec<String> = Vec::new();

        // 1) alias_map
        tried.push("alias_map".to_string());
        if let Some(alias_path) = alias_map::default_aliases_path() {
            let map = alias_map::load_alias_map(&self.fs, &alias_path)?;
            if let Some(p) = alias_map::resolve_alias(&map, name) {
                let root = self.fs.canonicalize(&p).unwrap_or(p);

                // SECURITY: Ensure the aliased path is within one of the allowed workspace roots.
                let mut allowed = false;
                for ws_root in &self.workspace_roots {
                    // We check if the 'root' path starts with the workspace root.
                    // We use the canonicalized version for safety.
                    if root.starts_with(ws_root) {
                        allowed = true;
                        break;
                    }
                }

                if allowed {
                    return Ok(ResolveResponse {
                        status: ResolveStatus::Resolved,
                        resolved_id: Some(format!("cortex.repo.{}", name.replace('/', "."))),
                        kind: Some("local".to_string()),
                        root: Some(root.to_string_lossy().to_string()),
                        capabilities: vec![
                            "read_file".to_string(),
                            "list_files".to_string(),
                            "search".to_string(),
                        ],
                        tried,
                        fix_hint: None,
                    });
                } else {
                    // Intentionally ignore this alias if it violates security boundaries.
                    // Ideally we would log a warning here (eprintln or log crate), but to keep
                    // this logic pure we simply treat it as not found and proceed to other candidates.
                }
            }
        }

        let candidates = workspace::list_repo_candidates(&self.fs, &self.workspace_roots)?;

        // 2) git_remote_match
        tried.push("git_remote_match".to_string());
        let mut remote_matches: Vec<PathBuf> = Vec::new();
        for repo_dir in &candidates {
            if let Some(url) = git_remote::read_origin_url(&self.fs, repo_dir)? {
                if git_remote::remote_matches(name, &url) {
                    remote_matches.push(repo_dir.clone());
                }
            }
        }
        remote_matches.sort_by(|a, b| a.to_string_lossy().cmp(&b.to_string_lossy()));
        if let Some(winner) = remote_matches.first() {
            let root = self.fs.canonicalize(winner).unwrap_or(winner.clone());
            return Ok(ResolveResponse {
                status: ResolveStatus::Resolved,
                resolved_id: Some(format!("cortex.repo.{}", name.replace('/', "."))),
                kind: Some("local".to_string()),
                root: Some(root.to_string_lossy().to_string()),
                capabilities: vec![
                    "read_file".to_string(),
                    "list_files".to_string(),
                    "search".to_string(),
                ],
                tried,
                fix_hint: None,
            });
        }

        // 3) folder_name_match
        tried.push("folder_name_match".to_string());
        let mut name_matches: Vec<PathBuf> = Vec::new();
        for repo_dir in &candidates {
            if workspace::folder_name_matches(name, repo_dir) {
                name_matches.push(repo_dir.clone());
            }
        }
        name_matches.sort_by(|a, b| a.to_string_lossy().cmp(&b.to_string_lossy()));
        if let Some(winner) = name_matches.first() {
            let root = self.fs.canonicalize(winner).unwrap_or(winner.clone());
            return Ok(ResolveResponse {
                status: ResolveStatus::Resolved,
                resolved_id: Some(format!("cortex.repo.{}", name.replace('/', "."))),
                kind: Some("local".to_string()),
                root: Some(root.to_string_lossy().to_string()),
                capabilities: vec![
                    "read_file".to_string(),
                    "list_files".to_string(),
                    "search".to_string(),
                ],
                tried,
                fix_hint: None,
            });
        }

        Ok(ResolveResponse {
            status: ResolveStatus::Unresolved,
            resolved_id: None,
            kind: None,
            root: None,
            capabilities: vec![],
            tried,
            fix_hint: Some(format!(
                "Add an alias in ~/.cortex/mcp-aliases.json for '{}' to pin the correct path.",
                name
            )),
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::io::memfs::MemFs;
    use std::path::PathBuf;

    #[test]
    fn test_resolver_folder_name_match() {
        let fs = MemFs::new();
        fs.add_dir("/User/dev/cortex");

        // MemFs needs .git/config to be considered a repo by workspace scanner
        fs.add_file("/User/dev/cortex/.git/config", "");

        let roots = vec![PathBuf::from("/User/dev")];
        let engine = ResolveEngine::new(fs, roots);

        let resp = engine.resolve("cortex").unwrap();
        assert_eq!(resp.status, ResolveStatus::Resolved);
        assert_eq!(resp.root.unwrap(), "/User/dev/cortex");
    }

    #[test]
    fn test_resolver_alias_security_containment() {
        let fs = MemFs::new();
        // Setup a fake home and a fake sensitive path
        fs.add_dir("/User/bart/Dev");
        fs.add_dir("/etc/secret");

        let roots = vec![PathBuf::from("/User/bart/Dev")];
        // Mock alias map pointing to outside root
        fs.add_file(
            "/User/bart/.cortex/mcp-aliases.json",
            r#"{
            "hack": "/etc/secret"
        }"#,
        );

        let engine = ResolveEngine::new(fs, roots);

        // Without the fix, this would resolve. With the fix, it should fail (return unresolved or ignore the alias)
        // Our logic treats disallowed alias as "not found" so it falls through to other matchers.
        // Since no other matchers will match "hack", it should return Unresolved or a fix_hint depending on logic.
        // Actually, if it falls through, it will eventually hit the "Unresolved" at the bottom.
        let resp = engine.resolve("hack").unwrap();

        // It should NOT be Resolved pointing to /etc/secret
        if resp.status == ResolveStatus::Resolved {
            assert_ne!(
                resp.root.unwrap(),
                "/etc/secret",
                "Security bypass: alias allowed outside workspace root!"
            );
        }

        // Ideally it's Unresolved
        assert_eq!(resp.status, ResolveStatus::Unresolved);
    }
}
