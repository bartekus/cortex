# Contributing to Cortex

- **Go**: v1.24.10+ (See `go.mod`)
- **Rust**: Stable
- **Make**: GNU Make 3.81+

## Workflow

1. **Setup**: Install required linters and tools.
   ```bash
   make tools-install
   ```

2. **Verify**: Run all checks before pushing.
   ```bash
   make lint test
   ```

3. **Build**: Ensure local build works.
   ```bash
   make build
   ```
