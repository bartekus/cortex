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
- `internal/skills/docs_doc_patterns.go`
- `internal/skills/docs_feature_integrity.go`
- `internal/skills/docs_header_comments.go`
- `internal/skills/docs_orphan_docs.go`
- `internal/skills/docs_orphan_specs.go`
- `internal/skills/docs_policy.go`
- `internal/skills/docs_provider_governance.go`
- `internal/skills/docs_validate_spec.go`
- `internal/skills/docs_yaml.go`
- `internal/skills/format_gofumpt.go`
- `internal/skills/lint_gofumpt.go`
- `internal/skills/lint_golangci.go`
- `internal/skills/purity.go`
- `internal/skills/registry.go`
- `internal/skills/test_basic.go`
- `internal/skills/test_coverage.go`
