// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package featuretrace defines the data model for feature traceability reports.
//
// Feature: GOV_CORE
// Docs: docs/design/commit-reports-go-types.md
package featuretrace

import (
	"sort"
)

// FeaturePresence represents the presence of a feature across spec, implementation, tests, and commits.
type FeaturePresence struct {
	FeatureID           string
	Status              FeatureStatus
	HasSpec             bool
	SpecPath            string
	ImplementationFiles []string
	TestFiles           []string
	CommitSHAs          []string
}

// GenerateFeatureTraceabilityReport generates a feature traceability report from feature presence data.
func GenerateFeatureTraceabilityReport(features []FeaturePresence) (Report, error) {
	report := Report{
		SchemaVersion: "1.0",
		Summary: Summary{
			TotalFeatures:    len(features),
			Done:             0,
			WIP:              0,
			Todo:             0,
			Deprecated:       0,
			Removed:          0,
			FeaturesWithGaps: 0,
		},
		Features: make(map[string]Feature),
	}

	// Process each feature
	for i := range features {
		fp := &features[i]
		feature := Feature{
			Status: fp.Status,
			Spec: SpecInfo{
				Present: fp.HasSpec,
				Path:    fp.SpecPath,
			},
			Implementation: ImplementationInfo{
				Present: len(fp.ImplementationFiles) > 0,
				Files:   sortedCopy(fp.ImplementationFiles),
			},
			Tests: TestsInfo{
				Present: len(fp.TestFiles) > 0,
				Files:   sortedCopy(fp.TestFiles),
			},
			Commits: CommitsInfo{
				Present: len(fp.CommitSHAs) > 0,
				SHAs:    sortedCopy(fp.CommitSHAs),
			},
			Problems: []Problem{},
		}

		// Detect problems
		problems := detectProblems(fp, &feature)
		feature.Problems = problems

		// Update summary counts
		switch fp.Status {
		case FeatureStatusDone:
			report.Summary.Done++
		case FeatureStatusWIP:
			report.Summary.WIP++
		case FeatureStatusTodo:
			report.Summary.Todo++
		case FeatureStatusDeprecated:
			report.Summary.Deprecated++
		case FeatureStatusRemoved:
			report.Summary.Removed++
		}

		// Check if feature has gaps
		if len(problems) > 0 {
			report.Summary.FeaturesWithGaps++
		}

		report.Features[fp.FeatureID] = feature
	}

	return report, nil
}

// detectProblems detects traceability problems for a feature.
func detectProblems(fp *FeaturePresence, feature *Feature) []Problem {
	var problems []Problem

	// Missing spec
	if !fp.HasSpec {
		problems = append(problems, Problem{
			Code:     ProblemCodeMissingSpec,
			Severity: SeverityError,
			Message:  "Feature has no spec file.",
			Details:  map[string]any{},
		})
	}

	// Missing implementation (if spec exists)
	if fp.HasSpec && !feature.Implementation.Present {
		problems = append(problems, Problem{
			Code:     ProblemCodeMissingImplementation,
			Severity: SeverityWarning,
			Message:  "Feature has a spec but no implementation files.",
			Details:  map[string]any{},
		})
	}

	// Missing tests (if spec exists)
	if fp.HasSpec && !feature.Tests.Present {
		problems = append(problems, Problem{
			Code:     ProblemCodeMissingTests,
			Severity: SeverityWarning,
			Message:  "Feature has a spec but no tests.",
			Details:  map[string]any{},
		})
	}

	// Missing commits
	if !feature.Commits.Present {
		problems = append(problems, Problem{
			Code:     ProblemCodeMissingCommits,
			Severity: SeverityInfo,
			Message:  "Feature has no commits referencing this Feature ID yet.",
			Details:  map[string]any{},
		})
	}

	// Status done but missing tests
	if fp.Status == FeatureStatusDone && !feature.Tests.Present {
		problems = append(problems, Problem{
			Code:     ProblemCodeStatusDoneButMissingTests,
			Severity: SeverityWarning,
			Message:  "Feature status is 'done' but has no tests.",
			Details:  map[string]any{},
		})
	}

	// Status done but missing implementation
	if fp.Status == FeatureStatusDone && !feature.Implementation.Present {
		problems = append(problems, Problem{
			Code:     ProblemCodeStatusDoneButMissingImplementation,
			Severity: SeverityWarning,
			Message:  "Feature status is 'done' but has no implementation files.",
			Details:  map[string]any{},
		})
	}

	return problems
}

// sortedCopy returns a sorted copy of a string slice.
func sortedCopy(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	result := make([]string, len(s))
	copy(result, s)
	sort.Strings(result)
	return result
}
