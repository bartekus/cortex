# MCP Snapshot & Workspace Substrate (v1)

This specification defines the deterministic `snapshot` and `workspace` tools for MCP. It enforces strict determinism, hybrid coherence (leases vs. immutable snapshots), and safe write operations.

## 1. Global Conventions

### 1.1 Encoding
- All requests and responses are UTF-8 JSON.
- Binary data **MUST** be encoded as `base64:` prefixed strings.
- Textual diffs **MUST** use LF (`\n`) line endings.
- **Canonical JSON Algorithm**:
    - UTF-8 encoding.
    - Object keys sorted lexicographically (byte-order).
    - No insignificant whitespace (compact).
    - Standard JSON string escapes only (no `\u` escapes for basic ASCII).
    - Arrays preserved in given order (entries must be pre-sorted by the tool logic).
    - No trailing newline (except explicit separators defined below).

### 1.2 Paths
- All paths are **repo-relative**, POSIX style (`/` separators).
- Paths **MUST NOT** contain `..`, `~`, or absolute roots.
- All returned paths **MUST** be normalized and lexicographically sorted.
- Directory normalization: `src` and `src/` are identical. `.` represents the root.

### 1.3 Ordering Rules (Determinism)
- Arrays of paths are sorted lexicographically (byte-order).
- Maps with path keys are serialized in sorted key order.
- Matches within a file are ordered by `(line, col)`.
- Rejects in patch application are sorted by `path` then `hunk_index`.

## 2. Coherence Models

### 2.1 Hybrid Contract
The system operates in two distinct modes:
1.  **Worktree Mode (`worktree`)**: Live view of the filesystem. Coherence is managed via **leases**.
2.  **Snapshot Mode (`snapshot`)**: Immutable view of a captured state. Coherence is managed via **snapshot IDs** (content-addressed).

### 2.2 Fingerprint
A fingerprint uniquely identifies the state of the repo's HEAD, Index, and Working Tree.
- **Object Structure**:
  ```json
  {
    "head_oid": "...",       // SHA1 (hex). Empty string if unborn.
    "index_oid": "...",      // SHA1 (hex) from `git write-tree`. Empty if no tree possible.
    "status_hash": "..."     // SHA256 (hex) of `git status --porcelain=v1 -z` raw bytes.
  }
  ```
- **Serialization**: The fingerprint object is serialized using the Canonical JSON Algorithm.

### 2.3 Lease Semantics
- **Issuance**: A lease is issued by any `worktree`-mode read tool (`file`, `list`, `grep`, `diff`, `export`) or `workspace.apply_patch(mode=worktree)` if a valid `lease_id` is not provided.
- **Validation**: Every `worktree`-mode request with a `lease_id` validates it against the current live fingerprint.
- **Stale Lease**: If the current fingerprint differs from the lease's base fingerprint, the server returns a `STALE_LEASE` error containing the current fingerprint.
    - **No auto-refresh**. The client must retry.
    - Error must be structured (code "STALE_LEASE", details including fingerprint).
- **Touched Files**: The lease tracks files "touched" (read/listed) to support partial snapshot creation later.
    - `list`: Touches returned file entries + implicit parents.
    - `grep`: Touches **all candidate files** resolved under paths (deterministic order). Binary files are excluded from candidates (and thus not touched).
    - `diff`: Touches target path.
    - Mutators (`apply_patch`, `write_file`, `delete`): Touch affected paths.

### 2.4 Snapshot ID
- **Manifest**: A canonical JSON object mapping paths to blob hashes.
  ```json
  {
    "entries": [
      { "path": "src/a.ts", "blob": "sha256:..." },
      { "path": "src/b.ts", "blob": "sha256:..." }
    ]
  }
  ```
  Entries must be sorted lexicographically by path.

- **Derivation**:
  ```
  snapshot_id = sha256(
    canonical_fingerprint_json
    + "\n"
    + manifest_bytes
  )
  ```
- **Encoding**: `sha256:<hex>`.

## 3. Tool Specifications

### 3.1 Snapshot Tools

#### `snapshot.create`
- **Inputs**: `lease_id` (optional), `paths` (optional).
- **Behavior**: Creates an immutable snapshot manifest.
    - If `paths` are provided, captures those specific paths.
    - If `paths` omitted, captures all files **touched** by the lease.
- **Output**: `snapshot_id`.

#### `snapshot.list`
- **Mode `worktree`**: Lists live files, updates lease touched set.
- **Mode `snapshot`**: Lists files from the manifest.
    - **Strictness**: Returns only captured entries.
    - **Implicit Parents**: If `src/a.ts` is in manifest, `src` is listable.
    - **Unknown/Uncaptured**: Returns empty list (not error), `truncated=false`.

#### `snapshot.grep`
- **Candidates**: Deterministic walk (lexicographic). Ignore rules applied. Binary files excluded (frozen choice).
- **Limits**: Touches/searches candidates up to `max_files` limit. Returns `truncated=true` if limit hit.
- **Determinism**: Candidate selection order must be stable.

#### `snapshot.info`
- **Output**: `fingerprint` (object) + `manifest_stats` (files count, total bytes).
- **Note**: Does NOT return lease context.

#### `snapshot.changes`
- **Output**: List of changed files (status).
- **Determinism**: rename detection only if deterministic and spec'd.

#### `snapshot.export`
- **Output**: Deterministic bundle export format (order defined and stable).

### 3.2 Workspace Tools

#### `workspace.apply_patch`
- **Mode `worktree`**: Applies to live FS. Validates lease. Returns new `fingerprint` + `lease_id`.
- **Mode `snapshot`**: Applies to in-memory manifest (updates BlobStore). Returns new `snapshot_id`.
- **Format**: Unified Diff.
- **Policy**:
    - **Context Matching**: Byte-for-byte. No whitespace normalization.
    - **No Fuzzing**: Must match exactly.
    - **Rejects**: Structured list of `{ "path": "...", "hunks": [{ "index": 0, "reason": "context_mismatch" }] }`.
    - **Sorting**: Rejects sorted by `path` then `hunk_index`.

#### `workspace.write_file` / `workspace.delete`
- **Safety**:
    1.  `canonicalize(repo_root)`.
    2.  Resolve target: `canonicalize(target)` (or `parent` then join `filename` for new files).
    3.  **Reject** `PERMISSION_DENIED` if resolved path is not within `repo_root` prefix.
    4.  **Reject** `..` traversal.
    5.  **Reject** absolute paths outside repo.
    6.  **Reject** symlink escape out of repo.

## 4. Error Model

All errors MUST conform to a structured format, especially `STALE_LEASE`.

```json
{
  "error": {
    "code": "NOT_FOUND | INVALID_ARGUMENT | REPO_CHANGED | PERMISSION_DENIED | TOO_LARGE | INTERNAL | STALE_LEASE",
    "message": "human readable",
    "details": { "fingerprint": { ... } } // For STALE_LEASE, mandatory
  }
}
```

## 5. Schema Validation Rules
- **Snapshot-only tools**: Success branch (immutable) vs Error branch.
- **Hybrid tools**:
    - Success branch is `oneOf` [`ImmutableResponse`, `WorktreeResponse`].
    - `ImmutableResponse`: `cache_hint: "immutable"`.
    - `WorktreeResponse`: `cache_hint: "until_dirty"`, **MUST** include `lease_id` and `fingerprint`.
- **Cache Hints**: Must be `const` in schema and verified at runtime.
