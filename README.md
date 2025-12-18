# Cortex

Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

It serves as the intelligent operational layer for the [Stagecraft](https://github.com/bartekus/stagecraft) ecosystem, providing deterministic context generation, repository scanning (via XRAY), and governance enforcement.

## Quickstart

```bash
# Verify environment (runs tests, lint, format checks)
make lint
make test

# Build binaries (bin/cortex, rust/target/release/xray, rust/target/release/cortex-mcp)
make build

# Run the CLI
./bin/cortex --help
```
./bin/cortex --help
```

## Verification

To verify the installation matches the release artifact:

1. Download the archive for your OS/Arch.
2. Extract the archive.
3. Run the binary:
   ```bash
   ./cortex --version
   ```
4. Verify the version matches the release tag.

## Contributing

We prioritize determinism and parity between local and CI environments.

- **Go**: Formatted with `gofumpt` and `goimports`.
- **Rust**: Formatted with `cargo fmt`.
- **Lint**: Regulated by `golangci-lint` (default presets) and `cargo clippy` (strict).

See `Makefile` for the canonical commands used in CI.
