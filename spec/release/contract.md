---
feature: REL_ARTIFACT_LAYOUT
version: v1
status: approved
domain: release
inputs:
  build_system:
    - goreleaser
    - github-actions
outputs:
  artifacts:
    - cortex_<version>_<os>_<arch>.tar.gz
    - cortex_<version>_windows_<arch>.zip
    - checksums.txt
---
# Release Artifact Layout

**Feature**: `REL_ARTIFACT_LAYOUT`
**Status**: Approved

## Purpose
Defines the structure, naming, and content of the artifacts produced during a release. Consumers (installers, distributors) rely on this contract.

## Interface: Artifacts
Releases are built via GoReleaser and GitHub Actions `release.yml`.

### Binary Assets
| Artifact Name | Content | Platforms |
| :--- | :--- | :--- |
| `cortex_<version>_<os>_<arch>.tar.gz` | `cortex` CLI binary + `README.md` + `LICENSE` + embedded Rust helpers | Linux/macOS |
| `cortex_<version>_windows_<arch>.zip` | `cortex.exe` + `README.md` + `LICENSE` + embedded Rust helpers | Windows |
| `checksums.txt` | SHA256 checksums for all artifacts | All |

### Bundled Components
The `cortex` archive **MUST** contain the following helper binaries alongside the main CLI:
- `xray` (Platform-specific binary)
- `cortex-mcp` (Platform-specific binary)

*Note: The presence of these sidecars is required for the `cortex context` and `cortex xray` commands to function.*

## Guarantees
1.  **Versioning**: Version string matches the git tag (semver `vX.Y.Z`).
2.  **Integrity**: `checksums.txt` is generated after all artifacts are finalized.

## Examples
**Layout of `cortex_1.0.0_darwin_arm64.tar.gz`**:
```text
.
├── cortex          # Main CLI
├── LICENSE
├── README.md
├── xray            # Helper (bundled)
└── cortex-mcp      # Helper (bundled)
```
