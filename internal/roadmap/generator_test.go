// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: GOV_STATUS_ROADMAP
// Spec: spec/commands/status-roadmap.md

package roadmap

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateGolden is a flag to update golden files during development.
// Usage: go test -update ./internal/tools/roadmap
var updateGolden = flag.Bool("update", false, "update golden files")

func TestGenerateMarkdown_GoldenTest(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	stats := CalculateStats(phases)
	blockers := IdentifyBlockers(phases)

	markdown := GenerateMarkdown(stats, blockers)

	goldenPath := filepath.Join(testDataDir, "feature-completion-analysis.md.golden")

	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(markdown), 0o600); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	//nolint:gosec // G304: file path is from testdata directory, safe
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	got := strings.TrimSpace(markdown)
	wantStr := strings.TrimSpace(string(want))

	if got != wantStr {
		t.Errorf("GenerateMarkdown() output does not match golden file")
		t.Logf("Diff:\n%s", diffStrings(wantStr, got))
	}
}

func TestGenerateMarkdown_DeterministicOutput(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	stats := CalculateStats(phases)
	blockers := IdentifyBlockers(phases)

	// Generate markdown twice
	markdown1 := GenerateMarkdown(stats, blockers)
	markdown2 := GenerateMarkdown(stats, blockers)

	if markdown1 != markdown2 {
		t.Error("GenerateMarkdown() output is not deterministic")
		t.Logf("First run length: %d", len(markdown1))
		t.Logf("Second run length: %d", len(markdown2))
	}
}

func TestGenerateMarkdown_IncludesAllSections(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	stats := CalculateStats(phases)
	blockers := IdentifyBlockers(phases)

	markdown := GenerateMarkdown(stats, blockers)

	requiredSections := []string{
		"# Feature Completion Analysis",
		"## Executive Summary",
		"## Phase-by-Phase Completion",
		"## Roadmap Alignment",
		"## Priority Recommendations",
		"## Detailed Phase Analysis",
		"## Critical Path Analysis",
		"## Next Steps",
	}

	for _, section := range requiredSections {
		if !strings.Contains(markdown, section) {
			t.Errorf("GenerateMarkdown() missing required section: %s", section)
		}
	}
}

func TestGenerateMarkdown_EmptyPhases(t *testing.T) {
	t.Helper()

	phases := make(map[string]*Phase)
	stats := CalculateStats(phases)
	blockers := IdentifyBlockers(phases)

	markdown := GenerateMarkdown(stats, blockers)

	// Should still generate valid markdown with zero counts
	if !strings.Contains(markdown, "Total Features") {
		t.Error("GenerateMarkdown() with empty phases missing Total Features")
	}

	if !strings.Contains(markdown, "0") {
		t.Error("GenerateMarkdown() with empty phases should show 0 counts")
	}
}

func TestGenerateMarkdown_PhasesSortedCorrectly(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	stats := CalculateStats(phases)
	blockers := IdentifyBlockers(phases)

	markdown := GenerateMarkdown(stats, blockers)

	// Verify phases appear in correct order: Architecture, Phase 0-10, Governance
	// Extract phase order from markdown
	lines := strings.Split(markdown, "\n")
	phaseOrder := []string{}
	inTable := false
	for _, line := range lines {
		if strings.Contains(line, "## Phase-by-Phase Completion") {
			inTable = true
			continue
		}
		if inTable && strings.HasPrefix(line, "| **") {
			// Extract phase name from table row
			parts := strings.Split(line, "|")
			if len(parts) > 1 {
				phaseName := strings.TrimSpace(strings.Trim(parts[1], "*"))
				phaseOrder = append(phaseOrder, phaseName)
			}
		}
		if inTable && strings.HasPrefix(line, "⸻") {
			break
		}
	}

	// Verify Architecture comes before Phase 0
	archIdx := -1
	phase0Idx := -1
	for i, phase := range phaseOrder {
		if strings.Contains(phase, "Architecture") {
			archIdx = i
		}
		if strings.Contains(phase, "Phase 0") {
			phase0Idx = i
		}
	}

	if archIdx >= 0 && phase0Idx >= 0 && archIdx >= phase0Idx {
		t.Error("GenerateMarkdown() phases not sorted correctly: Architecture should come before Phase 0")
	}
}

// diffStrings returns a simple diff representation (for test output only).
func diffStrings(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")

	maxLen := len(wantLines)
	if len(gotLines) > maxLen {
		maxLen = len(gotLines)
	}

	var diff strings.Builder
	for i := 0; i < maxLen; i++ {
		var wantLine, gotLine string
		if i < len(wantLines) {
			wantLine = wantLines[i]
		}
		if i < len(gotLines) {
			gotLine = gotLines[i]
		}

		if wantLine != gotLine {
			fmt.Fprintf(&diff, "Line %d:\n", i+1)
			fmt.Fprintf(&diff, "  want: %q\n", wantLine)
			fmt.Fprintf(&diff, "  got:  %q\n", gotLine)
		}
	}

	return diff.String()
}
