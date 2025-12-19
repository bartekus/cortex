// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package suggestions converts commit health and feature traceability reports
// into actionable suggestions.
//
// Feature: CLI_COMMAND_COMMIT
// Spec: spec/cli/commit.md
package suggestions

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/reports/commithealth"
	"github.com/bartekus/cortex/internal/reports/featuretrace"
)

// Suggestion represents a single actionable suggestion.
type Suggestion struct {
	ID       string         `json:"id"`
	Type     SuggestionType `json:"type"`
	Severity Severity       `json:"severity"`
	Message  string         `json:"message"`
	Details  map[string]any `json:"details"`
	Fix      *Fix           `json:"fix,omitempty"`
}

// SuggestionType represents the category of a suggestion.
type SuggestionType string

// Suggestion type constants.
const (
	SuggestionTypeCommitFormat        SuggestionType = "commit_format"
	SuggestionTypeFeatureID           SuggestionType = "feature_id"
	SuggestionTypeFeatureTraceability SuggestionType = "feature_traceability"
	SuggestionTypeSummary             SuggestionType = "summary"
)

// Severity represents the severity level of a suggestion.
type Severity string

// Severity level constants.
const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// Fix represents actionable guidance for resolving a suggestion.
// In v1, this is intentionally minimal and reserved for future auto-fix flows.
type Fix struct {
	Action           string         `json:"action,omitempty"`
	SuggestedMessage string         `json:"suggested_message,omitempty"`
	Command          string         `json:"command,omitempty"`
	Details          map[string]any `json:"details,omitempty"`
}

// Report represents the complete suggestions report for JSON output.
type Report struct {
	SchemaVersion string       `json:"schema_version"`
	Summary       Summary      `json:"summary"`
	Suggestions   []Suggestion `json:"suggestions"`
}

// Summary contains aggregate statistics for suggestions.
type Summary struct {
	TotalSuggestions int            `json:"total_suggestions"`
	BySeverity       map[string]int `json:"by_severity"`
	ByType           map[string]int `json:"by_type"`
}

// BuildReport builds a structured report from suggestions for JSON output.
// This mirrors the pattern used in commithealth and featuretrace packages.
func BuildReport(sugs []Suggestion) Report {
	summary := Summary{
		TotalSuggestions: len(sugs),
		BySeverity:       make(map[string]int),
		ByType:           make(map[string]int),
	}

	for _, s := range sugs {
		summary.BySeverity[string(s.Severity)]++
		summary.ByType[string(s.Type)]++
	}

	// Ensure empty slice marshals as [] instead of null
	if sugs == nil {
		sugs = []Suggestion{}
	}

	return Report{
		SchemaVersion: "1.0",
		Summary:       summary,
		Suggestions:   sugs,
	}
}

// GenerateSuggestions converts commit health and feature traceability reports
// into a slice of raw suggestions.
//
// v1 behaviour:
//   - Walk all commit violations in commithealth.Report
//   - Map each violation to a Suggestion (type, severity, ID, message, details)
//   - (Scaffold) Reserve hooks for featuretrace-based suggestions
//
// The caller is expected to pass the result through PrioritizeSuggestions and
// FilterSuggestions before rendering.
func GenerateSuggestions(
	commitReport *commithealth.Report,
	featureReport *featuretrace.Report,
) ([]Suggestion, error) {
	var out []Suggestion

	// 1. Suggestions derived from commit health violations.
	out = append(out, suggestionsFromCommitHealth(commitReport)...)

	// 2. Suggestions derived from feature traceability (scaffold).
	//
	// NOTE:
	// We do not know the exact shape of featuretrace.Report here, so this
	// function is deliberately a no-op placeholder. Once the featuretrace
	// report exposes feature-level problems (or a slice of FeaturePresence),
	// this hook should be implemented to map those problems into suggestions.
	out = append(out, suggestionsFromFeatureTrace(featureReport)...)

	return out, nil
}

