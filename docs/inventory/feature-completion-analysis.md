# Feature Completion Analysis

> **Source**: Generated from `spec/features.yaml` by `cortex status roadmap`
> **Last Updated**: See `spec/features.yaml` for the source of truth
>
> **Note**: This document is automatically generated. To regenerate, run `cortex status roadmap`.

⸻

## Executive Summary

- **Total Features**: 17
- **Completed**: 0 (0.0%)
- **In Progress**: 0 (0.0%)
- **Planned**: 17 (100.0%)

⸻

## Phase-by-Phase Completion

| Phase | Features | Done | WIP | Todo | Completion | Status |
|-------|----------|------|-----|------|------------|--------|
| **--- CLI Commands ---** | 7 | 0 | 0 | 7 | 0% | ⚠️ Not started |
| **--- CLI Interface ---** | 1 | 0 | 0 | 1 | 0% | ⚠️ Not started |
| **--- Core System ---** | 1 | 0 | 0 | 1 | 0% | ⚠️ Not started |
| **--- MCP Protocol ---** | 2 | 0 | 0 | 2 | 0% | ⚠️ Not started |
| **--- Release & Distribution ---** | 1 | 0 | 0 | 1 | 0% | ⚠️ Not started |
| **--- Skills & Reports ---** | 2 | 0 | 0 | 2 | 0% | ⚠️ Not started |
| **--- XRAY Engine ---** | 3 | 0 | 0 | 3 | 0% | ⚠️ Not started |

⸻

## Roadmap Alignment

### Strong Progress


### Critical Gaps

- ⚠️ **--- CLI Commands ---**: 0% complete — not started
- ⚠️ **--- CLI Interface ---**: 0% complete — not started
- ⚠️ **--- Core System ---**: 0% complete — not started
- ⚠️ **--- MCP Protocol ---**: 0% complete — not started
- ⚠️ **--- Release & Distribution ---**: 0% complete — not started
- ⚠️ **--- Skills & Reports ---**: 0% complete — not started
- ⚠️ **--- XRAY Engine ---**: 0% complete — not started

⸻

## Priority Recommendations

### 🔥 Immediate (Unblocks Other Work)

1. Complete `CLI_COMMAND_COMMIT` to unblock dependent features
1. Complete `CLI_COMMAND_CONTEXT` to unblock dependent features
1. Complete `CLI_COMMAND_FEATURE` to unblock dependent features
1. Complete `CLI_COMMAND_FEATURES` to unblock dependent features
1. Complete `CLI_COMMAND_GOV` to unblock dependent features
1. Complete `CLI_COMMAND_RUN` to unblock dependent features
1. Complete `CLI_COMMAND_STATUS` to unblock dependent features
1. Complete `CLI_CONTRACT` to unblock dependent features
1. Complete `MCP_ROUTER_CONTRACT` to unblock dependent features
1. Complete `MCP_TOOLS` to unblock dependent features
1. Complete `REL_ARTIFACT_LAYOUT` to unblock dependent features
1. Complete `REPORTS_CORE` to unblock dependent features
1. Complete `SKILLS_REGISTRY` to unblock dependent features
1. Complete `XRAY_CLI` to unblock dependent features
1. Complete `XRAY_INDEX_FORMAT` to unblock dependent features
1. Complete `XRAY_SCAN_POLICY` to unblock dependent features

## Detailed Phase Analysis

### --- CLI Commands ---

- Features: 7 (Done: 0, WIP: 0, Todo: 7)
- Completion: 0.0%

### --- CLI Interface ---

- Features: 1 (Done: 0, WIP: 0, Todo: 1)
- Completion: 0.0%

### --- Core System ---

- Features: 1 (Done: 0, WIP: 0, Todo: 1)
- Completion: 0.0%

### --- MCP Protocol ---

- Features: 2 (Done: 0, WIP: 0, Todo: 2)
- Completion: 0.0%

### --- Release & Distribution ---

- Features: 1 (Done: 0, WIP: 0, Todo: 1)
- Completion: 0.0%

### --- Skills & Reports ---

- Features: 2 (Done: 0, WIP: 0, Todo: 2)
- Completion: 0.0%

### --- XRAY Engine ---

- Features: 3 (Done: 0, WIP: 0, Todo: 3)
- Completion: 0.0%

⸻

## Critical Path Analysis

The following features are blocked by incomplete dependencies:

- `CLI_COMMAND_COMMIT` blocked by: CLI_CONTRACT
- `CLI_COMMAND_CONTEXT` blocked by: CLI_CONTRACT
- `CLI_COMMAND_FEATURE` blocked by: CLI_CONTRACT
- `CLI_COMMAND_FEATURES` blocked by: CLI_CONTRACT
- `CLI_COMMAND_GOV` blocked by: CLI_CONTRACT
- `CLI_COMMAND_RUN` blocked by: CLI_CONTRACT
- `CLI_COMMAND_STATUS` blocked by: CLI_CONTRACT
- `CLI_CONTRACT` blocked by: CORE_REPO_CONTRACT
- `MCP_ROUTER_CONTRACT` blocked by: CORE_REPO_CONTRACT
- `MCP_TOOLS` blocked by: MCP_ROUTER_CONTRACT
- `REL_ARTIFACT_LAYOUT` blocked by: CORE_REPO_CONTRACT
- `REPORTS_CORE` blocked by: CLI_COMMAND_COMMIT
- `SKILLS_REGISTRY` blocked by: CLI_COMMAND_RUN
- `XRAY_CLI` blocked by: XRAY_SCAN_POLICY
- `XRAY_INDEX_FORMAT` blocked by: CORE_REPO_CONTRACT
- `XRAY_SCAN_POLICY` blocked by: XRAY_INDEX_FORMAT

## Next Steps

1. Use `cortex status roadmap` to regenerate this document whenever `spec/features.yaml` changes.
2. Prioritize unblocking critical-path features.
3. Complete partially implemented phases before starting new ones.
