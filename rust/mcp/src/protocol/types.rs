use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResolveRequest {
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResolveResponse {
    pub status: ResolveStatus,
    pub resolved_id: Option<String>,
    pub kind: Option<String>,
    pub root: Option<String>,
    pub capabilities: Vec<String>,
    pub tried: Vec<String>,
    pub fix_hint: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum ResolveStatus {
    Resolved,
    Unresolved,
}
