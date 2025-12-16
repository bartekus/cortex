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
    // Sort keys lexicographically.
    let mut keys: Vec<String> = map.keys().cloned().collect();
    keys.sort();

    let mut out = Map::new();
    for k in keys {
        // Safe: key exists in original map.
        let child = map.get(&k).expect("key must exist").clone();
        out.insert(k, canonicalize_value(child));
    }

    Value::Object(out)
}

/// Validates that the index is sorted correctly.
/// Returns true if compliant, false if not.
pub fn validate_sort_order(index: &XrayIndex) -> bool {
    // Check files are strictly sorted by path
    for window in index.files.windows(2) {
        if window[0].path >= window[1].path {
            return false;
        }
    }

    // Check module_files are strictly sorted
    for window in index.module_files.windows(2) {
        if window[0] >= window[1] {
            return false;
        }
    }

    true
}
