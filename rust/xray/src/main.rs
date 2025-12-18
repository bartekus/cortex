use anyhow::Result;
use clap::{Parser, Subcommand};
use std::path::PathBuf;

mod canonical;
mod digest;
mod hash;
mod language;
mod loc;
/// # XRAY Determinism Policy
///
/// This crate enforces strict determinism for repository scanning.
///
/// ## Locked Policies
/// 1. **LOC Counting**: Logical lines (`str::lines().count()`). Distinct from POSIX `wc -l`.
/// 2. **Language Aggregation**: Files with "Unknown" language are EXCLUDED from the `languages` map.
/// 3. **Canonical JSON**: Output MUST be sorted. `serde_json` MUST have `preserve_order` feature enabled.
/// 4. **Digest Integrity**: The digest is computed on the *Canonical JSON* representation of the index.
///    The index MUST be strictly sorted (files by path, modules by path) before digest computation.
///    The digest function REFUSES to process unsorted inputs (no silent repairs).
/// 5. **Derived Fields**: `languages` and `top_dirs` are derived summaries. While generated deterministically,
///    they are currently NOT strictly validated against the `files` list during digest computation.
///
/// ## Architecture
/// - **Producer** (`traversal`): Responsible for producing valid, sorted, normalized data.
/// - **Consumer** (`digest`, `canonical`): Responsible for VALIDATING invariants. Fails if invalid.
mod schema;
mod traversal;
mod write;

#[cfg(test)]
mod invariant_tests;

#[derive(Parser)]
#[command(name = "xray")]
#[command(about = "Deterministic repository scanner", long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Scans the repository and updates .xraycache
    Scan {
        /// Target directory to scan (default: .)
        #[arg(default_value = ".")]
        target: String,

        /// Output directory override
        #[arg(long)]
        output: Option<String>,
    },
    /// Generate documentation (placeholder)
    Docs,
    /// Run all steps (placeholder)
    All,
}

fn main() -> Result<()> {
    let cli = Cli::parse();

    match cli.command {
        Commands::Scan { target, output } => run_scan(&target, output),
        Commands::Docs => {
            println!("Docs generation not implemented yet");
            Ok(())
        }
        Commands::All => {
            println!("All steps not implemented yet");
            Ok(())
        }
    }
}

fn run_scan(target: &str, output: Option<String>) -> Result<()> {
    let repo_root = std::env::current_dir()?;
    let repo_slug = repo_root
        .file_name()
        .unwrap_or_default()
        .to_string_lossy()
        .to_string();

    // 1. Scan Target
    let target_path = PathBuf::from(target);
    let scan_result = traversal::scan_target(&target_path)?;

    // 2. Build Index
    let mut index = schema::XrayIndex {
        root: repo_slug.clone(),
        target: target.to_string(),
        files: scan_result.files,
        stats: scan_result.stats,
        languages: scan_result.languages,
        top_dirs: scan_result.top_dirs,
        module_files: scan_result.module_files,
        ..Default::default()
    };

    // 3. Compute digest
    let digest_str = digest::calculate_digest(&index)?;
    index.digest = digest_str;

    // 4. Serialize
    let bytes = canonical::to_canonical_json(&index)?;

    // 5. Determine output path
    // Default: .xraycache/<slug>/data/index.json
    let out_dir = match output {
        Some(p) => PathBuf::from(p),
        None => repo_root.join(".xraycache").join(&repo_slug).join("data"),
    };

    let out_file = out_dir.join("index.json");

    // 6. Write
    write::write_atomic(&out_file, &bytes)?;

    println!("XRAY scan complete. Digest: {}", index.digest);
    println!("Written to: {}", out_file.display());

    Ok(())
}
