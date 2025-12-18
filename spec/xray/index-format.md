# XRAY Index Format

**Feature**: `XRAY_INDEX_FORMAT`
**Status**: Approved

## Purpose
Defines the JSON schema and structural invariants of the XRAY index file (`index.json`). This file is the primary artifact of a repository scan.

## Schema: index.json
The root object MUST contain the following key fields:

| Field | Type | Description |
| :--- | :--- | :--- |
| `root` | String | The slug/name of the repository root. |
| `target` | String | Relative path scanned (e.g. `"."`). |
| `digest` | String | SHA256 digest of the *content* of the index (integrity check). |
| `files` | List[FileNode] | Flat list of all scanned files. |

### FileNode Object
| Field | Type | Description |
| :--- | :--- | :--- |
| `path` | String | Relative path from repo root. **MUST be sorted** within the `files` list. |
| `loc` | Integer | Logical Lines of Code (content lines). |
| `size` | Integer | Size in bytes. |
| `language` | String | Detected language (or "Unknown"). |
| `digest` | String | SHA256 content hash of the file. |

## Invariants
1.  **Sorting**: The `files` array MUST be strictly sorted by `path` (lexicographically).
2.  **Uniqueness**: No duplicate paths allowed.
3.  **Digest Integrity**: The root `digest` is calculated over the **Canonical JSON** representation of the index *excluding* the digest field itself.
    - **Keys**: All object keys MUST be sorted lexicographically (recursive).
    - **Whitespace**: No extra whitespace (minified).
    - **Encoding**: UTF-8.
    (Validated against canonical fixtures; automated validation is required).

## Example: Valid Index
```json
{
  "root": "cortex",
  "target": ".",
  "files": [
    {
      "path": "Makefile",
      "loc": 86,
      "size": 2071,
      "language": "Makefile",
      "digest": "sha256:..."
    },
    {
      "path": "cmd/cortex/main.go",
      "loc": 64,
      "size": 1735,
      "language": "Go",
      "digest": "sha256:..."
    }
  ],
  "stats": { ... },
  "digest": "a1b2c3d4..."
}
```
