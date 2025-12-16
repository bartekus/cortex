// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: GOV_CORE
// Spec: spec/governance/GOV_CORE.md

package mapping

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReportGolden(t *testing.T) {
	t.Helper()

	// Fixture root for the synthetic repo.
	fixtureRoot := fixtureRoot(t, "golden_repo")

	// Configure options to point at the fixture repo.
	opts := Options{
		RootDir: fixtureRoot,
	}

	report, err := Analyze(opts)
	if err != nil {
		t.Fatalf("Analyze() on golden_repo fixture failed: %v", err)
	}

	// Marshal the report to indented JSON to match the golden file format.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		t.Fatalf("failed to encode report as JSON: %v", err)
	}

	got := buf.Bytes()

	goldenPath := filepath.Join(fixtureRoot, "golden", "feature-mapping-report.json")

	// Allow explicit golden updates when needed.
	if os.Getenv("UPDATE_MAPPING_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil { //nolint:gosec // G301: test directory
			t.Fatalf("failed to create golden directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, got, 0o600); err != nil { //nolint:gosec // G306: test file
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath) //nolint:gosec // G304: path is from test fixture, safe
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	// Normalise trailing newlines for comparison.
	gotTrimmed := bytes.TrimSpace(got)
	wantTrimmed := bytes.TrimSpace(want)

	if !bytes.Equal(gotTrimmed, wantTrimmed) {
		t.Errorf("mapping report does not match golden.\nGolden: %s\nActual:\n%s\n",
			goldenPath, string(got))
	}
}
