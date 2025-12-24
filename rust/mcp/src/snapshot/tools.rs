use super::lease::{Fingerprint, LeaseStore};
use super::store::{BlobStore, Manifest, ManifestEntry};
use anyhow::{anyhow, Context, Result};
use serde_json::Value;
use std::path::{Path, PathBuf};
use std::sync::Arc;

// Tool implementations will go here.
// We need struct definitions for arguments if we want strict typing,
// or we can parse generic JSON. Given we have schemas validation on the router side (assumed?),
// or we parse manually.
// For now, let's assume we receive parsed args or Value.

pub struct SnapshotTools {
    pub lease_store: Arc<LeaseStore>,
    pub blob_store: Arc<BlobStore>,
}

impl SnapshotTools {
    pub fn new(lease_store: Arc<LeaseStore>, blob_store: Arc<BlobStore>) -> Self {
        Self {
            lease_store,
            blob_store,
        }
    }

    // Helper to resolve paths - STRICT safety
    fn resolve_path(&self, repo_root: &Path, rel_path: &str) -> Result<PathBuf> {
        let path = repo_root.join(rel_path);
        // Canonicalize to ensure we don't escape root via .. or symlinks
        let canonical_root = repo_root
            .canonicalize()
            .context("Failed to canonicalize repo root")?;

        // If file doesn't exist, we can't canonicalize it directly.
        // But for snapshot.create usually files exist.
        // For general safety, we check if it starts with root.

        // For existing files:
        if path.exists() {
            let canonical_path = path
                .canonicalize()
                .context("Failed to canonicalize target path")?;
            if !canonical_path.starts_with(&canonical_root) {
                return Err(anyhow!("Path escapes repo root: {}", rel_path));
            }
            Ok(canonical_path)
        } else {
            // If not exists, strict check of the joined path components?
            // Or allow only existing? snapshot.create captures existing state.
            Err(anyhow!("Path does not exist: {}", rel_path))
        }
    }

    pub fn create_snapshot(
        &self,
        repo_root: &Path,
        lease_id: Option<String>,
        paths: Option<Vec<String>>,
    ) -> Result<String> {
        // 1. Determine paths to capture
        let paths_to_capture = if let Some(p) = paths {
            p
        } else if let Some(lid) = lease_id {
            // Get touched files from lease
            self.lease_store
                .get_touched_files(&lid)
                .ok_or_else(|| anyhow!("Lease not found: {}", lid))?
        } else {
            // If no lease and no paths, maybe capture nothing? or everything?
            // Implementation plan says: "Default snapshot.create scope is 'touched files from the lease' if paths omitted."
            // If neither, return error or empty snapshot?
            return Err(anyhow!("Must provide either lease_id or paths"));
        };

        let mut manifest_entries = Vec::new();

        for rel_path in paths_to_capture {
            let abs_path = self.resolve_path(repo_root, &rel_path)?;
            if abs_path.is_dir() {
                // Should walk dir? "If paths are provided, captures those specific paths."
                // If a directory is provided, we recursively capture?
                // Or strict?
                // Usually snapshot tools capture explicit files.
                // Let's assume recursion for dirs.
                for entry in walkdir::WalkDir::new(&abs_path) {
                    let entry = entry?;
                    if entry.file_type().is_file() {
                        let entry_path = entry.path();
                        // Rel path from repo root
                        let file_rel = entry_path
                            .strip_prefix(repo_root)?
                            .to_string_lossy()
                            .to_string();
                        // Normalize path separators if needed (Rust usually handles)
                        let content = std::fs::read(entry_path)?;
                        let blob_hash = self.blob_store.put(&content);
                        manifest_entries.push(ManifestEntry {
                            path: file_rel,
                            blob: blob_hash,
                        });
                    }
                }
            } else {
                let content = std::fs::read(&abs_path)?;
                let blob_hash = self.blob_store.put(&content);
                manifest_entries.push(ManifestEntry {
                    path: rel_path,
                    blob: blob_hash,
                });
            }
        }

        // Create manifest
        let manifest = Manifest::new(manifest_entries);
        let manifest_json = manifest.to_canonical_json()?;

        // Compute Snapshot ID
        // snapshot_id = sha256( fingerprint + "\n" + manifest )
        // We need fingerprint of the repo at this moment.
        let fingerprint = Fingerprint::compute(repo_root)?;
        let fingerprint_json = serde_json::to_string(&fingerprint)?; // This needs strict canonicalization ideally
                                                                     // But serde_json::to_string isn't guaranteed deterministic map order without sort_keys feature?
                                                                     // Wait, serialize_json depends on struct definition order. Fingerprint is a struct.
                                                                     // fields are serialized in order.

        use sha2::{Digest, Sha256};
        let mut hasher = Sha256::new();
        hasher.update(fingerprint_json.as_bytes());
        hasher.update(b"\n");
        hasher.update(&manifest_json);
        let snapshot_id = format!("sha256:{}", hex::encode(hasher.finalize()));

        // Store manifest? The BlobStore stores blobs. Where is manifest stored?
        // We probably need to store manifest in BlobStore too, under the snapshot_id (or just as a blob).
        // Spec says snapshot_id is content addressed.
        // We should put the manifest JSON in the store. Keyed by snapshot_id?
        // Wait, BlobStore puts by content hash.
        // We should map snapshot_id -> manifest_json manually?
        // Or put manifest in blob store, and snapshot_id implies checking blob store?
        // But snapshot_id is hash of (fp + manifest), not just manifest.
        // So we need a SnapshotStore or map.
        // Use BlobStore for now?
        // Actually, we can just put it in blob store with the snapshot_id as key directly (bypass hash check)?
        // Or separate map.

        // For simplicity/v1 scope:
        // We can just add a `put_exact` to BlobStore or store metadata.
        // Let's assuming we put it in BlobStore under `snapshot_id`.

        // Persist mapping snapshot_id -> manifest
        self.blob_store
            .put_exact(snapshot_id.clone(), manifest_json);

        Ok(snapshot_id)
    }

