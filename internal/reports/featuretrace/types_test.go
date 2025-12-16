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
	"os"
	"path/filepath"
	"testing"
)

func TestReportJSONMatchesGolden(t *testing.T) {
	t.Parallel()

	report := Report{
		SchemaVersion: "1.0",
		Summary: Summary{
			TotalFeatures:    2,
			Done:             1,
			WIP:              0,
			Todo:             1,
			Deprecated:       0,
			Removed:          0,
			FeaturesWithGaps: 1,
		},
		Features: map[string]Feature{
			"CLI_DEPLOY": {
				Status: FeatureStatusDone,
				Spec: SpecInfo{
					Present: true,
					Path:    "spec/commands/deploy.md",
				},
				Implementation: ImplementationInfo{
					Present: true,
					Files: []string{
						"cmd/deploy.go",
						"internal/core/deploy/deploy.go",
					},
				},
				Tests: TestsInfo{
					Present: true,
					Files: []string{
						"cmd/deploy_test.go",
						"internal/core/deploy/deploy_test.go",
					},
				},
				Commits: CommitsInfo{
					Present: true,
					SHAs: []string{
						"abc123",
						"def456",
					},
				},
				Problems: nil,
			},
			"CLI_VALIDATE_COMMIT": {
				Status: FeatureStatusTodo,
				Spec: SpecInfo{
					Present: true,
					Path:    "spec/commands/validate-commit.md",
				},
				Implementation: ImplementationInfo{
					Present: false,
					Files:   nil,
				},
				Tests: TestsInfo{
					Present: false,
					Files:   nil,
				},
				Commits: CommitsInfo{
					Present: false,
					SHAs:    nil,
				},
				Problems: []Problem{
					{
						Code:     ProblemCodeMissingImplementation,
						Severity: SeverityWarning,
						Message:  "Feature has a spec but no implementation files.",
						Details:  map[string]any{},
					},
					{
						Code:     ProblemCodeMissingTests,
						Severity: SeverityWarning,
						Message:  "Feature has a spec but no tests.",
						Details:  map[string]any{},
					},
					{
						Code:     ProblemCodeMissingCommits,
						Severity: SeverityInfo,
						Message:  "Feature has no commits referencing this Feature ID yet.",
						Details:  map[string]any{},
					},
				},
			},
		},
	}

	got := marshalCompactJSON(t, report)

	goldenPath := filepath.Join("testdata", "report_basic.golden.json")
	want := readFile(t, goldenPath)

	if !bytes.Equal(got, want) {
		t.Fatalf("JSON output does not match golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}

	// Roundtrip check.
	var roundtrip Report
	if err := json.Unmarshal(got, &roundtrip); err != nil {
		t.Fatalf("failed to unmarshal JSON back into Report: %v", err)
	}
}

func marshalCompactJSON(t *testing.T, v any) []byte {
	t.Helper()

	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		t.Fatalf("json.Compact failed: %v", err)
	}

	return buf.Bytes()
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // G304: golden file path is derived from test directory
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	return bytes.TrimSpace(data)
}
