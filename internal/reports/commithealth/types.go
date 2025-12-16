// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package commithealth defines the data model for commit health reports.
//
// Feature: GOV_CORE
// Docs: docs/design/commit-reports-go-types.md
package commithealth

// Report represents the complete commit health report.
type Report struct {
	SchemaVersion string            `json:"schema_version"`
	GeneratedAt   string            `json:"generated_at,omitempty"`
	Repo          RepoInfo          `json:"repo"`
	Range         CommitRange       `json:"range"`
	Summary       Summary           `json:"summary"`
	Rules         []Rule            `json:"rules"`
	Commits       map[string]Commit `json:"commits"`
}

// RepoInfo contains repository metadata.
type RepoInfo struct {
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
}

// CommitRange describes the commit range analyzed.
type CommitRange struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Description string `json:"description"`
}

// Summary contains aggregate statistics.
type Summary struct {
	TotalCommits     int                   `json:"total_commits"`
	ValidCommits     int                   `json:"valid_commits"`
	InvalidCommits   int                   `json:"invalid_commits"`
	ViolationsByCode map[ViolationCode]int `json:"violations_by_code"`
}

// Rule describes a commit validation rule.
type Rule struct {
	Code        ViolationCode `json:"code"`
	Description string        `json:"description"`
	Severity    Severity      `json:"severity"`
}

// Commit represents a single commit's health status.
type Commit struct {
	Subject    string      `json:"subject"`
	IsValid    bool        `json:"is_valid"`
	Violations []Violation `json:"violations"`
}

// Violation represents a single validation violation.
type Violation struct {
	Code     ViolationCode  `json:"code"`
	Severity Severity       `json:"severity"`
	Message  string         `json:"message"`
	Details  map[string]any `json:"details"`
}

// ViolationCode represents known violation codes.
type ViolationCode string

// Violation code constants for commit message validation.
const (
	ViolationCodeMissingFeatureID           ViolationCode = "MISSING_FEATURE_ID"
	ViolationCodeInvalidType                ViolationCode = "INVALID_TYPE"
	ViolationCodeInvalidFeatureIDFormat     ViolationCode = "INVALID_FEATURE_ID_FORMAT"
	ViolationCodeFeatureIDNotInSpec         ViolationCode = "FEATURE_ID_NOT_IN_SPEC"
	ViolationCodeFeatureIDBranchMismatch    ViolationCode = "FEATURE_ID_BRANCH_MISMATCH"
	ViolationCodeMultipleFeatureIDs         ViolationCode = "MULTIPLE_FEATURE_IDS"
	ViolationCodeSummaryTooLong             ViolationCode = "SUMMARY_TOO_LONG"
	ViolationCodeSummaryHasTrailingPeriod   ViolationCode = "SUMMARY_HAS_TRAILING_PERIOD"
	ViolationCodeSummaryStartsWithUppercase ViolationCode = "SUMMARY_STARTS_WITH_UPPERCASE"
	ViolationCodeSummaryHasUnicode          ViolationCode = "SUMMARY_HAS_UNICODE"
	ViolationCodeInvalidFormatGeneric       ViolationCode = "INVALID_FORMAT_GENERIC"
)

// Severity represents violation severity levels.
type Severity string

// Severity level constants for commit validation violations.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)
