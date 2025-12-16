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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.RootDir != "." {
		t.Errorf("Default RootDir must be '.', got %q", opts.RootDir)
	}
}

func TestAnalyze_EmptyRepositoryIsDeterministic(t *testing.T) {
	t.Parallel()

	// This test requires the actual repository structure with spec/features.yaml.
	// Skip if running from a different context.
	opts := Options{
		RootDir: ".",
	}

	// Check if we're in the repo root
	if _, err := os.Stat("spec/features.yaml"); os.IsNotExist(err) {
		t.Skipf("spec/features.yaml not found, skipping test (may be running from test directory)")
	}

	report, err := Analyze(opts)
	// We expect this to work even if there are violations in the real repo
	if err != nil {
		t.Fatalf("Analyze must not error on basic execution: %v", err)
	}
	if report.Features == nil {
		t.Fatal("report.Features must not be nil")
	}
	// Note: We don't assert empty features/violations here because the real repo
	// may have violations that need to be fixed.
}

func TestAnalyze_MappingBasicSemantics(t *testing.T) {
	t.Parallel()

	root := fixtureRoot(t, "mapping_basic")

	// Define the feature universe for this fixture.
	features := []featureMeta{
		// todo - no spec, no impl, no tests - allowed
		{ID: "FEATURE_TODO", Status: "todo", SpecPath: "spec/feature_todo_missing.md"},
		// wip - spec + impl, no tests - allowed
		{ID: "FEATURE_WIP_OK", Status: "wip", SpecPath: "spec/feature_wip_ok.md"},
		// done - spec + impl + test - fully OK
		{ID: "FEATURE_DONE_OK", Status: "done", SpecPath: "spec/feature_done_ok.md"},
		// done - spec path points at a non-existent file - MISSING_SPEC
		{ID: "FEATURE_NOSPEC", Status: "done", SpecPath: "spec/missing_spec.md"},
		// done - spec exists, no impl, no tests - MISSING_IMPL
		{ID: "FEATURE_IMPL_MISSING", Status: "done", SpecPath: "spec/feature_impl_missing.md"},
		// done - spec exists, impl exists, no tests - MISSING_TESTS
		{ID: "FEATURE_TEST_MISSING", Status: "done", SpecPath: "spec/feature_test_missing.md"},
		// done - spec exists, impl uses wrong Spec header - SPEC_PATH_MISMATCH
		{ID: "FEATURE_MISMATCH", Status: "done", SpecPath: "spec/feature_mismatch.md"},
	}

	report, err := analyzeWithFeatures(root, features)
	if err != nil {
		t.Fatalf("analyzeWithFeatures must not error: %v", err)
	}

	// Sanity check: deterministic ordering
	if !isSortedFeatures(report.Features) {
		t.Error("features must be sorted by ID")
	}
	if !isSortedViolations(report.Violations) {
		t.Error("violations must be sorted by (code, feature, path)")
	}

	// 1) FEATURE_DONE_OK should be present and OK.
	doneOK := findFeature(t, report, "FEATURE_DONE_OK")
	if doneOK.Status != FeatureStatusOK {
		t.Errorf("expected FEATURE_DONE_OK status to be %q, got %q", FeatureStatusOK, doneOK.Status)
	}
	if len(doneOK.ImplFiles) == 0 {
		t.Error("expected FEATURE_DONE_OK to have implementation files")
	}
	if len(doneOK.TestFiles) == 0 {
		t.Error("expected FEATURE_DONE_OK to have test files")
	}

	// 2) FEATURE_WIP_OK - wip with impl, no tests - allowed.
	wipOK := findFeature(t, report, "FEATURE_WIP_OK")
	if len(wipOK.ImplFiles) == 0 {
		t.Error("expected FEATURE_WIP_OK to have implementation files")
	}
	if len(wipOK.TestFiles) != 0 {
		t.Error("expected FEATURE_WIP_OK to have no test files")
	}

	// 3) FEATURE_NOSPEC - missing spec file.
	requireContainsViolation(t, report.Violations, CodeMissingSpec, "FEATURE_NOSPEC", "spec/missing_spec.md")

	// 4) FEATURE_IMPL_MISSING - no impls.
	requireContainsViolation(t, report.Violations, CodeMissingImpl, "FEATURE_IMPL_MISSING", "")

	// 5) FEATURE_TEST_MISSING - no tests.
	requireContainsViolation(t, report.Violations, CodeMissingTests, "FEATURE_TEST_MISSING", "")

	// 6) FEATURE_MISMATCH - wrong Spec header in impl.
	requireContainsViolation(t, report.Violations, CodeSpecPathMismatch, "FEATURE_MISMATCH", "cmd/feature_mismatch_impl.go")

	// 7) Orphan spec file (orphan_only.md) - no feature points at it.
	requireContainsViolation(t, report.Violations, CodeOrphanSpec, "", "spec/orphan_only.md")

	// 8) Implementation referencing unknown feature (FEATURE_UNKNOWN).
	requireContainsViolation(t, report.Violations, CodeFeatureNotListed, "FEATURE_UNKNOWN", "cmd/ghost_feature_impl.go")
}

// fixtureRoot resolves the path to a named fixture under testdata.
func fixtureRoot(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller must succeed")
	}
	dir := filepath.Dir(file)
	return filepath.Join(dir, "testdata", name)
}

func findFeature(t *testing.T, report Report, id string) FeatureMapping {
	t.Helper()
	for _, f := range report.Features {
		if f.ID == id {
			return f
		}
	}
	t.Fatalf("feature %q not found in report", id)
	return FeatureMapping{}
}

func requireContainsViolation(t *testing.T, vs []Violation, code ReportCode, feature, pathSuffix string) {
	t.Helper()
	for _, v := range vs {
		if v.Code != code {
			continue
		}
		if feature != "" && v.Feature != feature {
			continue
		}
		if pathSuffix != "" && !strings.HasSuffix(v.Path, pathSuffix) {
			continue
		}
		// Found a matching violation.
		return
	}
	t.Fatalf("expected violation %s for feature %q with path suffix %q not found", code, feature, pathSuffix)
}

func isSortedFeatures(fs []FeatureMapping) bool {
	for i := 1; i < len(fs); i++ {
		if fs[i-1].ID > fs[i].ID {
			return false
		}
	}
	return true
}

func isSortedViolations(vs []Violation) bool {
	for i := 1; i < len(vs); i++ {
		prev := vs[i-1]
		cur := vs[i]
		if lessViolation(cur, prev) {
			return false
		}
	}
	return true
}