// suggestionsFromCommitHealth walks the commit-health report and converts each
// violation into a Suggestion. It assumes the following (inferred) shapes:
//
//	type Report struct {
//	    Commits map[string]Commit
//	    // ...
//	}
//
//	type Commit struct {
//	    Subject    string
//	    IsValid    bool
//	    Violations []Violation
//	}
//
//	type Violation struct {
//	    Code     commithealth.ViolationCode
//	    Severity commithealth.Severity
//	    Message  string
//	    Details  map[string]any
//	}
//
// If the actual shapes differ, adjust this helper accordingly.
func suggestionsFromCommitHealth(report *commithealth.Report) []Suggestion {
	if len(report.Commits) == 0 {
		return nil
	}

	suggestions := make([]Suggestion, 0, len(report.Commits)) // lower bound; may grow

	for sha, commit := range report.Commits {
		if len(commit.Violations) == 0 {
			continue
		}

		for _, v := range commit.Violations {
			s := Suggestion{
				ID:       fmt.Sprintf("commit-%s-%s", sha, v.Code),
				Type:     mapViolationCodeToSuggestionType(v.Code),
				Severity: mapCommitSeverity(v.Severity),
				Message:  fmt.Sprintf("Commit %s: %s", sha, v.Message),
				Details: map[string]any{
					"commit_sha":     sha,
					"subject":        commit.Subject,
					"violation_code": string(v.Code),
					"severity":       string(v.Severity),
				},
				// Fix is intentionally nil in v1; future phases may populate this.
				Fix: nil,
			}

			suggestions = append(suggestions, s)
		}
	}

	return suggestions
}

// suggestionsFromFeatureTrace is deliberately minimal in v1.
// It exists as a hook for Phase 3.D+ when feature-level "problems" are
// exposed in featuretrace.Report.
//
// For now it returns an empty slice to keep behaviour well-defined and
// deterministic; commit-based suggestions are the only source.
func suggestionsFromFeatureTrace(_ *featuretrace.Report) []Suggestion {
	// TODO (Phase 3.D+):
	//  - Expose feature-level problems (or iterate FeaturePresence entries)
	//  - Derive suggestions for:
	//      * Features marked done but missing spec / impl / tests
	//      * Features with no referencing commits
	//  - Use IDs of the form: feature-<featureID>-<problem_code>
	//  - Map to SuggestionTypeFeatureTraceability
	return nil
}

// mapViolationCodeToSuggestionType maps a commit-health violation code onto a
// SuggestionType. This mapping is intentionally conservative and can be
// extended as new rules are added.
func mapViolationCodeToSuggestionType(code commithealth.ViolationCode) SuggestionType {
	switch code {
	case commithealth.ViolationCodeMissingFeatureID,
		commithealth.ViolationCodeMultipleFeatureIDs,
		commithealth.ViolationCodeInvalidFeatureIDFormat,
		commithealth.ViolationCodeFeatureIDNotInSpec:
		return SuggestionTypeFeatureID

	case commithealth.ViolationCodeSummaryTooLong,
		commithealth.ViolationCodeSummaryHasTrailingPeriod,
		commithealth.ViolationCodeSummaryStartsWithUppercase:
		return SuggestionTypeSummary

	case commithealth.ViolationCodeInvalidFormatGeneric:
		// Generic format issues without a more specific category.
		return SuggestionTypeCommitFormat

	default:
		// Unknown codes fall back to the generic commit_format bucket.
		return SuggestionTypeCommitFormat
	}
}

// mapCommitSeverity translates commithealth.Severity into the local Severity
// type, defaulting to SeverityWarning for unknown values (middle ground).
func mapCommitSeverity(s commithealth.Severity) Severity {
	switch s {
	case commithealth.SeverityError:
		return SeverityError
	case commithealth.SeverityWarning:
		return SeverityWarning
	case commithealth.SeverityInfo:
		return SeverityInfo
	default:
		// Defensive default: treat unknown severities as warnings.
		return SeverityWarning
	}
}

