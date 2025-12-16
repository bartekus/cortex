# Cortex Library

Core primitives for Stagecraft's developer and governance tooling.

## API Surface

### `projectroot`

Resolves the repository root based on strict priority markers.

*   `Find(startPath string) (string, error)`

### `featureindex`

Loads the feature registry from the repository contract.

*   `Load(rootPath string) (Registry, error)`

### `projectmeta`

Determines repository metadata.

*   `DetermineRepoName(rootPath string) string`

## Rules
*   **Pure Go**: No dependencies on `stagecraft/internal/...`.
*   **Reusable**: Can be used by other tools/projects.
