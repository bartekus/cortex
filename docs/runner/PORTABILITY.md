# Cortex Runner Portability & Repo-Sensitivity

## Overview
The Cortex Runner is designed to operate deterministically across different repository environments, specifically:
1.  **Cortex Monorepo**: The full-featured production environment with complete `spec/`, `docs/`, `pkg/`, and `internal/` structures.
2.  **Cortex Helper/Standalone**: A subset repository (like the one you are currently in) which may lack cortex-specific governance files.
3.  **Future Poly-repos**: Specialized repositories with limited scope.

## The SKIP Contract
To maintain portability, every Skill MUST adhere to the **SKIP Contract**:

> If a Skill's prerequisites (e.g., target directories, config files, binaries) are missing, it MUST return `StatusSkip` with a clear Note, NOT `StatusFail`.

### Examples
- **`test:binary`**: Checks for `cmd/cortex`. If missing, checks for `cmd/cortex`. If neither, `SKIP "No known binary target"`.
- **`docs:validate-spec`**: Checks for `spec/` directory. If missing, `SKIP "No spec directory found"`.
- **`docs:feature-integrity`**: Checks for `spec/features.yaml`. If missing, `SKIP "spec/features.yaml not found"`.

### Why?
This ensures that `cortex run all` is safe to run anywhere. It acts as a "capabilities negotiation" between the runner and the repo.
- In **Cortex**, it runs the full governance suite.
- In **Cortex** (this repo), it checks code quality and build health but skips irrelevant product specs.

## Implementation Guidelines
1.  **Repo Root Anchoring**: All paths must be relative to `deps.RepoRoot` or `deps.StateDir`. Never assume CWD.
2.  **Explicit Checks**: Use `os.Stat` or `scanner.FilterOptions` to verify existence before proceeding.
3.  **Filtered Scanning**: Use `scanner.TrackedFilesFiltered` to only operate on files Git knows about, respecting `.gitignore` automatically.
