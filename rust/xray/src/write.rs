use anyhow::{Context, Result};
use std::fs::{self, File};
use std::io::Write;
use std::path::Path;

/// Atomically writes content to a file.
/// 
/// 1. Writes to path.tmp
/// 2. Renames path.tmp -> path
pub fn write_atomic(path: &Path, content: &[u8]) -> Result<()> {
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent).context("Failed to create parent directory")?;
    }

    let temp_path = path.with_extension("tmp");
    
    // Write to temp
    {
        let mut file = File::create(&temp_path).context("Failed to create temp file")?;
        file.write_all(content).context("Failed to write content to temp file")?;
        file.sync_all().context("Failed to sync temp file")?;
    }

    // Rename
    fs::rename(&temp_path, path).context("Failed to rename temp file to target")?;

    Ok(())
}
