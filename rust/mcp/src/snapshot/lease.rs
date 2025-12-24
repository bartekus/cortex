use anyhow::{anyhow, Context, Result};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::collections::{BTreeSet, HashMap};
use std::path::Path;
use std::process::Command;
use std::sync::{Arc, Mutex};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct Fingerprint {
    pub head_oid: String,
    pub index_oid: String,
    pub status_hash: String,
}

impl Fingerprint {
    pub fn compute(repo_root: &Path) -> Result<Self> {
        // 1. Get HEAD OID
        let head_output = Command::new("git")
            .arg("rev-parse")
            .arg("HEAD")
            .current_dir(repo_root)
            .output()?;

        let head_oid = if head_output.status.success() {
            String::from_utf8_lossy(&head_output.stdout)
                .trim()
                .to_string()
        } else {
            // Assume unborn if failure or handle explicitly?
            // Spec says "Empty string if unborn". 'git rev-parse HEAD' usually fails on unborn.
            String::new()
        };

        // 2. Get Index OID
        let index_output = Command::new("git")
            .arg("write-tree")
            .current_dir(repo_root)
            .output()?;

        let index_oid = if index_output.status.success() {
            String::from_utf8_lossy(&index_output.stdout)
                .trim()
                .to_string()
        } else {
            // Spec says: "If no tree possible (defined condition) -> empty string. Other failures -> INTERNAL error."
            // Git write-tree fails if there are merge conflicts in index or other malformed states.
            // We'll treat failures as error for now, unless we can detect "no tree possible" reliably.
            // On strict error mode, maybe we propagate error?
            // But for now, let's bubble up the error if it fails unexpectedly.
            // Except, if index is empty (new repo), write-tree still produces empty tree hash.
            // So real failure is error.
            if !index_output.stderr.is_empty() {
                return Err(anyhow!(
                    "git write-tree failed: {}",
                    String::from_utf8_lossy(&index_output.stderr)
                ));
            }
            String::new()
        };

        // 3. Get Status Hash
        // git status --porcelain=v1 -z
        let status_output = Command::new("git")
            .arg("status")
            .arg("--porcelain=v1")
            .arg("-z")
            .current_dir(repo_root)
            .output()
            .context("Failed to run git status")?;

        if !status_output.status.success() {
            return Err(anyhow!("git status failed"));
        }

        let mut hasher = Sha256::new();
        hasher.update(&status_output.stdout);
        let status_hash = format!("sha256:{}", hex::encode(hasher.finalize()));

        Ok(Self {
            head_oid,
            index_oid,
            status_hash,
        })
    }
}

pub struct Lease {
    pub id: String,
    pub base_fingerprint: Fingerprint,
    pub touched_files: BTreeSet<String>,
}

#[derive(Clone, Default)]
pub struct LeaseStore {
    leases: Arc<Mutex<HashMap<String, Lease>>>,
}

impl LeaseStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn issue_lease(&self, repo_root: &Path) -> Result<(String, Fingerprint)> {
        let fingerprint = Fingerprint::compute(repo_root)?;
        let id = Uuid::new_v4().to_string();

        let lease = Lease {
            id: id.clone(),
            base_fingerprint: fingerprint.clone(),
            touched_files: BTreeSet::new(),
        };

        let mut leases = self.leases.lock().unwrap();
        leases.insert(id.clone(), lease);
        Ok((id, fingerprint))
    }

    pub fn get_lease(&self, _lease_id: &str) -> Option<Lease> {
        // Return a clone? Or separate access method?
        // For now, simple retrieval. Ideally we want to verify and update touched files.
        // Wait, we need interior mutability for touched_files if we operate on the lease struct?
        // We should expose methods to check and touch.
        let _leases = self.leases.lock().unwrap();
        // We need to clone expensive structures? Fingerprint is small. Touched files can be large.
        // Maybe return a read guard equivalent or just clone for now (Simplicity).
        // Actually, Lease struct above owns data.
        // We should just use methods on LeaseStore to operate.
        // But for `snapshot.create` we need to read touched files.
        // For `worktree` tools we need to update touched files.
        None // Placeholder behavior if we change design, but let's implement get for now.
    }

    // Better API:

    pub fn check_lease(&self, lease_id: &str, repo_root: &Path) -> Result<Fingerprint> {
        let current_fingerprint = Fingerprint::compute(repo_root)?;

        let mut leases = self.leases.lock().unwrap();
        if let Some(lease) = leases.get_mut(lease_id) {
            if lease.base_fingerprint != current_fingerprint {
                // Return Stale Lease Error
                // BUT we can't return structured error easily here without `thiserror` mapping.
                // We will return the current fingerprint as Ok, but caller must compare?
                // No, caller needs to error.
                // We can return Result<Fingerprint, (StaleError, Fingerprint)>?
                // Let's return Ok(current) and let caller compare?
                // Or better: ensure caller handles the mismatch logic.
                // The spec says: "Return STALE_LEASE error containing current fingerprint".
                // So we should fail here if mismatch.
                return Err(
                    anyhow!("STALE_LEASE").context(serde_json::to_string(&current_fingerprint)?)
                );
            }
            Ok(current_fingerprint)
        } else {
            Err(anyhow!("Lease not found"))
        }
    }

    pub fn touch_files(&self, lease_id: &str, files: impl IntoIterator<Item = String>) {
        let mut leases = self.leases.lock().unwrap();
        if let Some(lease) = leases.get_mut(lease_id) {
            lease.touched_files.extend(files);
        }
    }

    pub fn get_touched_files(&self, lease_id: &str) -> Option<Vec<String>> {
        let leases = self.leases.lock().unwrap();
        leases
            .get(lease_id)
            .map(|l| l.touched_files.iter().cloned().collect())
    }
}
