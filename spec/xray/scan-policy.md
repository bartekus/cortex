---
feature: XRAY_SCAN_POLICY
version: v1
status: approved
domain: xray
inputs:
  flags: []
  args: []
outputs:
  artifacts:
    - .cortex/data/index.json
---
# XRAY Scan Policy

**Feature**: `XRAY_SCAN_POLICY`
**Status**: Approved

## Purpose
Defines the logic governing *how* a repository is scanned, including what is ignored, how languages are detected, and how determinism is enforced.

## Scan Scope
### Inclusion
- By default, scans all files in the target directory recursively.

### Exclusion (Ignored)
- **Dot-directories**: `.git`, `.github`, `.cortex`, etc. are ignored by default unless explicitly targeted.
- **Binary/Large Files**: Files exceeding size thresholds (default 2MB) or invalid UTF-8 are **Included** in the index but marked as skipped with 0 LOC.

## Language Detection
- **Method**: Extension-based detection (primary).
- **Unknowns**: Files with unrecognized extensions map to "Unknown" language.
- **Aggregation**: Files with "Unknown" language are **Excluded** from the global `languages` summary map (but remain in `files` list).

## Determinism Guarantees
1.  **LOC Counting**: Logical lines are counted via `str::lines().count()`. This is distinct from POSIX `wc -l` (which requires trailing newline).
2.  **Canonical Output**: The JSON output generation MUST use a stable, lexicographically sorted key order for all objects.
3.  **Stable Hash**: File digests are stable SHA256 of content.

## Failure Model
- **Permission Denied**: Logs error/warning but continues scan (soft fail).
- **Symlink Cycles**: Detected and broken to prevent infinite loops.
