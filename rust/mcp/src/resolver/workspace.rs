use crate::io::Fs;
use anyhow::Result;
use std::path::{Path, PathBuf};

/// List candidate repo directories under workspace roots.
///
/// Deterministic rules:
/// - Only direct children of each root are considered.
/// - Candidates must be directories.
/// - A directory is considered a repo candidate if it contains a `.git/config` file.
pub fn list_repo_candidates<F: Fs>(fs: &F, workspace_roots: &[PathBuf]) -> Result<Vec<PathBuf>> {
    let mut out: Vec<PathBuf> = Vec::new();

    for root in workspace_roots {
        let entries = fs.read_dir(root)?;
        for p in entries {
            if !fs.is_dir(&p) {
                continue;
            }
            let git_config = p.join(".git").join("config");
            if fs.exists(&git_config) {
                out.push(p);
            }
        }
    }

    out.sort_by(|a, b| a.to_string_lossy().cmp(&b.to_string_lossy()));
    out.dedup();
    Ok(out)
}

pub fn folder_name_matches(name: &str, path: &Path) -> bool {
    path.file_name()
        .map(|n| n.to_string_lossy() == name)
        .unwrap_or(false)
}
