use super::Fs;
use anyhow::{anyhow, Result};
use std::collections::{HashMap, HashSet};
use std::path::{Path, PathBuf};
use std::sync::Mutex;

#[derive(Debug, Default)]
pub struct MemFs {
    pub files: Mutex<HashMap<PathBuf, String>>,
    pub dirs: Mutex<HashSet<PathBuf>>,
}

impl MemFs {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn add_file(&self, path: impl Into<PathBuf>, content: impl Into<String>) {
        let path = path.into();
        self.ensure_parent(&path);
        self.files.lock().unwrap().insert(path, content.into());
    }
    
    pub fn add_dir(&self, path: impl Into<PathBuf>) {
        let path = path.into();
        self.ensure_parent(&path);
        self.dirs.lock().unwrap().insert(path);
    }

    fn ensure_parent(&self, path: &Path) {
        if let Some(parent) = path.parent() {
            if parent != Path::new("") && parent != Path::new("/") {
                let mut dirs = self.dirs.lock().unwrap();
                if !dirs.contains(parent) {
                    dirs.insert(parent.to_path_buf());
                    drop(dirs); // drop lock before recursing
                    self.ensure_parent(parent);
                }
            }
        }
    }
}

impl Fs for MemFs {
    fn read_to_string(&self, path: &Path) -> Result<String> {
        let files = self.files.lock().unwrap();
        files.get(path).cloned().ok_or_else(|| anyhow!("File not found: {:?}", path))
    }

    fn read_dir(&self, path: &Path) -> Result<Vec<PathBuf>> {
        let files = self.files.lock().unwrap();
        let dirs = self.dirs.lock().unwrap();
        
        let mut entries = Vec::new();
        
        // Find direct children in files
        for file_path in files.keys() {
            if let Some(parent) = file_path.parent() {
                if parent == path {
                    entries.push(file_path.clone());
                }
            }
        }
        
        // Find direct children in dirs
        for dir_path in dirs.iter() {
            if let Some(parent) = dir_path.parent() {
                if parent == path && dir_path != path {
                    entries.push(dir_path.clone());
                }
            }
        }
        
        entries.sort();
        entries.dedup();
        
        Ok(entries)
    }

    fn exists(&self, path: &Path) -> bool {
        self.files.lock().unwrap().contains_key(path) || self.dirs.lock().unwrap().contains(path)
    }

    fn is_dir(&self, path: &Path) -> bool {
        self.dirs.lock().unwrap().contains(path)
    }

    fn canonicalize(&self, path: &Path) -> Result<PathBuf> {
        // MemFs simulates absolute paths as-is
        Ok(path.to_path_buf())
    }
}