    pub fn list_snapshot(
        &self,
        repo_root: &Path,
        path: &str,
        mode: &str,
        lease_id: Option<String>,
        snapshot_id: Option<String>,
    ) -> Result<Value> {
        let mut entries = Vec::new();
        let target_path = self.resolve_path(repo_root, path)?;

        // Helper to format entry
        let format_entry = |sys_path: &Path, is_dir: bool| -> Result<Value> {
            let rel = sys_path
                .strip_prefix(repo_root)?
                .to_string_lossy()
                .to_string();
            Ok(serde_json::json!({
                "path": rel,
                "type": if is_dir { "dir" } else { "file" },
                // size and sha omitted for now or calculate? Schema says optional? No, checks spec.
                // Schema says required: path, type. size, sha optional.
            }))
        };

        if mode == "worktree" {
            // Lease Logic
            let (lid, fingerprint) = if let Some(l) = lease_id {
                let fp = self.lease_store.check_lease(&l, repo_root)?;
                (l, fp)
            } else {
                self.lease_store.issue_lease(repo_root)?
            };

            // List live files
            let mut touched = Vec::new();

            if target_path.exists() {
                if target_path.is_dir() {
                    for entry in walkdir::WalkDir::new(&target_path) {
                        let entry = entry?;
                        let ft = entry.file_type();
                        // Skip directories in output? Schema entries are files usually?
                        // Schema allows "dir".
                        // Spec says: "Lists files... Implicit parents."
                        // Usually walkdir yields files and dirs.
                        // But for now let's just list files as per previous logic.
                        if ft.is_file() {
                            let e_path = entry.path();
                            let p_str = e_path
                                .strip_prefix(repo_root)?
                                .to_string_lossy()
                                .to_string();
                            entries.push(format_entry(e_path, false)?);
                            touched.push(p_str);
                        }
                    }
                } else {
                    let p_str = target_path
                        .strip_prefix(repo_root)?
                        .to_string_lossy()
                        .to_string();
                    entries.push(format_entry(&target_path, false)?);
                    touched.push(p_str);
                }
            }

            self.lease_store.touch_files(&lid, touched);

            // Construct Response
            Ok(serde_json::json!({
                "snapshot_id": "", // Worktree mode doesn't imply snapshot id? Schema requires it?
                // Schema "snapshot.list worktree success" requires "snapshot_id".
                // Wait, if it's worktree, what snapshot_id?
                // Spec says "snapshot.list worktree success" -> check schema again.
                // Schema has "snapshot_id". Maybe it means HEAD snapshot id? Or empty?
                // Or maybe the schema assumes we return a "virtual" snapshot id?
                // Let's use empty string or "worktree".
                "path": path,
                "mode": "worktree",
                "entries": entries,
                "truncated": false,
                "lease_id": lid,
                "fingerprint": fingerprint, // Fingerprint struct serializes to object
                "cache_key": format!("{}:{}", lid, fingerprint.status_hash), // approximate
                "cache_hint": "until_dirty"
            }))
        } else if mode == "snapshot" {
            let sid =
                snapshot_id.ok_or_else(|| anyhow!("snapshot_id required for snapshot mode"))?;

            // Retrieve manifest
            let manifest_json = self
                .blob_store
                .get(&sid)
                .ok_or_else(|| anyhow!("Snapshot not found: {}", sid))?;
            let manifest: Manifest =
                serde_json::from_slice(&manifest_json).context("Failed to parse manifest")?;

            let normalized_prefix = if path == "." { "" } else { path };

            for entry in manifest.entries {
                let matches = if normalized_prefix.is_empty() {
                    true
                } else {
                    entry.path.starts_with(normalized_prefix)
                        && (entry.path.len() == normalized_prefix.len()
                            || entry
                                .path
                                .chars()
                                .nth(normalized_prefix.len())
                                .is_some_and(|c| c == '/' || c == std::path::MAIN_SEPARATOR))
                };

                if matches {
                    // Convert manifest entry to schema entry
                    entries.push(serde_json::json!({
                        "path": entry.path,
                        "type": "file", // Manifest only stores files?
                        "sha": entry.blob
                    }));
                }
            }

            Ok(serde_json::json!({
                "snapshot_id": sid,
                "path": path,
                "mode": "snapshot",
                "entries": entries,
                "truncated": false,
                "cache_key": sid,
                "cache_hint": "immutable"
            }))
        } else {
            Err(anyhow!("Invalid mode: {}", mode))
        }
    }

