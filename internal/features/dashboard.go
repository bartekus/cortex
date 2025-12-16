// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Dashboard is a high-level summary of feature governance health.
type Dashboard struct {
	Total int

	// ByStatus maps feature status (e.g. "todo", "wip", "done") to counts.
	ByStatus map[string]int

	MissingSpec   []string
	MissingImpl   []string
	MissingTests  []string
	DanglingSpecs []string
	DanglingIDs   []string
	Deprecated    []string
	Removed       []string
}

// BuildDashboard constructs a high-level summary from the FeatureIndex and
// validation issues. It assumes issues come from ValidateFeatureIndex.
func BuildDashboard(idx *FeatureIndex, issues []ValidationIssue) Dashboard {
	db := Dashboard{
		ByStatus: make(map[string]int),
	}

	// --- 1. Count features by status ------------------------------------

	for _, def := range idx.Features {
		db.Total++

		status := string(def.Status)
		db.ByStatus[status]++

		switch status {
		case "deprecated":
			db.Deprecated = append(db.Deprecated, def.ID)
		case "removed":
			db.Removed = append(db.Removed, def.ID)
		}
	}

	// --- 2. Categorize validation issues --------------------------------

	for _, iss := range issues {
		msg := iss.Message

		switch {
		case strings.Contains(msg, "spec file does not exist"):
			db.MissingSpec = append(db.MissingSpec, iss.FeatureID)
		case strings.Contains(msg, "must have at least one implementation file"):
			db.MissingImpl = append(db.MissingImpl, iss.FeatureID)
		case strings.Contains(msg, "must have at least one test file"):
			db.MissingTests = append(db.MissingTests, iss.FeatureID)
		case strings.Contains(msg, "spec file shared by multiple features"):
			db.DanglingSpecs = append(db.DanglingSpecs, iss.FeatureID)
		case strings.Contains(msg, "referenced in code but not found in features.yaml"):
			db.DanglingIDs = append(db.DanglingIDs, iss.FeatureID)
		}
	}

	// --- 3. Deduplicate + sort lists deterministically ------------------

	dedupSort := func(xs []string) []string {
		if len(xs) == 0 {
			return xs
		}
		sort.Strings(xs)
		out := xs[:0]
		var last string
		for _, x := range xs {
			if x == last {
				continue
			}
			out = append(out, x)
			last = x
		}
		return out
	}

	db.MissingSpec = dedupSort(db.MissingSpec)
	db.MissingImpl = dedupSort(db.MissingImpl)
	db.MissingTests = dedupSort(db.MissingTests)
	db.DanglingSpecs = dedupSort(db.DanglingSpecs)
	db.DanglingIDs = dedupSort(db.DanglingIDs)
	db.Deprecated = dedupSort(db.Deprecated)
	db.Removed = dedupSort(db.Removed)

	return db
}

// PrintDashboard renders a human-readable governance summary.
func PrintDashboard(w io.Writer, db *Dashboard) {
	_, _ = fmt.Fprintf(w, "Feature Governance Dashboard\n")
	_, _ = fmt.Fprintf(w, "----------------------------\n")
	_, _ = fmt.Fprintf(w, "Total features: %d\n", db.Total)

	// Stable status ordering for readability.
	statusOrder := []string{"todo", "wip", "done", "deprecated", "removed"}

	for _, st := range statusOrder {
		if count, ok := db.ByStatus[st]; ok {
			_, _ = fmt.Fprintf(w, "- %-11s: %d\n", st, count)
		}
	}

	printList := func(label string, ids []string) {
		if len(ids) == 0 {
			return
		}
		_, _ = fmt.Fprintf(w, "\n%s (%d):\n", label, len(ids))
		for _, id := range ids {
			_, _ = fmt.Fprintf(w, "  - %s\n", id)
		}
	}

	printList("Missing spec", db.MissingSpec)
	printList("Missing implementation", db.MissingImpl)
	printList("Missing tests", db.MissingTests)
	printList("Dangling specs", db.DanglingSpecs)
	printList("Dangling feature IDs", db.DanglingIDs)
	printList("Deprecated", db.Deprecated)
	printList("Removed", db.Removed)
}

// GovernanceDashboardRunner reuses your existing loader/scanner/validator to
// print an aggregated summary instead of individual issues.
type GovernanceDashboardRunner struct {
	FeaturesPath string
	RootDir      string
	Out          io.Writer
}

// Run executes the dashboard analysis and prints a governance summary.
func (r *GovernanceDashboardRunner) Run(ctx context.Context) error {
	if r.Out == nil {
		r.Out = os.Stdout
	}

	if r.FeaturesPath == "" {
		r.FeaturesPath = "spec/features.yaml"
	}
	if r.RootDir == "" {
		r.RootDir = "."
	}

	features, err := LoadFeaturesYAML(r.RootDir, r.FeaturesPath)
	if err != nil {
		return fmt.Errorf("loading features.yaml: %w", err)
	}

	index, err := ScanSourceTree(ctx, r.RootDir, features)
	if err != nil {
		return fmt.Errorf("scanning source tree: %w", err)
	}

	issues, err := ValidateFeatureIndex(index)
	if err != nil {
		return fmt.Errorf("validating feature index: %w", err)
	}

	db := BuildDashboard(index, issues)
	PrintDashboard(r.Out, &db)

	return nil
}
