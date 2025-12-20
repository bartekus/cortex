---
feature: SKILLS_REGISTRY
version: v1
status: approved
domain: skills
inputs:
  args:
    - name: skill_id
outputs:
  exit_codes:
    0: 0
    1: 1
---
# Skills Registry
## Summary
The Skills Registry defines the set of executable skills available to the `cortex run` command. Skills are atomic, identifiable units of work (linting, testing, formatting, checking).

## Surface
- **Invoke**: `cortex run <skill_id>`

## Registry
| ID | Type | Description |
| :--- | :--- | :--- |
| `docs:doc-patterns` | Governance | Validates documentation naming and structure. |
| `docs:feature-integrity` | Governance | Validates feature registry integrity. |
| `docs:header-comments` | Governance | Checks file headers (SPDX/Frontmatter). |
| `docs:orphan-docs` | Governance | Detects unlinked documentation files. |
| `docs:orphan-specs` | Governance | Detects specs not referenced in features.yaml. |
| `docs:policy` | Governance | General documentation policy checks. |
| `docs:provider-governance` | Governance | Provider-specific governance. |
| `docs:validate-spec` | Governance | Validates specification syntax. |
| `docs:yaml` | Governance | Lints YAML files. |
| `format:gofumpt` | Formatter | Formats Go code using gofumpt. |
| `lint:gofumpt` | Linter | Checks Go code formatting. |
| `lint:golangci` | Linter | Runs golangci-lint. |
| `purity` | Governance | Checks for non-deterministic artifacts. |
| `test:basic` | Test | Runs basic unit tests. |
| `test:coverage` | Test | Runs tests with coverage analysis. |

## References
- `internal/skills/`