    pub fn read_file(
        &self,
        repo_root: &Path,
        path: &str,
        mode: &str,
        lease_id: Option<String>,
        snapshot_id: Option<String>,
    ) -> Result<Vec<u8>> {
        if mode == "worktree" {
            let lid = lease_id.ok_or_else(|| anyhow!("lease_id required for worktree mode"))?;
            self.lease_store.check_lease(&lid, repo_root)?;

            let abs_path = self.resolve_path(repo_root, path)?;
            if !abs_path.exists() {
                return Err(anyhow!("File not found: {}", path));
            }
            if abs_path.is_dir() {
                return Err(anyhow!("Path is a directory: {}", path));
            }

            let content = std::fs::read(&abs_path)?;

            // Touch file
            self.lease_store.touch_files(&lid, vec![path.to_string()]);

            Ok(content)
        } else if mode == "snapshot" {
            let sid =
                snapshot_id.ok_or_else(|| anyhow!("snapshot_id required for snapshot mode"))?;
            let manifest_json = self
                .blob_store
                .get(&sid)
                .ok_or_else(|| anyhow!("Snapshot not found"))?;
            let manifest: Manifest = serde_json::from_slice(&manifest_json)?;

            // Find entry
            // Manifest is sorted, we could binary search if explicit path.
            // But Vec linear scan is fine for now.
            let entry = manifest
                .entries
                .iter()
                .find(|e| e.path == path)
                .ok_or_else(|| anyhow!("File not found in snapshot: {}", path))?;

            // Get blob
            self.blob_store
                .get(&entry.blob)
                .ok_or_else(|| anyhow!("Blob missing for {}", path))
        } else {
            Err(anyhow!("Invalid mode: {}", mode))
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub fn grep_snapshot(
        &self,
        repo_root: &Path,
        pattern: &str,
        path: &str, // search scope
        mode: &str,
        lease_id: Option<String>,
        snapshot_id: Option<String>,
        case_insensitive: bool,
    ) -> Result<Vec<(String, usize, String)>> {
        // Returns (path, line_number, content)
        // Note: Rippgrep usually outputs robustly.
        // For worktree:
        if mode == "worktree" {
            let lid = lease_id.ok_or_else(|| anyhow!("lease_id required"))?;
            self.lease_store.check_lease(&lid, repo_root)?;

            // Run grep (git grep or ripgrep?)
            // Use `git grep -n -I` to avoid binary and getting line numbers.
            // pattern might be regex. path is scope.

            let mut cmd = std::process::Command::new("git");
            cmd.arg("grep").arg("-n").arg("-I");
            if case_insensitive {
                cmd.arg("-i");
            }
            cmd.arg(pattern);

            // If path is provided and not ".", add it
            if path != "." {
                cmd.arg("--").arg(path);
            }

            cmd.current_dir(repo_root);

            let output = cmd.output()?;
            if !output.status.success() {
                // exit code 1 means no matches usually.
                if output.status.code() == Some(1) {
                    return Ok(Vec::new());
                }
                return Err(anyhow!(
                    "git grep failed: {}",
                    String::from_utf8_lossy(&output.stderr)
                ));
            }

            let stdout = String::from_utf8_lossy(&output.stdout);
            let mut results = Vec::new();
            let mut touched = Vec::new();

            for line in stdout.lines() {
                // Format: path:line:content
                // Need to parse carefully.
                // Limit results? "Total results capped at 50" (from tool def usually, but internal limits?)
                // Let's parse all, and caller limits? Or limit here.
                // Spec doesn't strictly limit internal aggregation but tool output.

                let parts: Vec<&str> = line.splitn(3, ':').collect();
                if parts.len() == 3 {
                    let p = parts[0].to_string();
                    let result_path = p.clone();

                    // Helper strictness: verify path is under `path` arg if not handled by git?
                    // Git handles it.

                    let line_num: usize = parts[1].parse().unwrap_or(0);
                    let content = parts[2].to_string();

                    results.push((p, line_num, content));
                    touched.push(result_path);
                }
            }

            // Touch ALL candidate files?
            // "Touches all candidate files resolved under paths (deterministic lexicographic order). Binary files are excluded."
            // `git grep` output only lists MATCHING files.
            // We need to touch files that were SEARCHED?
            // That's expensive.
            // Spec says: "grep(worktree): Touches all candidate files resolved under paths...".
            // This means strict determinism requires invalidation if ANY file in the search scope changes, even if it didn't match before.
            // Because a change might MAKE it match.

            // So we must list candidates under `path` and touch them all!
            // `git ls-files path`?

            let ls_cmd = std::process::Command::new("git")
                .arg("ls-files")
                .arg(path)
                .current_dir(repo_root)
                .output()?;
            if ls_cmd.status.success() {
                let files_list = String::from_utf8_lossy(&ls_cmd.stdout);
                let all_candidates: Vec<String> =
                    files_list.lines().map(|s| s.to_string()).collect();
                self.lease_store.touch_files(&lid, all_candidates);
            }

            // Return filtered matches
            Ok(results)
        } else if mode == "snapshot" {
            let sid = snapshot_id.ok_or_else(|| anyhow!("snapshot_id required"))?;
            let manifest_json = self
                .blob_store
                .get(&sid)
                .ok_or_else(|| anyhow!("Snapshot not found"))?;
            let manifest: Manifest = serde_json::from_slice(&manifest_json)?;

            let mut results = Vec::new();
            let regex = regex::RegexBuilder::new(pattern)
                .case_insensitive(case_insensitive)
                .build()
                .context("Invalid regex")?;

            let normalized_prefix = if path == "." { "" } else { path };

            for entry in manifest.entries {
                // Filter by path
                let under_path = if normalized_prefix.is_empty() {
                    true
                } else {
                    entry.path.starts_with(normalized_prefix)
                };

                if !under_path {
                    continue;
                }

                // Get content
                if let Some(blob) = self.blob_store.get(&entry.blob) {
                    // check binary?
                    // simple check: if contains null byte within first 8000 bytes?
                    if blob.iter().take(8000).any(|&b| b == 0) {
                        continue;
                    }

                    if let Ok(text) = String::from_utf8(blob) {
                        for (i, line) in text.lines().enumerate() {
                            if regex.is_match(line) {
                                results.push((entry.path.clone(), i + 1, line.to_string()));
                            }
                        }
                    }
                }
            }
            Ok(results)
        } else {
            Err(anyhow!("Invalid mode"))
        }
    }
}
