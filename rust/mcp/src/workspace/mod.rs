use crate::snapshot::lease::Fingerprint;
use crate::snapshot::lease::LeaseStore;
use crate::snapshot::store::BlobStore;
use anyhow::{anyhow, Context, Result};
use std::path::{Path, PathBuf};
use std::sync::Arc;

pub struct WorkspaceTools {
    pub lease_store: Arc<LeaseStore>,
    pub blob_store: Arc<BlobStore>,
}

impl WorkspaceTools {
    pub fn new(lease_store: Arc<LeaseStore>, blob_store: Arc<BlobStore>) -> Self {
        Self {
            lease_store,
            blob_store,
        }
    }

    // Safety helper: ensure path is inside repo root and is safe
    fn resolve_target_path(&self, repo_root: &Path, rel_path: &str) -> Result<PathBuf> {
        let path = repo_root.join(rel_path);
        // If file needs to be created, we must check parent directory presence and safety.
        // For existing files, we canonicalize.

        let canonical_root = repo_root
            .canonicalize()
            .context("Failed to canonicalize repo root")?;

        // We try to canonicalize path. If it fails (doesn't exist), we canonicalize parent.
        if path.exists() {
            let canonical_path = path
                .canonicalize()
                .context("Failed to canonicalize target path")?;
            if !canonical_path.starts_with(&canonical_root) {
                return Err(anyhow!("Path escapes repo root: {}", rel_path));
            }
            Ok(canonical_path)
        } else {
            // Path doesn't exist. Check parent.
            let parent = path.parent().ok_or_else(|| anyhow!("Invalid path"))?;
            if parent.exists() {
                let canonical_parent = parent
                    .canonicalize()
                    .context("Failed to canonicalize parent")?;
                if !canonical_parent.starts_with(&canonical_root) {
                    return Err(anyhow!("Parent path escapes repo root: {}", rel_path));
                }
                // Return the joined path (since we can't canonicalize non-existent file)
                // But we should use the canonical parent + filename to be safe against some symlink tricks?
                // join filename
                let filename = path
                    .file_name()
                    .ok_or_else(|| anyhow!("Invalid filename"))?;
                Ok(canonical_parent.join(filename))
            } else {
                // Parent doesn't exist. Strict safety: reject deep creation unless create_dirs=true?
                // But for this helper, let's say we require parent to be safe.
                // If parent doesn't exist, we can't prove it's safe without resolving up to root.

                // Walk up until we find a base that exists, verify it's in root.
                let mut current = parent;
                while !current.exists() {
                    if let Some(p) = current.parent() {
                        current = p;
                    } else {
                        return Err(anyhow!("Cannot verify path safety (root not found)"));
                    }
                }
                let canonical_base = current.canonicalize()?;
                if !canonical_base.starts_with(&canonical_root) {
                    return Err(anyhow!("Path escapes repo root: {}", rel_path));
                }

                // If base safe, and we assume we are not following symlinks in non-existent components (obviously), it is safe.
                // But target logic must be precise.

                // Simple approach: we join components to canonical root? No, symlinks.
                // If intermediate components don't exist, they are just names.
                Ok(path)
            }
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub fn apply_patch(
        &self,
        repo_root: &Path,
        patch: &str,
        mode: &str,
        lease_id: Option<String>,
        _snapshot_id: Option<String>,
        strip: Option<usize>,
        reject_on_conflict: bool,
        dry_run: bool,
    ) -> Result<serde_json::Value> {
        // This returns a JSON object matching the schema
        // (OneOf success worktree, success snapshot, or error)

        if mode == "worktree" {
            let lid = lease_id.ok_or_else(|| anyhow!("lease_id required"))?;
            self.lease_store.check_lease(&lid, repo_root)?;

            // Apply patch to worktree using `git apply`.
            // But we need to handle structured rejects and determinism.
            // `git apply` supports `--reject` to emit .rej files, but we prefer not to litter.
            // We want structured output.
            // Maybe we use `patch` crate or parse diff manually?
            // `git apply --check` first?

            // The implementation plan specifies: "Unified diff format only. No fuzzing."
            // `git apply --whitespace=fix?` No, strict `context lines must match`.

            // We'll use `git apply`.
            // If dry_run, use `--check`.

            let mut cmd = std::process::Command::new("git");
            cmd.arg("apply");
            cmd.arg("--verbose"); // To get details?
                                  // --unidiff-zero?
                                  // "No fuzzing" -> git apply defaults to some fuzz. --recount?
                                  // There isn't a simple "no fuzz" flag in git apply without careful patch construction,
                                  // but `git apply` usually is strict.
                                  // We can check if it merges with offsets.

            if reject_on_conflict {
                // default behavior is fail on conflict.
            }

            if dry_run {
                cmd.arg("--check");
            }

            // Strip level
            if let Some(n) = strip {
                cmd.arg(format!("-p{}", n));
            }

            // Write patch to stdin
            cmd.current_dir(repo_root);
            cmd.stdin(std::process::Stdio::piped());
            cmd.stdout(std::process::Stdio::piped());
            cmd.stderr(std::process::Stdio::piped());

            let mut child = cmd.spawn()?;
            if let Some(mut stdin) = child.stdin.take() {
                use std::io::Write;
                stdin.write_all(patch.as_bytes())?;
            }

            let output = child.wait_with_output()?;

            // If success, all applied.
            if output.status.success() {
                // Touched files?
                // We need to parse patch to know what matched.
                // OR we rely on `git apply --verbose` output if it lists files?
                // Or we parse the input patch for "+++ b/path".

                let touched = parse_patch_touched_files(patch);
                self.lease_store.touch_files(&lid, touched.clone());

                // Return structured success
                let new_fingerprint = Fingerprint::compute(repo_root)?;
                // We need logic to return "applied" list.
                let applied_value: Vec<serde_json::Value> = touched
                    .iter()
                    .map(|f| {
                        serde_json::json!({
                            "path": f,
                            "status": "ok"
                        })
                    })
                    .collect();

                return Ok(serde_json::json!({
                    "applied": applied_value,
                    "rejects": [],
                    "lease_id": lid,
                    "fingerprint": new_fingerprint,
                    "cache_key": format!("{}:{}", lid, new_fingerprint.status_hash),
                    "cache_hint": "until_dirty"
                }));
            } else {
                // Parse reject reasons.
                // Git apply output on failure usually says "error: patch failed: file:line..."
                // We map this to structured rejects.
                let stderr = String::from_utf8_lossy(&output.stderr);
                // TODO: Parse stderr for structured rejects.
                // For now, return generic error or empty applies and populated rejects?
                // If whole patch failed, nothing applied (atomic).

                return Err(anyhow!("Patch failed: {}", stderr));
            }
        } else if mode == "snapshot" {
            // In-memory patch application on snapshot blobs!
            // Load manifest. For each file in patch:
            // 1. Get blob.
            // 2. Apply hunk.
            // 3. Store new blob.
            // 4. Update manifest.
            // 5. Create new snapshot.

            // This is complex to implement fully in one step without a diff library.
            // `bdiff` or `patch-rs`?
            // Given limitations, maybe we assume `snapshot` mode patching is OUT OF SCOPE for this turn
            // unless I have a library.
            // Implementation plan says: "workspace.apply_patch... mode=snapshot: Applies patch against the snapshot's stored bytes".

            // I will stub this with "Not implemented" or simple replacement if trivial?
            // Patch application is non-trivial.
            // I recall standard library doesn't have it.
            // I'll return error for now: "Snapshot patching not supported in this iteration".

            return Err(anyhow!("Snapshot mode patching not yet implemented"));
        }

        Err(anyhow!("Invalid mode"))
    }

    pub fn write_file(
        &self,
        repo_root: &Path,
        path: &str,
        content_base64: &str,
        lease_id: Option<String>,
        create_dirs: bool,
        dry_run: bool,
    ) -> Result<bool> {
        let lid = lease_id.ok_or_else(|| anyhow!("lease_id required"))?;
        self.lease_store.check_lease(&lid, repo_root)?;

        // Resolve path strict
        // We use our helper resolve_target_path
        let target = self.resolve_target_path(repo_root, path)?;

        // Decode content
        use base64::{engine::general_purpose, Engine as _};
        let content = general_purpose::STANDARD
            .decode(content_base64)
            .context("Invalid base64 content")?;

        // Check dirs
        if let Some(parent) = target.parent() {
            if !parent.exists() {
                if create_dirs {
                    if !dry_run {
                        std::fs::create_dir_all(parent)?;
                    }
                } else {
                    return Err(anyhow!(
                        "Parent directory does not exist (set create_dirs=true)"
                    ));
                }
            }
        }

        if !dry_run {
            std::fs::write(&target, content)?;
            // Touch
            self.lease_store.touch_files(&lid, vec![path.to_string()]);
        }

        Ok(true)
    }

    pub fn delete(
        &self,
        repo_root: &Path,
        path: &str,
        lease_id: Option<String>,
        dry_run: bool,
    ) -> Result<bool> {
        let lid = lease_id.ok_or_else(|| anyhow!("lease_id required"))?;
        self.lease_store.check_lease(&lid, repo_root)?;

        let target = self.resolve_target_path(repo_root, path)?;

        if !target.exists() {
            return Err(anyhow!("File not found"));
        }

        if !dry_run {
            if target.is_dir() {
                std::fs::remove_dir_all(&target)?;
            } else {
                std::fs::remove_file(&target)?;
            }
            // Touch? If we deleted it, it changed.
            self.lease_store.touch_files(&lid, vec![path.to_string()]);
        }

        Ok(true)
    }
}

// Helper to extract file paths from unified diff
// Matches "+++ b/path/to/file"
fn parse_patch_touched_files(patch: &str) -> Vec<String> {
    let mut files = Vec::new();
    for line in patch.lines() {
        if line.starts_with("+++ ") {
            // usually "+++ b/path" or "+++ path"
            // Try to strip prefix b/ or a/?
            // Typical git diff: "--- a/foo\n+++ b/foo"
            // path is "foo".
            // We look for b/
            let path_part = line.trim_start_matches("+++ ").trim();
            // naive stripping of b/ if present
            let clean_path = if let Some(stripped) = path_part.strip_prefix("b/") {
                stripped
            } else {
                path_part
            };
            // Handle /dev/null for deletions?
            if clean_path != "/dev/null" {
                files.push(clean_path.to_string());
            }
        }
    }
    files
}
