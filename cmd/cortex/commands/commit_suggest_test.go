// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

// Feature: CLI_COMMAND_COMMIT
// Spec: spec/cli/commit.md

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateCommitSuggestGoldens = flag.Bool("update-commit-suggest-goldens", false, "update commit suggest golden files")

// TestCommitSuggest_JSONGolden verifies that the JSON output of
// `cortex commit suggest --format=json` is deterministic and matches the
// golden file.
//
// This is an end-to-end CLI test that exercises:
//   - Reading existing report files from .cortex/reports
//   - Generating suggestions via the suggestions package
//   - Applying prioritization and filtering
//   - Rendering the JSON report
func TestCommitSuggest_JSONGolden(t *testing.T) {
	// NOTE: This test MUST NOT use t.Parallel() because it relies on os.Chdir,
	// which is a global process-wide setting and can interfere with other tests.

	output := runCommitSuggestCLI(t, "json")

	goldenPath := filepath.Join("testdata", "commit_suggest_json.golden")

	if *updateCommitSuggestGoldens {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o750); err != nil { //nolint:gosec // G301: test directory
			t.Fatalf("creating testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(output), 0o600); err != nil { //nolint:gosec // G306: test file
			t.Fatalf("writing golden file: %v", err)
		}
		return
	}

	want, err := os.ReadFile(goldenPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("reading golden file %s: %v", goldenPath, err)
	}

	if output != string(want) {
		t.Fatalf("JSON output does not match golden file.\nGolden path: %s", goldenPath)
	}
}

// TestCommitSuggest_TextGolden verifies that the text output of
// `cortex commit suggest --format=text` is deterministic and matches the
// golden file.
//
// This is an end-to-end CLI test that exercises the text formatter.
func TestCommitSuggest_TextGolden(t *testing.T) {
	// NOTE: This test MUST NOT use t.Parallel() because it relies on os.Chdir,
	// which is a global process-wide setting and can interfere with other tests.

	output := runCommitSuggestCLI(t, "text")

	goldenPath := filepath.Join("testdata", "commit_suggest_text.golden")

	if *updateCommitSuggestGoldens {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o750); err != nil { //nolint:gosec // G301: test directory
			t.Fatalf("creating testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(output), 0o600); err != nil { //nolint:gosec // G306: test file
			t.Fatalf("writing golden file: %v", err)
		}
		return
	}

	want, err := os.ReadFile(goldenPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("reading golden file %s: %v", goldenPath, err)
	}

	if output != string(want) {
		t.Fatalf("text output does not match golden file.\nGolden path: %s", goldenPath)
	}
}

// runCommitSuggestCLI sets up a minimal deterministic repo with pre-generated
// reports, then runs the commit suggest command with the given format and
// returns stdout as a string.
//
// The fixture reports are intentionally minimal:
//
//   - commit-health.json: empty commit set (no violations)
//   - feature-traceability.json: empty feature set
//
// This ensures deterministic behaviour while still exercising the complete
// CLI wiring and rendering logic. More complex fixtures can be introduced
// later if needed.
func runCommitSuggestCLI(t *testing.T, format string) string {
	t.Helper()

	// Create an isolated temporary repo directory.
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "cortex")
	if err := os.Mkdir(repoDir, 0o750); err != nil {
		t.Fatalf("creating repo dir: %v", err)
	}

	// Prepare .cortex/reports with minimal, valid JSON fixtures.
	reportsDir := filepath.Join(repoDir, ".cortex", "reports")
	if err := os.MkdirAll(reportsDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("creating reports dir: %v", err)
	}

	commitHealthPath := filepath.Join(reportsDir, "commit-health.json")
	featureTracePath := filepath.Join(reportsDir, "feature-traceability.json")

	// Minimal valid commit health report (no commits, no violations).
	const commitHealthFixture = `{
  "schema_version": "1.0",
  "repo": {
    "name": "cortex",
    "default_branch": "main"
  },
  "range": {
    "from": "origin/main",
    "to": "HEAD",
    "description": "origin/main..HEAD"
  },
  "summary": {
    "total_commits": 0,
    "valid_commits": 0,
    "invalid_commits": 0,
    "violations_by_code": {}
  },
  "rules": [],
  "commits": {}
}
`

	// Create go.mod marker
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module test"), 0o600); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	// Minimal valid feature traceability report (no features).
	const featureTraceFixture = `{
  "schema_version": "1.0",
  "summary": {
    "total_features": 0,
    "done": 0,
    "wip": 0,
    "todo": 0,
    "deprecated": 0,
    "removed": 0,
    "features_with_gaps": 0
  },
  "features": {}
}
`

	if err := os.WriteFile(commitHealthPath, []byte(commitHealthFixture), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("writing commit-health fixture: %v", err)
	}

	if err := os.WriteFile(featureTracePath, []byte(featureTraceFixture), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("writing feature-traceability fixture: %v", err)
	}

	// Change working directory to the temp repo so the command discovers
	// .cortex/reports relative to CWD.
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working directory: %v", err)
	}

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("changing working directory: %v", err)
	}

	defer func() {
		_ = os.Chdir(origWD)
	}()

	// Construct the command and capture output.
	cmd := NewCommitSuggestCommand()
	cmd.SetArgs([]string{
		"--format=" + format,
		"--severity=info",
		"--max-suggestions=0",
	})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("running commit suggest command (format=%s): %v", format, err)
	}

	return buf.String()
}
