use crate::schema::XrayIndex;
use crate::canonical::to_canonical_json;
use sha2::{Digest, Sha256};
use anyhow::Result;

/// Calculates the repository digest.
///
/// The digest is: SHA-256( CanonicalJSON( Index( digest="" ) ) )
/// 
/// 1. Clone the index.
/// 2. Set digest to empty string.
/// 3. Serialize to canonical JSON.
/// 4. Hash it.
pub fn calculate_digest(index: &XrayIndex) -> Result<String> {
    let mut clone = XrayIndex {
        schema_version: index.schema_version.clone(),
        root: index.root.clone(),
        target: index.target.clone(),
        files: index.files.clone(),
        languages: index.languages.clone(), // BTreeMaps are already sorted
        top_dirs: index.top_dirs.clone(),
        module_files: index.module_files.clone(),
        stats: index.stats.clone(),
        digest: "".to_string(), // MUST be empty for calculation
    };

    // Ensure strict sorting before hashing
    clone.files.sort_by(|a, b| a.path.cmp(&b.path));
    clone.module_files.sort();

    let bytes = to_canonical_json(&clone)?;
    let mut hasher = Sha256::new();
    hasher.update(&bytes);
    let result = hasher.finalize();

    Ok(hex::encode(result))
}
