use anyhow::Result;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::collections::HashMap;
use std::sync::{Arc, RwLock};

#[derive(Clone, Default)]
pub struct BlobStore {
    blobs: Arc<RwLock<HashMap<String, Vec<u8>>>>,
    snapshots: Arc<RwLock<HashMap<String, Vec<u8>>>>, // Map snapshot_id -> manifest bytes match
}

impl BlobStore {
    pub fn new() -> Self {
        Self {
            blobs: Arc::new(RwLock::new(HashMap::new())),
            snapshots: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn put(&self, data: &[u8]) -> String {
        let hash = format!("sha256:{}", hex::encode(Sha256::digest(data)));
        let mut blobs = self.blobs.write().unwrap();
        blobs.insert(hash.clone(), data.to_vec());
        hash
    }

    pub fn get(&self, hash: &str) -> Option<Vec<u8>> {
        let blobs = self.blobs.read().unwrap();
        blobs.get(hash).cloned()
    }

    pub fn put_snapshot(&self, id: String, manifest_bytes: Vec<u8>) {
        let mut snaps = self.snapshots.write().unwrap();
        snaps.insert(id, manifest_bytes);
    }

    pub fn get_snapshot(&self, id: &str) -> Option<Vec<u8>> {
        let snaps = self.snapshots.read().unwrap();
        snaps.get(id).cloned()
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct Entry {
    pub blob: String,
    pub path: String,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Manifest {
    pub entries: Vec<Entry>,
}

impl Manifest {
    pub fn new(mut entries: Vec<Entry>) -> Self {
        // Enforce deterministic order (lexicographic by path)
        entries.sort_by(|a, b| a.path.cmp(&b.path));
        Self { entries }
    }

    /// Serializes to Canonical JSON
    pub fn to_canonical_json(&self) -> Result<String> {
        // serde_json::to_string ensures strict JSON output (no whitespace by default)
        // Order of keys in objects: serde_json default (alphabetical for BTreeMap-like but struct fields order?)
        // Struct fields are serialized in definition order usually?
        // Canonical JSON requires sorted keys.
        // Option A: Use a BTreeMap to intermediate.
        // Option B: Assume `entries` (array) is ordered (we did that).
        // `Entry` has `path`, `blob`. P comes after B. So `blob` then `path`.
        // If we want lexicographic keys: "blob", "path".
        // `Entries` key in manifest: "entries".
        // So we need to ensure struct fields are serialized in key order.
        // It's safer to convert to serde_json::Value and let it sort?
        // serde_json (since v1.0) preserves insertion order of maps by default if "preserve_order" is on?
        // No, standard `to_string` sorts map keys.
        // But structs?
        // Let's use `serde_json::to_value` then `to_string`. Value uses BTreeMap for objects usually (if feature "preserve_order" not enabled).

        let val = serde_json::to_value(self)?;
        // This ensures keys are sorted
        let s = serde_json::to_string(&val)?;
        Ok(s)
    }

    pub fn compute_snapshot_id(&self, fingerprint_json: &str) -> Result<String> {
        let manifest_json = self.to_canonical_json()?;
        let raw = format!("{}\n{}", fingerprint_json, manifest_json);
        let hash = hex::encode(Sha256::digest(raw.as_bytes()));
        Ok(format!("sha256:{}", hash))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_manifest_sorting() {
        let entries = vec![
            Entry {
                path: "src/b.ts".into(),
                blob: "sha256:2".into(),
            },
            Entry {
                path: "src/a.ts".into(),
                blob: "sha256:1".into(),
            },
        ];
        let manifest = Manifest::new(entries);
        assert_eq!(manifest.entries[0].path, "src/a.ts");
        assert_eq!(manifest.entries[1].path, "src/b.ts");
    }

    #[test]
    fn test_canonical_json() {
        let entries = vec![Entry {
            path: "a".into(),
            blob: "b".into(),
        }];
        let manifest = Manifest::new(entries);
        let json = manifest.to_canonical_json().unwrap();
        // check keys sorted. "entries" is only key here.
        // inside entry: blob, then path.
        // {"entries":[{"blob":"b","path":"a"}]}
        assert_eq!(json, r#"{"entries":[{"blob":"b","path":"a"}]}"#);
    }
}
