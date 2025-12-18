use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;
use std::sync::{Arc, RwLock};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Mount {
    pub name: String,
    pub root: String,
    pub resolved_id: Option<String>,
    pub kind: Option<String>,
    pub capabilities: Vec<String>,
}

#[derive(Clone, Default)]
pub struct MountRegistry {
    // RwLock for thread-safe concurrent access in the future,
    // though stdio server is currently sequential.
    mounts: Arc<RwLock<BTreeMap<String, Mount>>>,
}

impl MountRegistry {
    pub fn new() -> Self {
        Self {
            mounts: Arc::new(RwLock::new(BTreeMap::new())),
        }
    }

    pub fn register(&self, mount: Mount) {
        let mut map = self.mounts.write().unwrap();
        map.insert(mount.name.clone(), mount);
    }

    pub fn list(&self) -> Vec<Mount> {
        let map = self.mounts.read().unwrap();
        map.values().cloned().collect()
    }

    pub fn get(&self, name: &str) -> Option<Mount> {
        let map = self.mounts.read().unwrap();
        map.get(name).cloned()
    }
}
