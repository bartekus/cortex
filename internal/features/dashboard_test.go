// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import (
	"testing"
)

func TestBuildDashboard_BasicBuckets(t *testing.T) {
	t.Parallel()

	idx := &FeatureIndex{
		RootDir: ".",
		Features: map[string]*FeatureSpec{
			"F_TODO":       {ID: "F_TODO", Status: FeatureStatusTodo},
			"F_DONE":       {ID: "F_DONE", Status: FeatureStatusDone},
			"F_DEPRECATED": {ID: "F_DEPRECATED", Status: FeatureStatusDeprecated},
		},
		Impls: make(map[string][]FileReference),
		Tests: make(map[string][]FileReference),
	}

	issues := []ValidationIssue{
		{FeatureID: "F_DONE", Message: "done feature spec file does not exist"},
		{FeatureID: "F_DONE", Message: "done feature must have at least one test file"},
		{FeatureID: "F_TODO", Message: "Feature ID \"F_TODO\" referenced in code but not found in features.yaml"},
	}

	db := BuildDashboard(idx, issues)

	if db.Total != 3 {
		t.Fatalf("expected Total=3, got %d", db.Total)
	}

	if db.ByStatus["todo"] != 1 {
		t.Errorf("expected 1 todo feature, got %d", db.ByStatus["todo"])
	}

	if db.ByStatus["done"] != 1 {
		t.Errorf("expected 1 done feature, got %d", db.ByStatus["done"])
	}

	if db.ByStatus["deprecated"] != 1 {
		t.Errorf("expected 1 deprecated feature, got %d", db.ByStatus["deprecated"])
	}

	if len(db.MissingSpec) == 0 {
		t.Errorf("expected MissingSpec to be populated")
	}

	if len(db.MissingTests) == 0 {
		t.Errorf("expected MissingTests to be populated")
	}

	// Check that F_DEPRECATED is in the Deprecated slice
	found := false
	for _, id := range db.Deprecated {
		if id == "F_DEPRECATED" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected F_DEPRECATED to be in Deprecated slice")
	}
}
