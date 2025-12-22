package projectroot

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// Feature: CORE_REPO_CONTRACT
// Spec: spec/system/contract.md

func TestCoreRepoContract(t *testing.T) {
	root, err := Find(".")
	if err != nil {
		t.Fatalf("failed to get project root: %v", err)
	}

	// 1. Assert Critical Files Exist
	criticalFiles := []string{
		"Makefile",
		".github/workflows/ci.yml",
		"go.mod",
		"rust/Cargo.toml",
	}

	for _, relPath := range criticalFiles {
		path := filepath.Join(root, relPath)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing critical contract file: %s", relPath)
		}
	}

	// 2. Assert Makefile Targets (using simple anchors)
	makefilePath := filepath.Join(root, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}
	sContent := string(content)

	requiredTargets := []string{
		`^repo:`,
		`^gov:`,
		`^context:`,
		`^test:`,
		`^build:`,
	}

	for _, targetRegex := range requiredTargets {
		re := regexp.MustCompile("(?m)" + targetRegex)
		if !re.MatchString(sContent) {
			t.Errorf("Makefile missing required target definition matching %q", targetRegex)
		}
	}
}
