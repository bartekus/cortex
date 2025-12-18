Cortex repo layout:

(./) - Go Root
(./rust) - Rust Root

From Rust Root:

Input:
```bash
cargo test -p cortex-mcp -p xray
cargo test -p cortex-mcp -p xray --tests
cargo build -p cortex-mcp -p xray --release
RUSTFLAGS="-D warnings" cargo build -p cortex-mcp -p xray --release
```
