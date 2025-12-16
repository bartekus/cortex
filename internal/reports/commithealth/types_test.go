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
	"os"
	"path/filepath"
	"testing"
)

func TestReportJSONMatchesGolden(t *testing.T) {
	t.Parallel()

	report := Report{
		SchemaVersion: "1.0",
		Repo: RepoInfo{
			Name:          "stagecraft",
			DefaultBranch: "main",
		},
		Range: CommitRange{
			From:        "origin/main",
			To:          "HEAD",
			Description: "origin/main..HEAD",
		},
		Summary: Summary{
			TotalCommits:   2,
			ValidCommits:   1,
			InvalidCommits: 1,
			ViolationsByCode: map[ViolationCode]int{
				ViolationCodeMissingFeatureID:   1,
				ViolationCodeMultipleFeatureIDs: 1,
			},
		},
		Rules: []Rule{
			{
				Code:        ViolationCodeMissingFeatureID,
				Description: "Commit message is missing a Feature ID in the required format.",
				Severity:    SeverityError,
			},
			{
				Code:        ViolationCodeMultipleFeatureIDs,
				Description: "Commit message references multiple Feature IDs; only one is allowed per commit.",
				Severity:    SeverityError,
			},
		},
		Commits: map[string]Commit{
			"abc123": {
				Subject:    "feat(CLI_DEPLOY): add rollback support",
				IsValid:    true,
				Violations: nil,
			},
			"def456": {
				Subject: "feat(CLI_PLAN, CLI_DEPLOY): refactor planning and deployment",
				IsValid: false,
				Violations: []Violation{
					{
						Code:     ViolationCodeMultipleFeatureIDs,
						Severity: SeverityError,
						Message:  "Commit message must reference exactly one Feature ID.",
						Details: map[string]any{
							"feature_ids": []string{"CLI_PLAN", "CLI_DEPLOY"},
						},
					},
					{
						Code:     ViolationCodeMissingFeatureID,
						Severity: SeverityError,
						Message:  "Commit message must include a Feature ID in <type>(<FEATURE_ID>): <summary>.",
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

	// Roundtrip check: ensure we can unmarshal back into Report.
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
