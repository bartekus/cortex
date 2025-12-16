package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bartekus/cortex/internal/projectroot"
)

func TestIntegrationXrayScan(t *testing.T) {
	// Find repo root
	root, err := projectroot.Find(".")
	if err != nil {
		t.Fatalf("Failed to find repo root: %v", err)
	}

	// Locate binary (must be built)
	bin := filepath.Join(root, "ai.agent/rust/xray/target/debug/xray")
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: xray binary not found at %s. Build it to run this test.", bin)
	}

	// Use fixture as target
	fixturePath := filepath.Join(root, "ai.agent/rust/xray/tests/fixtures/min_repo")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Fatalf("Fixture not found at %s", fixturePath)
	}

	// Output dir (Temp)
	outDir := t.TempDir()

	// Run command: xray scan <fixture> --output <temp> --xray-bin <bin>
	cmd := NewContextXrayCommand()
	// Inject flags. Target is positional.
	// Set flag explicitly to bypass Cobra parsing complexities in test harness with DisableFlagParsing
	_ = cmd.PersistentFlags().Set("xray-bin", bin)
	cmd.SetArgs([]string{"scan", fixturePath, "--output", outDir})

	// Capture stdout/stderr to avoid polluting test output
	// cmd.SetOut(io.Discard)
	// cmd.SetErr(io.Discard)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify index.json exists
	indexPath := filepath.Join(outDir, "index.json")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("Expected index.json at %s, but missing", indexPath)
	}

	// Read content and check for minimal expected JSON
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index.json: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "\"schemaVersion\"") {
		t.Errorf("Index JSON missing schemaVersion")
	}
}

// TestIntegrationContextBuild validates the full 'context build' flow hermetically.
func TestIntegrationContextBuild(t *testing.T) {
	// 1. Setup Temp Repo
	root := t.TempDir()

	// 2. Locate XRAY Binary (must exist)
	repoRoot, err := projectroot.Find(".")
	if err != nil {
		t.Fatalf("Failed to find repo root: %v", err)
	}
	bin := filepath.Join(repoRoot, "ai.agent/rust/xray/target/debug/xray")
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		t.Skipf("Skipping integration test: xray binary not found at %s. Build it to run this test.", bin)
	}

	// 3. Populate Temp Repo with Fixture Data (min_repo)
	fixturePath := filepath.Join(repoRoot, "ai.agent/rust/xray/tests/fixtures/min_repo")
	// Simple copy of relevant files to temp root
	// We need .git or markers? projectroot.Find checks for .git, go.mod, etc.
	// XRAY only cares if target is a dir. But 'context build' calls projectroot.Find(".")?
	// Wait, 'context build' calls projectroot.Find(".") internally.
	// If we run the command, we must set its CWD to `root` (temp dir).
	// But projectroot.Find(".") looks for markers.
	// So we must create a marker in `root`.
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("Failed to create .git marker: %v", err)
	}

	// Copy fixture files
	// README.md and main.go
	copyFile(t, filepath.Join(fixturePath, "README.md"), filepath.Join(root, "README.md"))
	copyFile(t, filepath.Join(fixturePath, "main.go"), filepath.Join(root, "main.go"))

	// 4. Run `context build`
	// We must pass --xray-bin via persistent flag on parent or via env?
	// We added "xray-bin" to NewContextCommand (parent).
	// So we can set it there.
	cmd := NewContextCommand()
	// cmd is root for context. build is subcommand.
	// We want to execute `context build`.
	// Arguments: "build"
	// Flags: --xray-bin <bin>
	cmd.SetArgs([]string{"build", "--xray-bin", bin})

	// Hack: We need to ensure `projectroot.Find(".")` works inside the command.
	// Since `projectroot.Find` checks CWD, and we cannot easily change process CWD safely in parallel tests.
	// However, `runContextBuild` calls `projectroot.Find(".")`.
	// If we cannot change CWD, we might fail.
	// `exec.Command` has Dir, but `projectroot.Find` uses `os.Getwd()`.
	// This is a common Go testing pain.
	// Option: Change `runContextBuild` to accept a root arg? No, CLI contract.
	// Option: Use `os.Chdir` in test (and logical lock/t.Cleanup)?
	// `projectroot.Find(".")` uses `filepath.Abs(".")`.
	// I will use `os.Chdir` but with t.Parallel() disabled (default).
	// Ideally verifying logic without chdir.
	// But `runContextBuild` logic is: `repoRoot, err := projectroot.Find(".")`.
	// So I MUST chdir for this test to target the temp repo.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Chdir(cwd)
	})

	// Silence output
	// cmd.SetOut(io.Discard)
	// cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("context build failed: %v", err)
	}

	// 5. Assertions
	ctxDir := filepath.Join(root, ".ai-context")
	assertExists(t, filepath.Join(ctxDir, "meta.json"))
	assertExists(t, filepath.Join(ctxDir, "files", "manifest.json"))
	assertExists(t, filepath.Join(ctxDir, "files", "chunks.ndjson")) // Checks Phase 4B
	assertExists(t, filepath.Join(ctxDir, "digest.txt"))

	// Check manifest sorted
	// (README.md, main.go)
	mBytes, _ := os.ReadFile(filepath.Join(ctxDir, "files", "manifest.json"))
	if !strings.Contains(string(mBytes), "main.go") {
		t.Errorf("Manifest missing main.go")
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, in, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file at %s, but missing", path)
	}
}
