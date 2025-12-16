// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: GOV_CORE
// Docs: docs/design/commit-reports-go-types.md
package featuretrace

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestGenerateFeatureTraceabilityReport_CompleteFeature(t *testing.T) {
	t.Parallel()

	features := []FeaturePresence{
		{
			FeatureID:           "CLI_DEPLOY",
			Status:              FeatureStatusDone,
			HasSpec:             true,
			SpecPath:            "spec/commands/deploy.md",
			ImplementationFiles: []string{"cmd/deploy.go", "internal/core/deploy/deploy.go"},
			TestFiles:           []string{"cmd/deploy_test.go", "internal/core/deploy/deploy_test.go"},
			CommitSHAs:          []string{"abc123", "def456"},
		},
	}

	report, err := GenerateFeatureTraceabilityReport(features)
	if err != nil {
		t.Fatalf("GenerateFeatureTraceabilityReport failed: %v", err)
	}

	if report.Summary.TotalFeatures != 1 {
		t.Errorf("expected TotalFeatures=1, got %d", report.Summary.TotalFeatures)
	}
	if report.Summary.Done != 1 {
		t.Errorf("expected Done=1, got %d", report.Summary.Done)
	}
	if report.Summary.FeaturesWithGaps != 0 {
		t.Errorf("expected FeaturesWithGaps=0, got %d", report.Summary.FeaturesWithGaps)
	}

	feature, ok := report.Features["CLI_DEPLOY"]
	if !ok {
		t.Fatal("feature CLI_DEPLOY not found in report")
	}

	if feature.Status != FeatureStatusDone {
		t.Errorf("expected status=done, got %s", feature.Status)
	}
	if !feature.Spec.Present {
		t.Error("expected spec to be present")
	}
	if feature.Spec.Path != "spec/commands/deploy.md" {
		t.Errorf("expected spec path=spec/commands/deploy.md, got %s", feature.Spec.Path)
	}
	if !feature.Implementation.Present {
		t.Error("expected implementation to be present")
	}
	if len(feature.Implementation.Files) != 2 {
		t.Errorf("expected 2 implementation files, got %d", len(feature.Implementation.Files))
	}
	if !feature.Tests.Present {
		t.Error("expected tests to be present")
	}
	if !feature.Commits.Present {
		t.Error("expected commits to be present")
	}
	if len(feature.Problems) != 0 {
		t.Errorf("expected no problems, got %d", len(feature.Problems))
	}
}

func TestGenerateFeatureTraceabilityReport_FeatureWithGaps(t *testing.T) {
	t.Parallel()

	features := []FeaturePresence{
		{
			FeatureID:           "CLI_VALIDATE_COMMIT",
			Status:              FeatureStatusTodo,
			HasSpec:             true,
			SpecPath:            "spec/commands/validate-commit.md",
			ImplementationFiles: []string{},
			TestFiles:           []string{},
			CommitSHAs:          []string{},
		},
	}

	report, err := GenerateFeatureTraceabilityReport(features)
	if err != nil {
		t.Fatalf("GenerateFeatureTraceabilityReport failed: %v", err)
	}

	if report.Summary.TotalFeatures != 1 {
		t.Errorf("expected TotalFeatures=1, got %d", report.Summary.TotalFeatures)
	}
	if report.Summary.Todo != 1 {
		t.Errorf("expected Todo=1, got %d", report.Summary.Todo)
	}
	if report.Summary.FeaturesWithGaps != 1 {
		t.Errorf("expected FeaturesWithGaps=1, got %d", report.Summary.FeaturesWithGaps)
	}

	feature, ok := report.Features["CLI_VALIDATE_COMMIT"]
	if !ok {
		t.Fatal("feature CLI_VALIDATE_COMMIT not found in report")
	}

	if !feature.Spec.Present {
		t.Error("expected spec to be present")
	}
	if feature.Implementation.Present {
		t.Error("expected implementation to be absent")
	}
	if feature.Tests.Present {
		t.Error("expected tests to be absent")
	}
	if feature.Commits.Present {
		t.Error("expected commits to be absent")
	}

	// Check for problems
	if len(feature.Problems) == 0 {
		t.Error("expected problems to be detected")
	}

	foundMissingImpl := false
	foundMissingTests := false
	foundMissingCommits := false
	for _, p := range feature.Problems {
		if p.Code == ProblemCodeMissingImplementation {
			foundMissingImpl = true
		}
		if p.Code == ProblemCodeMissingTests {
			foundMissingTests = true
		}
		if p.Code == ProblemCodeMissingCommits {
			foundMissingCommits = true
		}
	}

	if !foundMissingImpl {
		t.Error("expected MISSING_IMPLEMENTATION problem")
	}
	if !foundMissingTests {
		t.Error("expected MISSING_TESTS problem")
	}
	if !foundMissingCommits {
		t.Error("expected MISSING_COMMITS problem")
	}
}

