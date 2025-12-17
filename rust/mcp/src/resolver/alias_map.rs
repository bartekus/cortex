use crate::io::Fs;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct AliasMap {
    // Deterministic order via BTreeMap.
    pub aliases: BTreeMap<String, String>,
}

pub fn default_aliases_path() -> Option<PathBuf> {
    dirs::home_dir().map(|h| h.join(".cortex").join("mcp-aliases.json"))
}

pub fn load_alias_map<F: Fs>(fs: &F, path: &Path) -> Result<AliasMap> {
    if !fs.exists(path) {
        return Ok(AliasMap::default());
    }
    let raw = fs.read_to_string(path)?;
    let parsed: AliasMap = serde_json::from_str(&raw)?;
    Ok(parsed)
}

pub fn resolve_alias(map: &AliasMap, name: &str) -> Option<PathBuf> {
    map.aliases.get(name).map(PathBuf::from)
}
