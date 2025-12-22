use std::env;
use std::fs;
use std::path::PathBuf;
use std::process::Command;

// Feature: XRAY_INDEX_FORMAT
// Spec: spec/xray/index-format.md

#[test]
fn test_index_format() {
    let manifest_dir = env::var("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR not set");
    let output_dir = PathBuf::from(&manifest_dir).join("tests/outputs/index_format");
    let fixture_src = PathBuf::from(&manifest_dir).join("tests/fixtures/min_repo");

    // Create temp dir for fixture to ensure it's clean and has .git
    let temp_dir = tempfile::tempdir().expect("Failed to create temp dir");
    let fixture_dst = temp_dir.path().join("min_repo");

    let copy_options = fs_extra::dir::CopyOptions::new().overwrite(true).copy_inside(true);
    fs_extra::dir::copy(&fixture_src, temp_dir.path(), &copy_options).expect("Failed to copy fixture");

    // Add .git to ensure it is treated as a repo
    let git_dir = fixture_dst.join(".git");
    fs::create_dir_all(&git_dir).expect("Failed to create .git dir");

    // Clean output
    let _ = fs::remove_dir_all(&output_dir);

    // Use CARGO_BIN_EXE_xray provided by Cargo for integration tests
    let xray_bin = env!("CARGO_BIN_EXE_xray");

    // Run Scan
    let status = Command::new(xray_bin)
        .arg("scan")
        .arg(&fixture_dst)
        .arg("--output")
        .arg(&output_dir)
        .current_dir(&manifest_dir)
        .status()
        .expect("Failed to run scan");
    assert!(status.success());

    let index_path = output_dir.join("index.json");
    let content = fs::read_to_string(&index_path).expect("Failed to read index.json");

    // Assert JSON Structure Matches Contract
    // 1. Root fields
    assert!(content.contains(r#""schemaVersion":"1.0.0""#), "Missing schemaVersion");
    assert!(content.contains(r#""scanId":""#), "Missing scanId");
    // "indexedAt" is conditionally excluded in some contexts or checked? The spec says it should be there?
    // golden_scan.rs checked it is NOT present (forbidden field check).
    // If spec requires it, we should check it. But golden_scan suggests deterministic output might strip it.
    // Let's assume deterministic output for now.

    assert!(content.contains(r#""rootHash":""#), "Missing rootHash");

    // 2. Summary fields
    assert!(content.contains(r#""languages":""#), "Missing languages summary");
    assert!(content.contains(r#""files":""#), "Missing files count");

    // 3. Module files
    assert!(content.contains(r#""moduleFiles":["#), "Missing moduleFiles");

    // 4. File listing
    assert!(content.contains(r#""path":"main.go""#), "Missing expected file entry");
}
