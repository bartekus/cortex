// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CLI_COMMAND_COMMIT
// Spec: spec/cli/commit.md
package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartekus/cortex/internal/reports"
	"github.com/bartekus/cortex/internal/reports/commithealth"
)

func TestRunCommitReport_GeneratesReport(t *testing.T) {
	t.Parallel()

	// Create temporary repository structure
	tmpDir := t.TempDir()

	// Create .cortex/reports directory
	reportsDir := filepath.Join(tmpDir, ".cortex", "reports")
	if err := os.MkdirAll(reportsDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create reports directory: %v", err)
	}

	// Create minimal spec/features.yaml
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create spec directory: %v", err)
	}

	featuresYAML := `features:
  - id: CLI_DEPLOY
    title: "Deploy command"
    status: done
`
	if err := os.WriteFile(filepath.Join(specDir, "features.yaml"), []byte(featuresYAML), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Synthesize a small commit history in memory
	commits := []commithealth.CommitMetadata{
		{
			SHA:     "abc123",
			Message: "feat(CLI_DEPLOY): add deploy support",
		},
	}

	knownFeatures := map[string]bool{
		"CLI_DEPLOY": true,
	}

	repoInfo := commithealth.RepoInfo{
		Name:          "cortex",
		DefaultBranch: "main",
	}

	rangeInfo := commithealth.CommitRange{
		From:        "origin/main",
		To:          "HEAD",
		Description: "origin/main..HEAD",
	}

	report, err := commithealth.GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		t.Fatalf("GenerateCommitHealthReport failed: %v", err)
	}

	reportPath := filepath.Join(reportsDir, "commit-health.json")
	if err := reports.WriteJSONAtomic(reportPath, report); err != nil {
		t.Fatalf("WriteJSONAtomic failed: %v", err)
	}

	info, err := os.Stat(reportPath)
	if err != nil {
		t.Fatalf("failed to stat report file: %v", err)
	}

	if info.Size() == 0 {
		t.Fatalf("expected non-empty commit-health.json report")
	}
}

func TestRunCommitReport_ReportStructure(t *testing.T) {
	t.Parallel()

	// Test that the generated report has the expected structure
	// This is a unit test that doesn't require actual git

	commits := []commithealth.CommitMetadata{
		{
			SHA:     "abc123",
			Message: "feat(CLI_DEPLOY): add rollback support",
		},
	}

	knownFeatures := map[string]bool{
		"CLI_DEPLOY": true,
	}

	repoInfo := commithealth.RepoInfo{
		Name:          "cortex",
		DefaultBranch: "main",
	}

	rangeInfo := commithealth.CommitRange{
		From:        "origin/main",
		To:          "HEAD",
		Description: "origin/main..HEAD",
	}

	report, err := commithealth.GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		t.Fatalf("GenerateCommitHealthReport failed: %v", err)
	}

	// Verify report structure
	if report.SchemaVersion != "1.0" {
		t.Errorf("expected schema_version=1.0, got %s", report.SchemaVersion)
	}

	if report.Repo.Name != "cortex" {
		t.Errorf("expected repo.name=cortex, got %s", report.Repo.Name)
	}

	// Verify report can be marshaled to JSON
	_, err = json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report to JSON: %v", err)
	}
}

func TestLoadFeatureRegistry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create spec/features.yaml
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create spec directory: %v", err)
	}

	featuresYAML := `features:
  - id: CLI_DEPLOY
    title: "Deploy command"
    status: done
  - id: CLI_PLAN
    title: "Plan command"
    status: done
`
	if err := os.WriteFile(filepath.Join(specDir, "features.yaml"), []byte(featuresYAML), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	registry, err := loadFeatureRegistry(tmpDir)
	if err != nil {
		t.Fatalf("loadFeatureRegistry failed: %v", err)
	}

	// Verify registry contains expected features
	if !registry["CLI_DEPLOY"] {
		t.Error("CLI_DEPLOY not found in registry")
	}
	if !registry["CLI_PLAN"] {
		t.Error("CLI_PLAN not found in registry")
	}
}

