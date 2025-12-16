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
package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartekus/cortex/internal/reports/featuretrace"
)

func TestRunFeatureTraceability_GeneratesReport(t *testing.T) {
	// Not parallel: uses os.Chdir which is global process state

	// Create temporary repository structure
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "stagecraft")
	if err := os.Mkdir(repoDir, 0o750); err != nil {
		t.Fatalf("creating repo dir: %v", err)
	}

	// Create .stagecraft/reports directory
	reportsDir := filepath.Join(repoDir, ".stagecraft", "reports")
	if err := os.MkdirAll(reportsDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create reports directory: %v", err)
	}

	// Create spec file with Feature header
	specDir := filepath.Join(repoDir, "spec", "commands")
	if err := os.MkdirAll(specDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create spec directory: %v", err)
	}

	// Create go.mod marker
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module test"), 0o600); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	specContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

# Deploy Command
`
	if err := os.WriteFile(filepath.Join(specDir, "deploy.md"), []byte(specContent), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Create implementation file
	implDir := filepath.Join(repoDir, "internal", "core")
	if err := os.MkdirAll(implDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create internal directory: %v", err)
	}

	implContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

package core
`
	if err := os.WriteFile(filepath.Join(implDir, "deploy.go"), []byte(implContent), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write impl file: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Run the feature traceability command
	cmd := NewFeatureTraceabilityCommand()
	if err := runFeatureTraceability(cmd, nil); err != nil {
		t.Fatalf("runFeatureTraceability failed: %v", err)
	}

	// Verify the report file was created
	reportPath := filepath.Join(repoDir, ".stagecraft", "reports", "feature-traceability.json")
	data, err := os.ReadFile(reportPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("failed to read report file: %v", err)
	}

	// Unmarshal into the concrete report type to ensure the JSON matches the schema
	var report featuretrace.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("failed to unmarshal report JSON: %v", err)
	}

	// Basic structural assertions for determinism and correctness
	if report.SchemaVersion != "1.0" {
		t.Errorf("expected schema_version=1.0, got %s", report.SchemaVersion)
	}

	if report.Summary.TotalFeatures != 1 {
		t.Errorf("expected total_features=1, got %d", report.Summary.TotalFeatures)
	}
}

func TestRunFeatureTraceability_ReportStructure(t *testing.T) {
	t.Parallel()

	// Test that the generated report has the expected structure
	features := []featuretrace.FeaturePresence{
		{
			FeatureID:           "CLI_DEPLOY",
			Status:              featuretrace.FeatureStatusDone,
			HasSpec:             true,
			SpecPath:            "spec/commands/deploy.md",
			ImplementationFiles: []string{"internal/core/deploy.go"},
			TestFiles:           []string{"internal/core/deploy_test.go"},
			CommitSHAs:          []string{"abc123"},
		},
	}

	report, err := featuretrace.GenerateFeatureTraceabilityReport(features)
	if err != nil {
		t.Fatalf("GenerateFeatureTraceabilityReport failed: %v", err)
	}

	// Verify report structure
	if report.SchemaVersion != "1.0" {
		t.Errorf("expected schema_version=1.0, got %s", report.SchemaVersion)
	}

	if report.Summary.TotalFeatures != 1 {
		t.Errorf("expected total_features=1, got %d", report.Summary.TotalFeatures)
	}

	// Verify report can be marshaled to JSON
	_, err = json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report to JSON: %v", err)
	}
}

func TestRunFeatureTraceability_GoldenFile(t *testing.T) {
	// Not parallel: uses os.Chdir which is global process state

	// Create temporary repository structure
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "stagecraft")
	if err := os.Mkdir(repoDir, 0o750); err != nil {
		t.Fatalf("creating repo dir: %v", err)
	}

	// Create .stagecraft/reports directory
	reportsDir := filepath.Join(repoDir, ".stagecraft", "reports")
	if err := os.MkdirAll(reportsDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create reports directory: %v", err)
	}

	// Create spec file with Feature header
	specDir := filepath.Join(repoDir, "spec", "commands")
	if err := os.MkdirAll(specDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create spec directory: %v", err)
	}

	// Create go.mod marker
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module test"), 0o600); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	specContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

# Deploy Command
`
	if err := os.WriteFile(filepath.Join(specDir, "deploy.md"), []byte(specContent), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Create implementation file
	implDir := filepath.Join(repoDir, "internal", "core")
	if err := os.MkdirAll(implDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create internal directory: %v", err)
	}

	implContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

package core
`
	if err := os.WriteFile(filepath.Join(implDir, "deploy.go"), []byte(implContent), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write impl file: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Run the feature traceability command
	cmd := NewFeatureTraceabilityCommand()
	if err := runFeatureTraceability(cmd, nil); err != nil {
		t.Fatalf("runFeatureTraceability failed: %v", err)
	}

	// Read the generated report file
	reportPath := filepath.Join(repoDir, ".stagecraft", "reports", "feature-traceability.json")
	data, err := os.ReadFile(reportPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("failed to read report file: %v", err)
	}

	// Format JSON for deterministic comparison (indent with 2 spaces)
	var report featuretrace.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("failed to unmarshal report JSON: %v", err)
	}

	// Re-marshal with consistent formatting for golden comparison
	formatted, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal report: %v", err)
	}
	formatted = append(formatted, '\n') // Add trailing newline for consistency

	// Compare with golden file
	expected := readGoldenFile(t, "feature_traceability_report")

	if *updateGolden {
		writeGoldenFile(t, "feature_traceability_report", string(formatted))
		expected = string(formatted)
	}

	if string(formatted) != expected {
		t.Errorf("report mismatch:\nGot:\n%s\nExpected:\n%s", string(formatted), expected)
	}
}
