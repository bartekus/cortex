# Schema Lint Rules

These rules are enforced by `tests/harness/schema_lint.ts` to ensure strict determinism and valid caching headers across all MCP tool responses.

## 1. Response Structure
Every response schema **MUST** be a `oneOf` object with exactly two top-level branches (unless it's a hybrid tool, see below):
1.  **Success Branch**: An object definition for successful execution.
2.  **Error Branch**: A `$ref` to `common.schema.json#/$defs/error`.

## 2. Success Branch Requirements
Every success branch object **MUST**:
- Set `additionalProperties: false`.
- Include `cache_key` (string).
- Include `cache_hint` (const string).

## 3. Cache Hint Contracts
The value of `cache_hint` depends on the tool mode and branch:

### Snapshot-Only Tools
Tools that only operate on immutable snapshots (e.g., `snapshot.create`, `snapshot.info`).
- **Structure**: `oneOf` [Success, Error]
- **Success Branch**: `cache_hint: "immutable"`

### Hybrid Tools
Tools that support both `worktree` and `snapshot` modes (e.g., `snapshot.file`, `snapshot.list`, `snapshot.grep`, `snapshot.diff`, `snapshot.export`).
- **Structure**: `oneOf` [WorktreeSuccess, SnapshotSuccess, Error]
- **SnapshotSuccess Branch**:
    - `cache_hint: "immutable"`
- **WorktreeSuccess Branch**:
    - `cache_hint: "until_dirty"`
    - **MUST** include `lease_id`.
    - **MUST** include `fingerprint`.

## 4. Defs and Refs
- All `$ref` usage must be local (relative paths within `spec/schemas`).
- No external HTTP refs.
- No `..` parent references outside the schemas directory.
