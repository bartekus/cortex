// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_FRONTEND_GENERIC
// Spec: spec/providers/frontend/generic.md
// Docs: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md
package featuretrace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFeaturePresence_DeterministicOrdering(t *testing.T) {
	t.Parallel()

	// Create a temporary test repository structure
	tmpDir := t.TempDir()

	// Create nested directory structure to test OS-independent ordering
	err := os.MkdirAll(filepath.Join(tmpDir, "spec", "commands"), 0o750) //nolint:gosec // G301: test directory
	if err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, "internal", "core"), 0o750) //nolint:gosec // G301: test directory
	if err != nil {
		t.Fatalf("failed to create internal dir: %v", err)
	}

	// Create spec file with Feature header
	specContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

# Deploy Command
`
	err = os.WriteFile(filepath.Join(tmpDir, "spec", "commands", "deploy.md"), []byte(specContent), 0o600) //nolint:gosec // G306: test file
	if err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Create implementation file
	implContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

package core
`
	err = os.WriteFile(filepath.Join(tmpDir, "internal", "core", "deploy.go"), []byte(implContent), 0o600) //nolint:gosec // G306: test file
	if err != nil {
		t.Fatalf("failed to write impl file: %v", err)
	}

	// Create test file
	testContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

package core
`
	err = os.WriteFile(filepath.Join(tmpDir, "internal", "core", "deploy_test.go"), []byte(testContent), 0o600) //nolint:gosec // G306: test file
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create another feature to test sorting
	anotherSpecContent := `// Feature: CLI_PLAN
// Spec: spec/commands/plan.md

# Plan Command
`
	err = os.WriteFile(filepath.Join(tmpDir, "spec", "commands", "plan.md"), []byte(anotherSpecContent), 0o600) //nolint:gosec // G306: test file
	if err != nil {
		t.Fatalf("failed to write another spec file: %v", err)
	}

	// Scan
	cfg := ScanConfig{RootDir: tmpDir}
	features, err := ScanFeaturePresence(cfg)
	if err != nil {
		t.Fatalf("ScanFeaturePresence failed: %v", err)
	}

	// Verify results are sorted lexicographically by FeatureID
	if len(features) < 2 {
		t.Fatalf("expected at least 2 features, got %d", len(features))
	}

	// Verify sorting: CLI_DEPLOY should come before CLI_PLAN
	if features[0].FeatureID > features[1].FeatureID {
		t.Errorf("features not sorted: got %s before %s", features[0].FeatureID, features[1].FeatureID)
	}

	// Verify CLI_DEPLOY has correct presence flags
	var deployFeature *FeaturePresence
	for i := range features {
		if features[i].FeatureID == "CLI_DEPLOY" {
			deployFeature = &features[i]
			break
		}
	}

	if deployFeature == nil {
		t.Fatal("CLI_DEPLOY feature not found")
	}

	if !deployFeature.HasSpec {
		t.Error("CLI_DEPLOY should have spec")
	}
	if len(deployFeature.ImplementationFiles) == 0 {
		t.Error("CLI_DEPLOY should have implementation files")
	}
	if len(deployFeature.TestFiles) == 0 {
		t.Error("CLI_DEPLOY should have test files")
	}
}

func TestScanFeaturePresence_NoFeatureHeader(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create file without Feature header
	content := `package main

func main() {}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(content), 0o600) //nolint:gosec // G306: test file
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	cfg := ScanConfig{RootDir: tmpDir}
	features, err := ScanFeaturePresence(cfg)
	if err != nil {
		t.Fatalf("ScanFeaturePresence failed: %v", err)
	}

	// File without Feature header should be ignored
	if len(features) != 0 {
		t.Errorf("expected 0 features for file without header, got %d", len(features))
	}
}

func TestExtractFeatureIDFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "valid feature header",
			content:  "// Feature: CLI_DEPLOY\n// Spec: spec/commands/deploy.md\n",
			expected: "CLI_DEPLOY",
		},
		{
			name:     "no feature header",
			content:  "package main\n",
			expected: "",
		},
		{
			name:     "feature header in middle of file",
			content:  "package main\n// Feature: CLI_DEPLOY\nfunc main() {}\n",
			expected: "", // Only header comments should be considered
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpFile := filepath.Join(t.TempDir(), "test.go")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0o600) //nolint:gosec // G306: test file
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			featureID, err := extractFeatureIDFromFile(tmpFile)
			if err != nil {
				t.Fatalf("extractFeatureIDFromFile failed: %v", err)
			}

			if featureID != tt.expected {
				t.Errorf("expected feature ID %q, got %q", tt.expected, featureID)
			}
		})
	}
}
