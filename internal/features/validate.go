// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ValidationSeverity classifies an issue as a warning or error.
type ValidationSeverity string

// Validation severity constants.
const (
	SeverityWarning ValidationSeverity = "WARNING"
	SeverityError   ValidationSeverity = "ERROR"
)

// ValidationIssue represents a single governance violation or warning.
type ValidationIssue struct {
	Severity  ValidationSeverity
	FeatureID string
	File      string
	Line      int
	Message   string
}

// ValidateFeatureIndex applies Phase 4 rules to the given index and returns a
// sorted list of issues. It is deterministic.
func ValidateFeatureIndex(index *FeatureIndex) ([]ValidationIssue, error) {
	var issues []ValidationIssue

	// Build a map of all Feature IDs referenced in code/tests
	referencedFeatures := make(map[string]bool)
	for id := range index.Impls {
		referencedFeatures[id] = true
	}
	for id := range index.Tests {
		referencedFeatures[id] = true
	}

	// Rule 1: Check for dangling Feature IDs (referenced in code but not in features.yaml)
	for id := range referencedFeatures {
		if _, exists := index.Features[id]; !exists {
			// Find the first reference to report
			var ref FileReference
			if len(index.Impls[id]) > 0 {
				ref = index.Impls[id][0]
			} else if len(index.Tests[id]) > 0 {
				ref = index.Tests[id][0]
			}
			issues = append(issues, ValidationIssue{
				Severity:  SeverityError,
				FeatureID: id,
				File:      ref.File,
				Line:      ref.Line,
				Message:   fmt.Sprintf("Feature ID %q referenced in code but not found in features.yaml", id),
			})
		}
	}

	// Rule 2: Validate each feature according to its status
	for id, fs := range index.Features {
		impls := index.Impls[id]
		tests := index.Tests[id]

		switch fs.Status {
		case FeatureStatusTodo:
			// todo features: warnings allowed, no hard failures
			if len(impls) == 0 && len(tests) == 0 && fs.Spec != "" {
				// Check if spec exists
				if _, err := os.Stat(fs.Spec); err == nil {
					issues = append(issues, ValidationIssue{
						Severity:  SeverityWarning,
						FeatureID: id,
						File:      fs.Spec,
						Line:      0,
						Message:   "todo feature has spec but no implementation or tests",
					})
				}
			}

		case FeatureStatusWIP:
			// wip features: spec required, at least one impl OR test required
			if fs.Spec == "" {
				issues = append(issues, ValidationIssue{
					Severity:  SeverityError,
					FeatureID: id,
					File:      "",
					Line:      0,
					Message:   "wip feature must have a spec path",
				})
			} else {
				// Check if spec file exists
				if _, err := os.Stat(fs.Spec); err != nil {
					if os.IsNotExist(err) {
						issues = append(issues, ValidationIssue{
							Severity:  SeverityError,
							FeatureID: id,
							File:      fs.Spec,
							Line:      0,
							Message:   "wip feature spec file does not exist",
						})
					}
				}
			}

			if len(impls) == 0 && len(tests) == 0 {
				issues = append(issues, ValidationIssue{
					Severity:  SeverityError,
					FeatureID: id,
					File:      "",
					Line:      0,
					Message:   "wip feature must have at least one implementation or test file",
				})
			}

			// Check Spec: header consistency for wip features
			for _, ref := range impls {
				if ref.SpecPath != "" {
					// Normalize both paths for comparison relative to rootDir
					refSpec := normalizeSpecPathForComparison(ref.SpecPath, index.RootDir)
					expectedSpec := normalizeSpecPathForComparison(fs.Spec, index.RootDir)
					if refSpec != expectedSpec {
						issues = append(issues, ValidationIssue{
							Severity:  SeverityError,
							FeatureID: id,
							File:      ref.File,
							Line:      ref.Line,
							Message:   fmt.Sprintf("Spec header mismatch: got %q, expected %q", ref.SpecPath, fs.Spec),
						})
					}
				}
			}

		case FeatureStatusDone:
			// done features: spec required, impl required, tests required
			if fs.Spec == "" {
				issues = append(issues, ValidationIssue{
					Severity:  SeverityError,
					FeatureID: id,
					File:      "",
					Line:      0,
					Message:   "done feature must have a spec path",
				})
			} else {
				// Check if spec file exists
				if _, err := os.Stat(fs.Spec); err != nil {
					if os.IsNotExist(err) {
						issues = append(issues, ValidationIssue{
							Severity:  SeverityError,
							FeatureID: id,
							File:      fs.Spec,
							Line:      0,
							Message:   "done feature spec file does not exist",
						})
					}
				}
			}

			if len(impls) == 0 {
				issues = append(issues, ValidationIssue{
					Severity:  SeverityError,
					FeatureID: id,
					File:      "",
					Line:      0,
					Message:   "done feature must have at least one implementation file with Feature header",
				})
			}

			if len(tests) == 0 {
				issues = append(issues, ValidationIssue{
					Severity:  SeverityError,
					FeatureID: id,
					File:      "",
					Line:      0,
					Message:   "done feature must have at least one test file with Feature header",
				})
			}

			// Check Feature: and Spec: header consistency for done features
			for _, ref := range impls {
				if ref.SpecPath == "" {
					issues = append(issues, ValidationIssue{
						Severity:  SeverityError,
						FeatureID: id,
						File:      ref.File,
						Line:      ref.Line,
						Message:   "done feature implementation missing Spec header",
					})
				} else {
					// Normalize both paths for comparison relative to rootDir
					refSpec := normalizeSpecPathForComparison(ref.SpecPath, index.RootDir)
					expectedSpec := normalizeSpecPathForComparison(fs.Spec, index.RootDir)
					if refSpec != expectedSpec {
						issues = append(issues, ValidationIssue{
							Severity:  SeverityError,
							FeatureID: id,
							File:      ref.File,
							Line:      ref.Line,
							Message:   fmt.Sprintf("Spec header mismatch: got %q, expected %q", ref.SpecPath, fs.Spec),
						})
					}
				}
			}

			// Check that test files reference the correct Feature ID
			for _, ref := range tests {
				if ref.FeatureID != id {
					issues = append(issues, ValidationIssue{
						Severity:  SeverityError,
						FeatureID: id,
						File:      ref.File,
						Line:      ref.Line,
						Message:   fmt.Sprintf("test file references wrong Feature ID: got %q, expected %q", ref.FeatureID, id),
					})
				}
			}
		}
	}

	// Rule 3: Check for duplicate spec mappings (multiple features pointing to same spec)
	// This is allowed for some shared specs, but we should warn about it
	specToFeatures := make(map[string][]string)
	for id, fs := range index.Features {
		if fs.Spec != "" {
			specToFeatures[fs.Spec] = append(specToFeatures[fs.Spec], id)
		}
	}
	for spec, featureIDs := range specToFeatures {
		if len(featureIDs) > 1 {
			// Multiple features sharing a spec is allowed, but we log it for awareness
			// Only warn if it's a done feature, as shared specs are common for interfaces
			for _, fid := range featureIDs {
				if fs, exists := index.Features[fid]; exists && fs.Status == FeatureStatusDone {
					issues = append(issues, ValidationIssue{
						Severity:  SeverityWarning,
						FeatureID: fid,
						File:      spec,
						Line:      0,
						Message:   fmt.Sprintf("spec file shared by multiple features: %v", featureIDs),
					})
				}
			}
		}
	}

	// Rule 4: Check for orphan specs (spec files that don't match any feature)
	// This is a more complex check that would require walking spec/ directory
	// For now, we skip this as it's less critical than the above rules

	// Deterministic sorting of issues.
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].FeatureID != issues[j].FeatureID {
			return issues[i].FeatureID < issues[j].FeatureID
		}
		if issues[i].Severity != issues[j].Severity {
			return issues[i].Severity < issues[j].Severity
		}
		if issues[i].File != issues[j].File {
			return issues[i].File < issues[j].File
		}
		if issues[i].Line != issues[j].Line {
			return issues[i].Line < issues[j].Line
		}
		return issues[i].Message < issues[j].Message
	})

	return issues, nil
}

// normalizeSpecPathForComparison normalizes spec paths for comparison by resolving
// relative paths relative to rootDir. Both paths are normalized to absolute paths
// for consistent comparison.
func normalizeSpecPathForComparison(specPath, rootDir string) string {
	if specPath == "" {
		return ""
	}
	// If already absolute, clean and return
	if filepath.IsAbs(specPath) {
		return filepath.Clean(specPath)
	}
	// Resolve relative to rootDir
	if rootDir != "" {
		abs := filepath.Join(rootDir, specPath)
		return filepath.Clean(abs)
	}
	// Fallback: resolve relative to current directory
	abs, err := filepath.Abs(specPath)
	if err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(specPath)
}
