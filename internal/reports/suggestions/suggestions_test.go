// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CLI_COMMIT_SUGGEST
// Spec: spec/commands/commit-suggest.md
package suggestions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bartekus/cortex/internal/reports/commithealth"
	"github.com/bartekus/cortex/internal/reports/featuretrace"
)

func TestGenerateSuggestions_CommitHealthViolations(t *testing.T) {
	t.Parallel()

	// Build a minimal commithealth.Report with one commit and two violations
	commitSHA := "abc123"
	report := commithealth.Report{
		SchemaVersion: "1.0",
		Commits: map[string]commithealth.Commit{
			commitSHA: {
				Subject: "feat: add deploy support",
				IsValid: false,
				Violations: []commithealth.Violation{
					{
						Code:     commithealth.ViolationCodeMissingFeatureID,
						Severity: commithealth.SeverityError,
						Message:  "Commit message is missing a Feature ID in the required format.",
						Details:  map[string]any{},
					},
					{
						Code:     commithealth.ViolationCodeSummaryTooLong,
						Severity: commithealth.SeverityWarning,
						Message:  "Commit summary exceeds 72 characters.",
						Details:  map[string]any{},
					},
				},
			},
		},
	}

	// Empty feature report (scaffold)
	featureReport := featuretrace.Report{
		SchemaVersion: "1.0",
		Features:      make(map[string]featuretrace.Feature),
	}

	// Generate suggestions
	suggestions, err := GenerateSuggestions(&report, &featureReport)
	if err != nil {
		t.Fatalf("GenerateSuggestions failed: %v", err)
	}

	// Verify we got exactly two suggestions
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}

	// Find suggestions by ID
	var missingFeatureID *Suggestion
	var summaryTooLong *Suggestion

	for i := range suggestions {
		if suggestions[i].ID == fmt.Sprintf("commit-%s-%s", commitSHA, commithealth.ViolationCodeMissingFeatureID) {
			missingFeatureID = &suggestions[i]
		}
		if suggestions[i].ID == fmt.Sprintf("commit-%s-%s", commitSHA, commithealth.ViolationCodeSummaryTooLong) {
			summaryTooLong = &suggestions[i]
		}
	}

	if missingFeatureID == nil {
		t.Fatal("missing suggestion for MISSING_FEATURE_ID violation")
		return
	}
	if summaryTooLong == nil {
		t.Fatal("missing suggestion for SUMMARY_TOO_LONG violation")
		return
	}

	// Verify MISSING_FEATURE_ID suggestion
	if missingFeatureID.Type != SuggestionTypeFeatureID {
		t.Errorf("expected type %s, got %s", SuggestionTypeFeatureID, missingFeatureID.Type)
	}
	if missingFeatureID.Severity != SeverityError {
		t.Errorf("expected severity %s, got %s", SeverityError, missingFeatureID.Severity)
	}
	if missingFeatureID.Details["commit_sha"] != commitSHA {
		t.Errorf("expected commit_sha=%s, got %v", commitSHA, missingFeatureID.Details["commit_sha"])
	}
	if missingFeatureID.Details["subject"] != "feat: add deploy support" {
		t.Errorf("expected subject='feat: add deploy support', got %v", missingFeatureID.Details["subject"])
	}

	// Verify SUMMARY_TOO_LONG suggestion
	if summaryTooLong.Type != SuggestionTypeSummary {
		t.Errorf("expected type %s, got %s", SuggestionTypeSummary, summaryTooLong.Type)
	}
	if summaryTooLong.Severity != SeverityWarning {
		t.Errorf("expected severity %s, got %s", SeverityWarning, summaryTooLong.Severity)
	}
	if summaryTooLong.Details["commit_sha"] != commitSHA {
		t.Errorf("expected commit_sha=%s, got %v", commitSHA, summaryTooLong.Details["commit_sha"])
	}
}

