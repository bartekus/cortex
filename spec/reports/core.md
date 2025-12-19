---
feature: REPORTS_CORE
version: v1
status: approved
domain: reports
inputs: []
outputs:
  files:
    - .cortex/reports/*.json
---
# Reports: Core
## Summary
Cortex generates structured JSON reports to document repository health, feature traceability, and commit discipline.

## Reports

### Commit Health
- **File**: `.cortex/reports/commit-health.json`
- **Generator**: `cortex commit report`
- **Content**: analysis of commit messages, authorship, and conventional commit adherence.

### Feature Traceability
- **File**: `.cortex/reports/feature-traceability.json`
- **Generator**: `cortex feature`
- **Content**: Mapping of Feature IDs to commits, files, and tests.

## References
- `internal/reports/commithealth`
- `internal/reports/featuretrace`
