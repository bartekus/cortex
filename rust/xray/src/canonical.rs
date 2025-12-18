// SPDX-License-Identifier: AGPL-3.0-or-later

use crate::schema::XrayIndex;
use anyhow::{Context, Result};
use serde_json::{Map, Value};

/// Serializes the index to **Canonical JSON** (object keys sorted lexicographically, no extra whitespace).
///
/// Determinism requirements:
/// - All JSON objects MUST have keys sorted (lexicographically).
/// - Arrays MUST already be deterministically ordered by the caller/spec (e.g., files sorted by path).
/// - Output MUST be compact (no pretty-print / no whitespace variance).
///
/// Notes:
/// - `serde_json` will emit struct fields in struct declaration order, and map keys in map iteration order.
/// - Using `BTreeMap` helps, but does not guarantee recursive key ordering for *all* nested objects.
/// - Therefore we canonicalize by converting to `serde_json::Value` and recursively sorting object keys.
pub fn to_canonical_json(index: &XrayIndex) -> Result<Vec<u8>> {
    // Enforce invariants before serialization
    validate_invariants(index)?;

    let value = serde_json::to_value(index).context("Failed to convert index to JSON value")?;
    let canon = canonicalize_value(value);
    serde_json::to_vec(&canon).context("Failed to serialize canonical JSON")
}

fn canonicalize_value(v: Value) -> Value {
    match v {
        Value::Object(map) => canonicalize_object(map),
        Value::Array(arr) => Value::Array(arr.into_iter().map(canonicalize_value).collect()),
        other => other,
    }
}

fn canonicalize_object(map: Map<String, Value>) -> Value {
    // Collect into a Vec to handle sorting without re-lookup
    let mut entries: Vec<(String, Value)> = map.into_iter().collect();

    // Sort keys lexicographically.
    entries.sort_by(|a, b| a.0.cmp(&b.0));

    let mut out = Map::new();
    for (k, v) in entries {
        out.insert(k, canonicalize_value(v));
    }

    Value::Object(out)
}

/// Validates that the index is sorted correctly and adheres to invariants.
pub fn validate_invariants(index: &XrayIndex) -> Result<()> {
    // 1. Check files are strictly sorted by path
    for (i, window) in index.files.windows(2).enumerate() {
        if window[0].path >= window[1].path {
            if window[0].path == window[1].path {
                anyhow::bail!("Duplicate file path at index {}: {}", i, window[0].path);
            }
            anyhow::bail!(
                "Files not sorted at index {}: {} >= {}",
                i,
                window[0].path,
                window[1].path
            );
        }
    }

    // 2. Check module_files are strictly sorted
    for (i, window) in index.module_files.windows(2).enumerate() {
        if window[0] >= window[1] {
            if window[0] == window[1] {
                anyhow::bail!("Duplicate module file at index {}: {}", i, window[0]);
            }
            anyhow::bail!(
                "Module files not sorted at index {}: {} >= {}",
                i,
                window[0],
                window[1]
            );
        }
    }

    // 3. Stats consistency
    if index.files.len() != index.stats.file_count {
        anyhow::bail!(
            "File count mismatch: files.len()={} vs stats.file_count={}",
            index.files.len(),
            index.stats.file_count
        );
    }

    let computed_size: u64 = index.files.iter().map(|f| f.size).sum();
    if computed_size != index.stats.total_size {
        anyhow::bail!(
            "Total size mismatch: computed={} vs stats.total_size={}",
            computed_size,
            index.stats.total_size
        );
    }

    Ok(())
}