func TestGenerateSuggestions_ValidCommitsProduceNoSuggestions(t *testing.T) {
	t.Parallel()

	// Build a report with only valid commits (no violations)
	report := commithealth.Report{
		SchemaVersion: "1.0",
		Commits: map[string]commithealth.Commit{
			"abc123": {
				Subject:    "feat(CLI_DEPLOY): add deploy support",
				IsValid:    true,
				Violations: nil,
			},
		},
	}

	featureReport := featuretrace.Report{
		SchemaVersion: "1.0",
		Features:      make(map[string]featuretrace.Feature),
	}

	suggestions, err := GenerateSuggestions(&report, &featureReport)
	if err != nil {
		t.Fatalf("GenerateSuggestions failed: %v", err)
	}

	// Valid commits should produce no suggestions
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for valid commits, got %d", len(suggestions))
	}
}

func TestMapViolationCodeToSuggestionType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     commithealth.ViolationCode
		expected SuggestionType
	}{
		{
			name:     "MISSING_FEATURE_ID -> feature_id",
			code:     commithealth.ViolationCodeMissingFeatureID,
			expected: SuggestionTypeFeatureID,
		},
		{
			name:     "MULTIPLE_FEATURE_IDS -> feature_id",
			code:     commithealth.ViolationCodeMultipleFeatureIDs,
			expected: SuggestionTypeFeatureID,
		},
		{
			name:     "INVALID_FEATURE_ID_FORMAT -> feature_id",
			code:     commithealth.ViolationCodeInvalidFeatureIDFormat,
			expected: SuggestionTypeFeatureID,
		},
		{
			name:     "FEATURE_ID_NOT_IN_SPEC -> feature_id",
			code:     commithealth.ViolationCodeFeatureIDNotInSpec,
			expected: SuggestionTypeFeatureID,
		},
		{
			name:     "SUMMARY_TOO_LONG -> summary",
			code:     commithealth.ViolationCodeSummaryTooLong,
			expected: SuggestionTypeSummary,
		},
		{
			name:     "SUMMARY_HAS_TRAILING_PERIOD -> summary",
			code:     commithealth.ViolationCodeSummaryHasTrailingPeriod,
			expected: SuggestionTypeSummary,
		},
		{
			name:     "SUMMARY_STARTS_WITH_UPPERCASE -> summary",
			code:     commithealth.ViolationCodeSummaryStartsWithUppercase,
			expected: SuggestionTypeSummary,
		},
		{
			name:     "INVALID_FORMAT_GENERIC -> commit_format",
			code:     commithealth.ViolationCodeInvalidFormatGeneric,
			expected: SuggestionTypeCommitFormat,
		},
		{
			name:     "unknown code -> commit_format",
			code:     commithealth.ViolationCode("UNKNOWN_CODE"),
			expected: SuggestionTypeCommitFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapViolationCodeToSuggestionType(tt.code)
			if got != tt.expected {
				t.Errorf("mapViolationCodeToSuggestionType(%s) = %s, want %s", tt.code, got, tt.expected)
			}
		})
	}
}

func TestMapCommitSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity commithealth.Severity
		expected Severity
	}{
		{
			name:     "error -> error",
			severity: commithealth.SeverityError,
			expected: SeverityError,
		},
		{
			name:     "warning -> warning",
			severity: commithealth.SeverityWarning,
			expected: SeverityWarning,
		},
		{
			name:     "info -> info",
			severity: commithealth.SeverityInfo,
			expected: SeverityInfo,
		},
		{
			name:     "unknown -> warning (defensive default)",
			severity: commithealth.Severity("unknown"),
			expected: SeverityWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapCommitSeverity(tt.severity)
			if got != tt.expected {
				t.Errorf("mapCommitSeverity(%s) = %s, want %s", tt.severity, got, tt.expected)
			}
		})
	}
}

