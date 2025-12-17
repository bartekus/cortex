use std::path::{Path, Component};

pub fn normalize_path(path: &Path) -> String {
    let mut s = path.to_string_lossy().replace("\\", "/");
    if s.len() > 1 && s.ends_with('/') {
        s.pop();
    }
    s
}

pub fn path_depth(path: &Path) -> usize {
    path.components().filter(|c| matches!(c, Component::Normal(_))).count()
}
