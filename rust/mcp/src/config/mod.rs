// Config helpers
use serde::{Deserialize, Serialize};
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum BlobBackend {
    Fs,
    Db,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Compression {
    None,
    Zstd,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StorageConfig {
    pub data_dir: PathBuf,
    pub blob_backend: BlobBackend,
    pub compression: Compression,
}

impl Default for StorageConfig {
    fn default() -> Self {
        let data_dir = dirs::home_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join(".cortex/data");

        Self {
            data_dir,
            blob_backend: BlobBackend::Fs,
            compression: Compression::None, // Default to None for simplicity initially
        }
    }
}