func TestFormatSuggestionsText(t *testing.T) {
	t.Parallel()

	t.Run("empty suggestions", func(t *testing.T) {
		t.Parallel()
		out := FormatSuggestionsText(nil)
		expected := "No suggestions.\n"
		if out != expected {
			t.Errorf("FormatSuggestionsText(nil) = %q, want %q", out, expected)
		}
	})

	t.Run("suggestions grouped by severity", func(t *testing.T) {
		t.Parallel()
		suggestions := []Suggestion{
			{
				ID:       "commit-abc123-MISSING_FEATURE_ID",
				Type:     SuggestionTypeFeatureID,
				Severity: SeverityError,
				Message:  "Commit abc123: Missing Feature ID",
				Details: map[string]any{
					"commit_sha": "abc123",
					"subject":    "feat: add deploy support",
				},
			},
			{
				ID:       "commit-def456-SUMMARY_TOO_LONG",
				Type:     SuggestionTypeSummary,
				Severity: SeverityWarning,
				Message:  "Commit def456: Summary too long",
				Details: map[string]any{
					"commit_sha": "def456",
					"subject":    "feat(CLI_DEPLOY): add comprehensive deployment support with many features",
				},
			},
		}

		out := FormatSuggestionsText(suggestions)

		// Verify structure
		if !strings.Contains(out, "Commit Discipline Suggestions") {
			t.Error("output missing main heading")
		}
		if !strings.Contains(out, "Errors (1)") {
			t.Error("output missing errors section")
		}
		if !strings.Contains(out, "Warnings (1)") {
			t.Error("output missing warnings section")
		}
		if !strings.Contains(out, "Summary") {
			t.Error("output missing summary section")
		}
		if !strings.Contains(out, "Total suggestions: 2") {
			t.Error("output missing total count")
		}
		if !strings.Contains(out, "[E]") {
			t.Error("output missing error suggestion code")
		}
		if !strings.Contains(out, "[W]") {
			t.Error("output missing warning suggestion code")
		}
	})
}

func TestPrioritizeSuggestions(t *testing.T) {
	t.Parallel()

	suggestions := []Suggestion{
		{
			ID:       "commit-zzz-info",
			Severity: SeverityInfo,
			Message:  "Info suggestion",
		},
		{
			ID:       "commit-aaa-error",
			Severity: SeverityError,
			Message:  "Error suggestion",
		},
		{
			ID:       "commit-bbb-warning",
			Severity: SeverityWarning,
			Message:  "Warning suggestion",
		},
	}

	prioritized := PrioritizeSuggestions(suggestions)

	// Should be ordered: error, warning, info
	if prioritized[0].Severity != SeverityError {
		t.Errorf("first suggestion should be error, got %s", prioritized[0].Severity)
	}
	if prioritized[1].Severity != SeverityWarning {
		t.Errorf("second suggestion should be warning, got %s", prioritized[1].Severity)
	}
	if prioritized[2].Severity != SeverityInfo {
		t.Errorf("third suggestion should be info, got %s", prioritized[2].Severity)
	}

	// Within same severity, should be sorted by ID
	if prioritized[0].ID != "commit-aaa-error" {
		t.Errorf("error suggestions should be sorted by ID, got %s", prioritized[0].ID)
	}
}

func TestFilterSuggestions(t *testing.T) {
	t.Parallel()

	suggestions := []Suggestion{
		{ID: "1", Severity: SeverityError},
		{ID: "2", Severity: SeverityWarning},
		{ID: "3", Severity: SeverityInfo},
	}

	t.Run("filter by severity", func(t *testing.T) {
		t.Parallel()
		filtered := FilterSuggestions(suggestions, SeverityWarning, 0)
		if len(filtered) != 2 {
			t.Errorf("expected 2 suggestions (error + warning), got %d", len(filtered))
		}
		if filtered[0].Severity != SeverityError {
			t.Error("first should be error")
		}
		if filtered[1].Severity != SeverityWarning {
			t.Error("second should be warning")
		}
	})

	t.Run("filter by max count", func(t *testing.T) {
		t.Parallel()
		filtered := FilterSuggestions(suggestions, SeverityInfo, 2)
		if len(filtered) != 2 {
			t.Errorf("expected 2 suggestions, got %d", len(filtered))
		}
	})
}
