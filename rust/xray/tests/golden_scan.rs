use std::env;
use std::fs;
use std::path::PathBuf;
use std::process::Command;

// We use the binary itself or the library logic?
// Ideally integration tests run the binary. But for unit testing logic we can import modules if lib.rs exists.
// Since main.rs is a binary, we'll run it as a subprocess or we need to extract logic to lib.rs.
// For now, let's run the binary content using `cargo run` style or just extraction.
// Actually, to keep it simple, I will invoke the `cargo run` command in this test to simulate CLI usage,
// OR I can refactor main.rs to expose `run_scan`.
// Given the constraints, I will refactor `main.rs` slightly to verify logic, OR just run the build.
// A simpler robust way for "Golden Test" is indeed running the binary.

#[test]
fn test_determinism_empty_scan() {
    // 1. Setup paths
    let manifest_dir = env::var("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR not set");
    let target_dir = PathBuf::from(&manifest_dir).join("target/debug/xray");
    let fixture_dir = PathBuf::from(&manifest_dir).join("tests/fixtures/min_repo");
    let output_dir_1 = PathBuf::from(&manifest_dir).join("tests/outputs/run1");
    let output_dir_2 = PathBuf::from(&manifest_dir).join("tests/outputs/run2");

    // Clean outputs
    let _ = fs::remove_dir_all(&output_dir_1);
    let _ = fs::remove_dir_all(&output_dir_2);

    // Build binary (ensure it is up to date)
    let status = Command::new("cargo")
        .arg("build")
        .current_dir(&manifest_dir)
        .status()
        .expect("Failed to build xray");
    assert!(status.success(), "Build failed");

    // 2. Run Scan 1
    let status1 = Command::new("cargo")
        .arg("run")
        .arg("--")
        .arg("scan")
        .arg(&fixture_dir)
        .arg("--output")
        .arg(&output_dir_1)
        .current_dir(&manifest_dir)
        .status()
        .expect("Failed to run xray scan 1");
    assert!(status1.success());

    // 3. Run Scan 2
    let status2 = Command::new("cargo")
        .arg("run")
        .arg("--")
        .arg("scan")
        .arg(&fixture_dir)
        .arg("--output")
        .arg(&output_dir_2)
        .current_dir(&manifest_dir)
        .status()
        .expect("Failed to run xray scan 2");
    assert!(status2.success());

    // 4. Compare Outputs
    let file1 = output_dir_1.join("index.json");
    let file2 = output_dir_2.join("index.json");

    let content1 = fs::read_to_string(&file1).expect("Failed to read output 1");
    let content2 = fs::read_to_string(&file2).expect("Failed to read output 2");

    assert_eq!(content1, content2, "Outputs are not identical!");
    
    // 5. Verify Content (Basic)
    assert!(content1.contains("\"schemaVersion\":\"1.0.0\""));
    assert!(!content1.contains("indexedAt")); // Forbidden field check

    // 6. Verify Traversal and Ignore Logic
    // Should contain main.go
    assert!(content1.contains("\"path\":\"main.go\""), "main.go missing from index");
    // Verify Hash (Phase B)
    let expected_hash = "sha256:55a60bb97151b2b4b680462447ce60ec34511b14fa10d77440c97b9777101566";
    assert!(content1.contains(expected_hash), "main.go hash incorrect or missing");

    // Should NOT contain vendor/ignored.txt (it is in strict ignore list)
    assert!(!content1.contains("vendor/ignored.txt"), "vendor/ignored.txt should be skipped");
    assert!(!content1.contains("ignored.txt"), "ignored.txt should be skipped");
    
    // 7. Verify Phase C1 Aggregation
    assert!(content1.contains("\"Go\":1"), "Language Go missing");
    assert!(content1.contains("\"Rust\":1"), "Language Rust missing");
    assert!(content1.contains("\"Markdown\":1"), "Language Markdown missing");
    assert!(content1.contains("\"JSON\":1"), "Language JSON missing");
    
    assert!(content1.contains("\"cmd\":1"), "TopDir cmd missing");
    // "." might be implicit or explicit depending on map serialization.
    // "moduleFiles":["package.json"]
    assert!(content1.contains("\"package.json\""), "package.json missing from module_files");
    assert!(content1.contains("\".git\""), ".git missing from module_files");
}