func TestGenerateFeatureTraceabilityReport_MixedFeatures(t *testing.T) {
	t.Parallel()

	features := []FeaturePresence{
		{
			FeatureID:           "CLI_DEPLOY",
			Status:              FeatureStatusDone,
			HasSpec:             true,
			SpecPath:            "spec/commands/deploy.md",
			ImplementationFiles: []string{"cmd/deploy.go"},
			TestFiles:           []string{"cmd/deploy_test.go"},
			CommitSHAs:          []string{"abc123"},
		},
		{
			FeatureID:           "CLI_VALIDATE_COMMIT",
			Status:              FeatureStatusTodo,
			HasSpec:             true,
			SpecPath:            "spec/commands/validate-commit.md",
			ImplementationFiles: []string{},
			TestFiles:           []string{},
			CommitSHAs:          []string{},
		},
	}

	report, err := GenerateFeatureTraceabilityReport(features)
	if err != nil {
		t.Fatalf("GenerateFeatureTraceabilityReport failed: %v", err)
	}

	if report.Summary.TotalFeatures != 2 {
		t.Errorf("expected TotalFeatures=2, got %d", report.Summary.TotalFeatures)
	}
	if report.Summary.Done != 1 {
		t.Errorf("expected Done=1, got %d", report.Summary.Done)
	}
	if report.Summary.Todo != 1 {
		t.Errorf("expected Todo=1, got %d", report.Summary.Todo)
	}
	if report.Summary.FeaturesWithGaps != 1 {
		t.Errorf("expected FeaturesWithGaps=1, got %d", report.Summary.FeaturesWithGaps)
	}
}

func TestGenerateFeatureTraceabilityReport_EmptyFeatures(t *testing.T) {
	t.Parallel()

	features := []FeaturePresence{}

	report, err := GenerateFeatureTraceabilityReport(features)
	if err != nil {
		t.Fatalf("GenerateFeatureTraceabilityReport failed: %v", err)
	}

	if report.Summary.TotalFeatures != 0 {
		t.Errorf("expected TotalFeatures=0, got %d", report.Summary.TotalFeatures)
	}
	if len(report.Features) != 0 {
		t.Errorf("expected 0 features in report, got %d", len(report.Features))
	}
}

func TestGenerateFeatureTraceabilityReport_JSONMatchesGolden(t *testing.T) {
	t.Parallel()

	features := []FeaturePresence{
		{
			FeatureID:           "CLI_DEPLOY",
			Status:              FeatureStatusDone,
			HasSpec:             true,
			SpecPath:            "spec/commands/deploy.md",
			ImplementationFiles: []string{"cmd/deploy.go", "internal/core/deploy/deploy.go"},
			TestFiles:           []string{"cmd/deploy_test.go", "internal/core/deploy/deploy_test.go"},
			CommitSHAs:          []string{"abc123", "def456"},
		},
		{
			FeatureID:           "CLI_VALIDATE_COMMIT",
			Status:              FeatureStatusTodo,
			HasSpec:             true,
			SpecPath:            "spec/commands/validate-commit.md",
			ImplementationFiles: []string{},
			TestFiles:           []string{},
			CommitSHAs:          []string{},
		},
	}

	report, err := GenerateFeatureTraceabilityReport(features)
	if err != nil {
		t.Fatalf("GenerateFeatureTraceabilityReport failed: %v", err)
	}

	got := marshalCompactJSON(t, report)

	goldenPath := filepath.Join("testdata", "feature-traceability_report.golden.json")
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
