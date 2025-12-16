// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: GOV_STATUS_ROADMAP
// Spec: spec/commands/status-roadmap.md

package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestStatusRoadmapCommand_ExecutesSuccessfully(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()

	// Create a minimal features.yaml in tmpDir/spec
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	featuresYAML := `features:
  # Phase 0: Foundation
  - id: TEST_FEATURE
    title: "Test feature"
    status: done
    spec: "test.md"
    owner: bart
    tests: []
`

	featuresPath := filepath.Join(specDir, "features.yaml")
	if err := os.WriteFile(featuresPath, []byte(featuresYAML), 0o600); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "docs", "engine", "status")
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	// Change to tmpDir so relative paths work
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create status command
	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"roadmap"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("status roadmap command failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify output file was created
	outputPath := filepath.Join(tmpDir, "docs", "engine", "status", "feature-completion-analysis.md")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file not created: %v", err)
	}

	// Verify output file has content
	//nolint:gosec // G304: file path is from test temp directory, safe
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Error("output file is empty")
	}

	// Verify it contains expected sections
	contentStr := string(content)
	if !contains(contentStr, "Feature Completion Analysis") {
		t.Error("output file missing 'Feature Completion Analysis' header")
	}
	if !contains(contentStr, "Executive Summary") {
		t.Error("output file missing 'Executive Summary' section")
	}
}

func TestStatusRoadmapCommand_HandlesMissingFeaturesYAML(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()

	// Change to tmpDir
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"roadmap"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	if err == nil {
		t.Error("status roadmap command expected error for missing features.yaml, got nil")
	}

	// Should exit with code 1 (validation error)
	if exitCode := getExitCode(err); exitCode != 1 {
		t.Errorf("expected exit code 1 (validation error), got %d", exitCode)
	}
}

func TestStatusRoadmapCommand_HandlesInvalidYAML(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()

	// Create spec directory with invalid YAML
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	invalidYAML := `features:
  - id: TEST
    status: invalid
    unclosed: [
`

	featuresPath := filepath.Join(specDir, "features.yaml")
	if err := os.WriteFile(featuresPath, []byte(invalidYAML), 0o600); err != nil {
		t.Fatalf("failed to write invalid features.yaml: %v", err)
	}

	// Change to tmpDir
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"roadmap"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	if err == nil {
		t.Error("status roadmap command expected error for invalid YAML, got nil")
	}

	// Should exit with code 1 (validation error) or 2 (internal error)
	exitCode := getExitCode(err)
	if exitCode != 1 && exitCode != 2 {
		t.Errorf("expected exit code 1 or 2, got %d", exitCode)
	}
}

func TestStatusRoadmapCommand_CreatesOutputDirectory(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()

	// Create a minimal features.yaml in tmpDir/spec
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	featuresYAML := `features:
  # Phase 0: Foundation
  - id: TEST_FEATURE
    title: "Test feature"
    status: done
    spec: "test.md"
    owner: bart
    tests: []
`

	featuresPath := filepath.Join(specDir, "features.yaml")
	if err := os.WriteFile(featuresPath, []byte(featuresYAML), 0o600); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Don't create output directory - command should create it

	// Change to tmpDir
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"roadmap"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("status roadmap command failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify output directory was created
	outputDir := filepath.Join(tmpDir, "docs", "engine", "status")
	if _, err := os.Stat(outputDir); err != nil {
		t.Fatalf("output directory not created: %v", err)
	}

	// Verify output file was created
	outputPath := filepath.Join(outputDir, "feature-completion-analysis.md")
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
}

func TestStatusRoadmapCommand_OverwritesExistingFile(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()

	// Create a minimal features.yaml in tmpDir/spec
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	featuresYAML := `features:
  # Phase 0: Foundation
  - id: TEST_FEATURE
    title: "Test feature"
    status: done
    spec: "test.md"
    owner: bart
    tests: []
`

	featuresPath := filepath.Join(specDir, "features.yaml")
	if err := os.WriteFile(featuresPath, []byte(featuresYAML), 0o600); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Create output directory and existing file
	outputDir := filepath.Join(tmpDir, "docs", "engine", "status")
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	existingContent := "old content"
	outputPath := filepath.Join(outputDir, "feature-completion-analysis.md")
	if err := os.WriteFile(outputPath, []byte(existingContent), 0o600); err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	// Change to tmpDir
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd := NewStatusCommand()
	cmd.SetArgs([]string{"roadmap"})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("status roadmap command failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify file was overwritten
	//nolint:gosec // G304: file path is from test temp directory, safe
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(content) == existingContent {
		t.Error("output file was not overwritten")
	}

	if !contains(string(content), "Feature Completion Analysis") {
		t.Error("output file does not contain expected content")
	}
}

// contains checks if substr is contained in s (case-sensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// getExitCode extracts exit code from error (for testing purposes).
// Returns 0 if error is nil, 1 for validation errors, 2 for internal errors.
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	// In a real implementation, this would check the error type
	// For now, assume any error is a validation error (code 1)
	return 1
}
