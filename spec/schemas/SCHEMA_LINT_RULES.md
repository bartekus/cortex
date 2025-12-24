# Schema Lint Rules

The following rules are enforced by `tests/harness/schema_lint.ts` to ensure the correctness of the MCP tool schemas, specifically regarding hybrid coherence and cache safety.

## 1. Branching Correctness

### Hybrid Tools
Tools that support both `worktree` and `snapshot` modes must have a success response schema defined as a `oneOf` with exactly two branches:
1.  **Immutable Branch**: Coincides with `snapshot` mode.
2.  **Worktree Branch**: Coincides with `worktree` mode.

**Applicable Tools**:
- `snapshot.list`
- `snapshot.file`
- `snapshot.grep`
- `snapshot.diff`
- `workspace.apply_patch` (returns snapshot_id or lease_id)

### Snapshot-Only Tools
Tools that produce strictly immutable outputs (or are pure functions of input) must have a single success branch.

**Applicable Tools**:
- `snapshot.create`
- `snapshot.info`
- `snapshot.export`
- `snapshot.changes`

## 2. Cache Hint Correctness

Every success branch **MUST** include a `cache_hint` property that is a `const` string.

- **Immutable Branch**: `cache_hint` MUST be `"immutable"`.
- **Worktree Branch**: `cache_hint` MUST be `"until_dirty"`.

## 3. Worktree Requirements

The **Worktree Branch** of any hybrid tool response schema **MUST** include:
- `lease_id` (string)
- `fingerprint` (object ref)

## 4. Error Enums

The common error schema (`common.schema.json`) **MUST** include `STALE_LEASE` in the `code` enum.

## 5. Runtime Enforcement

The implementation **MUST** verify at runtime that the returned `cache_hint` matches the schema expectation for the active branch/mode.
