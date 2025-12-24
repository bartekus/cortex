use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

/// A content-addressed store for file blobs.
/// In a real implementation this might be backed by disk or DB, keeping it in-memory for now
/// as per the "work-with-app" scope, or maybe we just reference git blobs?
/// For the "snapshot" tool, creating a self-contained snapshot often implies stashing the content
/// aside from the worktree so it remains immutable even if worktree changes.
/// We'll use an in-memory map for the `v1` scope.
#[derive(Clone, Default)]
pub struct BlobStore {
    blobs: Arc<Mutex<HashMap<String, Vec<u8>>>>,
}

impl BlobStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn put(&self, content: &[u8]) -> String {
        let mut hasher = Sha256::new();
        hasher.update(content);
        let hash = format!("sha256:{}", hex::encode(hasher.finalize()));

        let mut blobs = self.blobs.lock().unwrap();
        blobs
            .entry(hash.clone())
            .or_insert_with(|| content.to_vec());
        hash
    }

    pub fn get(&self, hash: &str) -> Option<Vec<u8>> {
        let blobs = self.blobs.lock().unwrap();
        blobs.get(hash).cloned()
    }

    pub fn put_exact(&self, key: String, content: Vec<u8>) {
        let mut blobs = self.blobs.lock().unwrap();
        blobs.insert(key, content);
    }
}

/// A manifest representing a snapshot state.
/// Maps paths to blob hashes.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct Manifest {
    /// List of entries, MUST be sorted by path.
    pub entries: Vec<ManifestEntry>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct ManifestEntry {
    pub path: String,
    pub blob: String,
}

impl Manifest {
    pub fn new(entries: Vec<ManifestEntry>) -> Self {
        let mut sorted = entries;
        sorted.sort_by(|a, b| a.path.cmp(&b.path));
        Self { entries: sorted }
    }

    /// Computes the canonical JSON representation of the manifest.
    pub fn to_canonical_json(&self) -> Result<Vec<u8>> {
        // Use serde_json::to_vec for compact JSON.
        // BTreeMap ensure keys are sorted if we were using a map, but we use a Vec of structs.
        // Struct fields are serialized in definition order.
        // We need to ensure `entries` is sorted by path, which `new` enforces or we should enforce here.
        // But for "canonical manifest bytes" used in snapshot_id, we need a strictly defined format.
        // The spec say: { "entries": [ { "path": "...", "blob": "..." } ] }

        // We re-sort to be safe before serializing?
        // Or assume it's created via `new`.

        serde_json::to_vec(self).context("Failed to serialize manifest")
    }
}
