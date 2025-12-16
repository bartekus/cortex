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

// Report represents the complete feature traceability report.
type Report struct {
	SchemaVersion string             `json:"schema_version"`
	GeneratedAt   string             `json:"generated_at,omitempty"`
	Summary       Summary            `json:"summary"`
	Features      map[string]Feature `json:"features"`
}

// Summary contains aggregate statistics.
type Summary struct {
	TotalFeatures    int `json:"total_features"`
	Done             int `json:"done"`
	WIP              int `json:"wip"`
	Todo             int `json:"todo"`
	Deprecated       int `json:"deprecated"`
	Removed          int `json:"removed"`
	FeaturesWithGaps int `json:"features_with_gaps"`
}

// Feature represents traceability information for a single feature.
type Feature struct {
	Status         FeatureStatus      `json:"status"`
	Spec           SpecInfo           `json:"spec"`
	Implementation ImplementationInfo `json:"implementation"`
	Tests          TestsInfo          `json:"tests"`
	Commits        CommitsInfo        `json:"commits"`
	Problems       []Problem          `json:"problems"`
}

// FeatureStatus represents the lifecycle state of a feature.
type FeatureStatus string

// Feature status constants for feature lifecycle tracking.
const (
	FeatureStatusTodo       FeatureStatus = "todo"
	FeatureStatusWIP        FeatureStatus = "wip"
	FeatureStatusDone       FeatureStatus = "done"
	FeatureStatusDeprecated FeatureStatus = "deprecated"
	FeatureStatusRemoved    FeatureStatus = "removed"
)

// SpecInfo describes spec file presence and location.
type SpecInfo struct {
	Present bool   `json:"present"`
	Path    string `json:"path"` // empty string if not present
}

// ImplementationInfo describes implementation file presence and locations.
type ImplementationInfo struct {
	Present bool     `json:"present"`
	Files   []string `json:"files"` // sorted list of file paths
}

// TestsInfo describes test file presence and locations.
type TestsInfo struct {
	Present bool     `json:"present"`
	Files   []string `json:"files"` // sorted list of test file paths
}

// CommitsInfo describes commit presence and SHAs.
type CommitsInfo struct {
	Present bool     `json:"present"`
	SHAs    []string `json:"shas"` // sorted list of commit SHAs
}

// Problem represents a traceability problem for a feature.
type Problem struct {
	Code     ProblemCode    `json:"code"`
	Severity Severity       `json:"severity"`
	Message  string         `json:"message"`
	Details  map[string]any `json:"details"`
}

// ProblemCode represents known problem codes.
type ProblemCode string

// Problem code constants for feature traceability issues.
const (
	ProblemCodeMissingSpec                        ProblemCode = "MISSING_SPEC"
	ProblemCodeMissingImplementation              ProblemCode = "MISSING_IMPLEMENTATION"
	ProblemCodeMissingTests                       ProblemCode = "MISSING_TESTS"
	ProblemCodeMissingCommits                     ProblemCode = "MISSING_COMMITS"
	ProblemCodeOrphanSpec                         ProblemCode = "ORPHAN_SPEC"
	ProblemCodeOrphanFeatureIDInCommits           ProblemCode = "ORPHAN_FEATURE_ID_IN_COMMITS"
	ProblemCodeStatusDoneButMissingTests          ProblemCode = "STATUS_DONE_BUT_MISSING_TESTS"
	ProblemCodeStatusDoneButMissingImplementation ProblemCode = "STATUS_DONE_BUT_MISSING_IMPLEMENTATION"
	ProblemCodeUnreferencedSpecPath               ProblemCode = "UNREFERENCED_SPEC_PATH"
)

// Severity represents problem severity levels.
type Severity string

// Severity level constants for feature traceability problems.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)
