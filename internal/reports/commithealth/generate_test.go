// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: GOV_CORE
// Docs: docs/design/commit-reports-go-types.md
package commithealth

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestGenerateCommitHealthReport_ValidCommits(t *testing.T) {
	t.Parallel()

	commits := []CommitMetadata{
		{
			SHA:     "abc123",
			Message: "feat(CLI_DEPLOY): add rollback support",
		},
		{
			SHA:     "def456",
			Message: "fix(CORE_CONFIG): fix config loading bug",
		},
	}

	knownFeatures := map[string]bool{
		"CLI_DEPLOY":  true,
		"CORE_CONFIG": true,
	}

	repoInfo := RepoInfo{
		Name:          "stagecraft",
		DefaultBranch: "main",
	}

	rangeInfo := CommitRange{
		From:        "origin/main",
		To:          "HEAD",
		Description: "origin/main..HEAD",
	}

	report, err := GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		t.Fatalf("GenerateCommitHealthReport failed: %v", err)
	}

	if report.Summary.TotalCommits != 2 {
		t.Errorf("expected TotalCommits=2, got %d", report.Summary.TotalCommits)
	}
	if report.Summary.ValidCommits != 2 {
		t.Errorf("expected ValidCommits=2, got %d", report.Summary.ValidCommits)
	}
	if report.Summary.InvalidCommits != 0 {
		t.Errorf("expected InvalidCommits=0, got %d", report.Summary.InvalidCommits)
	}

	if len(report.Commits) != 2 {
		t.Errorf("expected 2 commits in report, got %d", len(report.Commits))
	}

	commit1, ok := report.Commits["abc123"]
	if !ok {
		t.Fatal("commit abc123 not found in report")
	}
	if !commit1.IsValid {
		t.Error("commit abc123 should be valid")
	}
	if len(commit1.Violations) != 0 {
		t.Errorf("commit abc123 should have no violations, got %d", len(commit1.Violations))
	}
}

func TestGenerateCommitHealthReport_InvalidCommits(t *testing.T) {
	t.Parallel()

	commits := []CommitMetadata{
		{
			SHA:     "abc123",
			Message: "feat(CLI_DEPLOY): add rollback support",
		},
		{
			SHA:     "def456",
			Message: "feat(CLI_PLAN, CLI_DEPLOY): refactor planning and deployment",
		},
		{
			SHA:     "ghi789",
			Message: "fix: address linter errors",
		},
		{
			SHA:     "jkl012",
			Message: "feat(UNKNOWN_FEATURE): add new feature",
		},
	}

	knownFeatures := map[string]bool{
		"CLI_DEPLOY": true,
		"CLI_PLAN":   true,
	}

	repoInfo := RepoInfo{
		Name:          "stagecraft",
		DefaultBranch: "main",
	}

	rangeInfo := CommitRange{
		From:        "origin/main",
		To:          "HEAD",
		Description: "origin/main..HEAD",
	}

	report, err := GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		t.Fatalf("GenerateCommitHealthReport failed: %v", err)
	}

	if report.Summary.TotalCommits != 4 {
		t.Errorf("expected TotalCommits=4, got %d", report.Summary.TotalCommits)
	}
	if report.Summary.ValidCommits != 1 {
		t.Errorf("expected ValidCommits=1, got %d", report.Summary.ValidCommits)
	}
	if report.Summary.InvalidCommits != 3 {
		t.Errorf("expected InvalidCommits=3, got %d", report.Summary.InvalidCommits)
	}

	// Check violations by code
	if report.Summary.ViolationsByCode[ViolationCodeMissingFeatureID] != 1 {
		t.Errorf("expected 1 MISSING_FEATURE_ID violation, got %d", report.Summary.ViolationsByCode[ViolationCodeMissingFeatureID])
	}
	if report.Summary.ViolationsByCode[ViolationCodeMultipleFeatureIDs] != 1 {
		t.Errorf("expected 1 MULTIPLE_FEATURE_IDS violation, got %d", report.Summary.ViolationsByCode[ViolationCodeMultipleFeatureIDs])
	}
	if report.Summary.ViolationsByCode[ViolationCodeFeatureIDNotInSpec] != 1 {
		t.Errorf("expected 1 FEATURE_ID_NOT_IN_SPEC violation, got %d", report.Summary.ViolationsByCode[ViolationCodeFeatureIDNotInSpec])
	}

	// Check commit def456 has multiple feature IDs violation
	commit2, ok := report.Commits["def456"]
	if !ok {
		t.Fatal("commit def456 not found in report")
	}
	if commit2.IsValid {
		t.Error("commit def456 should be invalid")
	}
	foundMultiple := false
	for _, v := range commit2.Violations {
		if v.Code == ViolationCodeMultipleFeatureIDs {
			foundMultiple = true
			break
		}
	}
	if !foundMultiple {
		t.Error("commit def456 should have MULTIPLE_FEATURE_IDS violation")
	}
}

func TestGenerateCommitHealthReport_EmptyHistory(t *testing.T) {
	t.Parallel()

	commits := []CommitMetadata{}

	knownFeatures := map[string]bool{}

	repoInfo := RepoInfo{
		Name:          "stagecraft",
		DefaultBranch: "main",
	}

	rangeInfo := CommitRange{
		From:        "origin/main",
		To:          "HEAD",
		Description: "origin/main..HEAD",
	}

	report, err := GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		t.Fatalf("GenerateCommitHealthReport failed: %v", err)
	}

	if report.Summary.TotalCommits != 0 {
		t.Errorf("expected TotalCommits=0, got %d", report.Summary.TotalCommits)
	}
	if len(report.Commits) != 0 {
		t.Errorf("expected 0 commits in report, got %d", len(report.Commits))
	}
}

func TestGenerateCommitHealthReport_JSONMatchesGolden(t *testing.T) {
	t.Parallel()

	commits := []CommitMetadata{
		{
			SHA:     "abc123",
			Message: "feat(CLI_DEPLOY): add rollback support",
		},
		{
			SHA:     "def456",
			Message: "feat(CLI_PLAN, CLI_DEPLOY): refactor planning and deployment",
		},
	}

	knownFeatures := map[string]bool{
		"CLI_DEPLOY": true,
		"CLI_PLAN":   true,
	}

	repoInfo := RepoInfo{
		Name:          "stagecraft",
		DefaultBranch: "main",
	}

	rangeInfo := CommitRange{
		From:        "origin/main",
		To:          "HEAD",
		Description: "origin/main..HEAD",
	}

	report, err := GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		t.Fatalf("GenerateCommitHealthReport failed: %v", err)
	}

	// Ensure deterministic output by sorting commits map keys
	got := marshalCompactJSON(t, report)

	goldenPath := filepath.Join("testdata", "commit-health_report.golden.json")
	want := readFile(t, goldenPath)

	if !bytes.Equal(got, want) {
		t.Fatalf("JSON output does not match golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}

	// Roundtrip check
	var roundtrip Report
	if err := json.Unmarshal(got, &roundtrip); err != nil {
		t.Fatalf("failed to unmarshal JSON back into Report: %v", err)
	}
}