func TestRunCommitReport_CLIEndToEnd(t *testing.T) {
	// Not parallel: uses os.Chdir which is global process state

	// Override history source with a deterministic fake
	oldNewHistorySource := newHistorySource
	defer func() {
		newHistorySource = oldNewHistorySource
	}()

	fakeCommits := []commithealth.CommitMetadata{
		{
			SHA:     "abc123",
			Message: "feat(CLI_DEPLOY): add deploy support",
		},
		{
			SHA:     "def456",
			Message: "chore: update docs",
		},
	}

	newHistorySource = func(rootDir string) commithealth.HistorySource {
		return fakeHistorySource{commits: fakeCommits}
	}

	// Create temporary repository structure
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "cortex")
	if err := os.Mkdir(repoDir, 0o750); err != nil {
		t.Fatalf("creating repo dir: %v", err)
	}

	// Create .cortex/reports directory (expected output location)
	reportsDir := filepath.Join(repoDir, ".cortex", "reports")
	if err := os.MkdirAll(reportsDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create reports directory: %v", err)
	}

	// Create minimal spec/features.yaml so the CLI can load known features
	specDir := filepath.Join(repoDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create spec directory: %v", err)
	}

	featuresYAML := `features:
  - id: CLI_DEPLOY
    title: "Deploy command"
    status: done
`
	if err := os.WriteFile(filepath.Join(specDir, "features.yaml"), []byte(featuresYAML), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Change to temp directory so the CLI uses it as repo root
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	// Use repoDir instead of tmpDir
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Run the real CLI command end-to-end
	// Note: We rely on WriteJSONAtomic to create the reports directory
	// The spec/features.yaml was created before chdir, so it should exist
	cmd := NewCommitReportCommand()

	// Re-verify we're in the correct directory right before CLI call
	// This minimizes race conditions with parallel tests that also use chdir
	currentWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if currentWd != repoDir {
		if err := os.Chdir(repoDir); err != nil {
			t.Fatalf("directory changed by parallel test, failed to restore: %v", err)
		}
	}

	if err := runCommitReport(cmd, nil); err != nil {
		t.Fatalf("runCommitReport failed: %v", err)
	}

	// Verify the report file was created
	reportPath := filepath.Join(repoDir, ".cortex", "reports", "commit-health.json")
	data, err := os.ReadFile(reportPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("failed to read report file: %v", err)
	}

	// Unmarshal into the concrete report type to ensure the JSON matches the schema
	var report commithealth.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("failed to unmarshal report JSON: %v", err)
	}

	// Basic structural assertions for determinism and correctness
	if report.SchemaVersion != "1.0" {
		t.Errorf("expected schema_version=1.0, got %s", report.SchemaVersion)
	}

	if report.Repo.Name != "cortex" {
		t.Errorf("expected repo.name=cortex, got %s", report.Repo.Name)
	}
}

type fakeHistorySource struct {
	commits []commithealth.CommitMetadata
}

func (f fakeHistorySource) Commits() ([]commithealth.CommitMetadata, error) {
	return f.commits, nil
}

func TestRunCommitReport_GoldenFile(t *testing.T) {
	t.Parallel()

	commits := []commithealth.CommitMetadata{
		{
			SHA:     "abc123",
			Message: "feat(CLI_DEPLOY): add deploy support",
		},
		{
			SHA:     "def456",
			Message: "chore: update docs",
		},
	}

	knownFeatures := map[string]bool{
		"CLI_DEPLOY": true,
	}

	repoInfo := commithealth.RepoInfo{
		Name:          "cortex",
		DefaultBranch: "main",
	}

	rangeInfo := commithealth.CommitRange{
		From:        "origin/main",
		To:          "HEAD",
		Description: "origin/main..HEAD",
	}

	report, err := commithealth.GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		t.Fatalf("GenerateCommitHealthReport failed: %v", err)
	}

	formatted, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal report: %v", err)
	}
	formatted = append(formatted, '\n')

	expected := readGoldenFile(t, "commit_report")

	if *updateGolden {
		writeGoldenFile(t, "commit_report", string(formatted))
		expected = string(formatted)
	}

	if string(formatted) != expected {
		t.Errorf("report mismatch:\nGot:\n%s\nExpected:\n%s", string(formatted), expected)
	}
}