// FormatSuggestionsText renders a deterministic, human-readable summary of
// suggestions.
//
// Ordering rules:
//   - Groups by severity: error, warning, info
//   - Within each group, suggestions are expected to already be ordered
//     (callers should pass the result of PrioritizeSuggestions + FilterSuggestions).
//
// This function does not re-sort suggestions beyond grouping; it assumes the
// caller has already applied the canonical ordering.
func FormatSuggestionsText(suggestions []Suggestion) string {
	if len(suggestions) == 0 {
		return "No suggestions.\n"
	}

	var buf bytes.Buffer

	// Top-level heading.
	buf.WriteString("Commit Discipline Suggestions\n")
	buf.WriteString("============================\n\n")

	// Group by severity but preserve intra-group order.
	grouped := map[Severity][]Suggestion{
		SeverityError:   {},
		SeverityWarning: {},
		SeverityInfo:    {},
	}

	for _, s := range suggestions {
		grouped[s.Severity] = append(grouped[s.Severity], s)
	}

	orderedSeverities := []Severity{SeverityError, SeverityWarning, SeverityInfo}

	// Severity headings in deterministic order.
	for _, sev := range orderedSeverities {
		items := grouped[sev]
		if len(items) == 0 {
			continue
		}

		// Section heading, e.g. "Errors (2)"
		heading := severityHeading(sev)
		fmt.Fprintf(&buf, "%s (%d)\n", heading, len(items))
		buf.WriteString(strings.Repeat("-", len(heading)+len(fmt.Sprintf(" (%d)", len(items)))) + "\n")

		// Deterministic per-suggestion output.
		for i := range items {
			s := &items[i]
			fmt.Fprintf(&buf, "[%s] %s\n", suggestionCode(s), s.Message)

			// Stable details: walk keys in lexicographic order.
			if len(s.Details) > 0 {
				keys := make([]string, 0, len(s.Details))
				for k := range s.Details {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, k := range keys {
					fmt.Fprintf(&buf, "  %s: %v\n", k, s.Details[k])
				}
			}

			buf.WriteString("\n")
		}
	}

	// Final summary section.
	total := len(suggestions)
	counts := map[Severity]int{}
	for _, s := range suggestions {
		counts[s.Severity]++
	}

	buf.WriteString("Summary\n")
	buf.WriteString("=======\n")
	fmt.Fprintf(&buf, "Total suggestions: %d\n", total)
	fmt.Fprintf(&buf, "  Errors: %d\n", counts[SeverityError])
	fmt.Fprintf(&buf, "  Warnings: %d\n", counts[SeverityWarning])
	fmt.Fprintf(&buf, "  Info: %d\n", counts[SeverityInfo])
	buf.WriteString("\n")

	return buf.String()
}

// severityHeading returns a human-readable heading label for a severity.
func severityHeading(sev Severity) string {
	switch sev {
	case SeverityError:
		return "Errors"
	case SeverityWarning:
		return "Warnings"
	case SeverityInfo:
		return "Info"
	default:
		return "Unknown"
	}
}

// suggestionCode produces a short, deterministic label for a suggestion.
// For now this is a simple prefix based on severity.
//
// Examples:
//   - error   -> "E"
//   - warning -> "W"
//   - info    -> "I"
func suggestionCode(s *Suggestion) string {
	switch s.Severity {
	case SeverityError:
		return "E"
	case SeverityWarning:
		return "W"
	case SeverityInfo:
		return "I"
	default:
		return "U"
	}
}

// PrioritizeSuggestions sorts suggestions by severity (error > warning > info),
// then by lexicographical order of ID for determinism.
//
// This is a simple v1 implementation. Future versions may add frequency-based
// prioritization or other heuristics.
func PrioritizeSuggestions(suggestions []Suggestion) []Suggestion {
	// Create a copy to avoid mutating the input
	result := make([]Suggestion, len(suggestions))
	copy(result, suggestions)

	// Sort by severity first, then by ID for determinism
	sort.Slice(result, func(i, j int) bool {
		// Severity ordering: error > warning > info
		severityOrder := map[Severity]int{
			SeverityError:   0,
			SeverityWarning: 1,
			SeverityInfo:    2,
		}

		sevI := severityOrder[result[i].Severity]
		sevJ := severityOrder[result[j].Severity]

		if sevI != sevJ {
			return sevI < sevJ
		}

		// Same severity: sort by ID lexicographically
		return result[i].ID < result[j].ID
	})

	return result
}

// FilterSuggestions filters suggestions by minimum severity and maximum count.
//
// minSeverity: Only include suggestions with severity >= minSeverity
//   - "error": only errors
//   - "warning": errors and warnings
//   - "info": all suggestions
//
// maxCount: Maximum number of suggestions to return (0 = unlimited)
func FilterSuggestions(suggestions []Suggestion, minSeverity Severity, maxCount int) []Suggestion {
	severityOrder := map[Severity]int{
		SeverityError:   0,
		SeverityWarning: 1,
		SeverityInfo:    2,
	}

	minOrder := severityOrder[minSeverity]

	var filtered []Suggestion
	for _, s := range suggestions {
		if severityOrder[s.Severity] <= minOrder {
			filtered = append(filtered, s)
		}
	}

	// Apply max count limit
	if maxCount > 0 && len(filtered) > maxCount {
		filtered = filtered[:maxCount]
	}

	return filtered
}
