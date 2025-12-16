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

import (
	"fmt"
	"regexp"
	"strings"
)

// GenerateCommitHealthReport generates a commit health report from commit history.
func GenerateCommitHealthReport(
	commits []CommitMetadata,
	knownFeatures map[string]bool,
	repoInfo RepoInfo,
	rangeInfo CommitRange,
) (Report, error) {
	report := Report{
		SchemaVersion: "1.0",
		Repo:          repoInfo,
		Range:         rangeInfo,
		Summary: Summary{
			TotalCommits:     len(commits),
			ValidCommits:     0,
			InvalidCommits:   0,
			ViolationsByCode: make(map[ViolationCode]int),
		},
		Rules:   getAllRules(),
		Commits: make(map[string]Commit),
	}

	// Process each commit
	for _, commit := range commits {
		subject := extractSubject(commit.Message)
		violations := validateCommitMessage(subject, knownFeatures)

		isValid := len(violations) == 0
		if isValid {
			report.Summary.ValidCommits++
		} else {
			report.Summary.InvalidCommits++
		}

		// Count violations by code
		for _, v := range violations {
			report.Summary.ViolationsByCode[v.Code]++
		}

		report.Commits[commit.SHA] = Commit{
			Subject:    subject,
			IsValid:    isValid,
			Violations: violations,
		}
	}

	return report, nil
}

// extractSubject extracts the first line (subject) from a commit message.
func extractSubject(message string) string {
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[0])
}

// validateCommitMessage validates a commit message subject and returns violations.
func validateCommitMessage(subject string, knownFeatures map[string]bool) []Violation {
	var violations []Violation

	// Pattern: <type>(<FEATURE_ID>): <summary>
	// Allowed types: feat, fix, refactor, docs, test, ci, chore
	pattern := regexp.MustCompile(`^(feat|fix|refactor|docs|test|ci|chore)\(([^)]+)\):\s*(.+)$`)
	matches := pattern.FindStringSubmatch(subject)

	if len(matches) == 0 {
		// Check if it's missing feature ID entirely (no parentheses pattern)
		if !strings.Contains(subject, "(") || !strings.Contains(subject, ")") {
			violations = append(violations, Violation{
				Code:     ViolationCodeMissingFeatureID,
				Severity: SeverityError,
				Message:  "Commit message is missing a Feature ID in the required format.",
				Details:  map[string]any{},
			})
		} else {
			violations = append(violations, Violation{
				Code:     ViolationCodeInvalidFormatGeneric,
				Severity: SeverityError,
				Message:  "Commit message does not match required format: <type>(<FEATURE_ID>): <summary>",
				Details:  map[string]any{},
			})
		}
		return violations
	}

	// Extract feature IDs from the parentheses
	featureIDStr := matches[2]
	featureIDs := parseFeatureIDs(featureIDStr)

	// Check for multiple feature IDs
	if len(featureIDs) > 1 {
		violations = append(violations, Violation{
			Code:     ViolationCodeMultipleFeatureIDs,
			Severity: SeverityError,
			Message:  "Commit message must reference exactly one Feature ID.",
			Details: map[string]any{
				"feature_ids": featureIDs,
			},
		})
	}

	// Validate each feature ID
	for _, featureID := range featureIDs {
		// Check format (SCREAMING_SNAKE_CASE)
		if !isValidFeatureIDFormat(featureID) {
			violations = append(violations, Violation{
				Code:     ViolationCodeInvalidFeatureIDFormat,
				Severity: SeverityError,
				Message:  fmt.Sprintf("Feature ID %q does not match SCREAMING_SNAKE_CASE format.", featureID),
				Details: map[string]any{
					"feature_id": featureID,
				},
			})
			continue
		}

		// Check if feature ID exists in spec
		if !knownFeatures[featureID] {
			violations = append(violations, Violation{
				Code:     ViolationCodeFeatureIDNotInSpec,
				Severity: SeverityError,
				Message:  fmt.Sprintf("Feature ID %q is not defined in spec/features.yaml.", featureID),
				Details: map[string]any{
					"feature_id": featureID,
				},
			})
		}
	}

	// Validate summary
	summary := matches[3]
	if len(summary) > 72 {
		violations = append(violations, Violation{
			Code:     ViolationCodeSummaryTooLong,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Commit summary is too long: %d characters (max 72).", len(summary)),
			Details: map[string]any{
				"length": len(summary),
			},
		})
	}

	if strings.HasSuffix(summary, ".") {
		violations = append(violations, Violation{
			Code:     ViolationCodeSummaryHasTrailingPeriod,
			Severity: SeverityWarning,
			Message:  "Commit summary should not end with a period.",
			Details:  map[string]any{},
		})
	}

	// Check if summary starts with uppercase (after colon)
	if summary != "" && summary[0] >= 'A' && summary[0] <= 'Z' {
		violations = append(violations, Violation{
			Code:     ViolationCodeSummaryStartsWithUppercase,
			Severity: SeverityWarning,
			Message:  "Commit summary should start with a lowercase letter.",
			Details:  map[string]any{},
		})
	}

	return violations
}

// parseFeatureIDs parses feature IDs from a string like "CLI_PLAN" or "CLI_PLAN, CLI_DEPLOY".
func parseFeatureIDs(s string) []string {
	// Split by comma and trim whitespace
	parts := strings.Split(s, ",")
	var featureIDs []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			featureIDs = append(featureIDs, trimmed)
		}
	}
	return featureIDs
}

// isValidFeatureIDFormat checks if a string matches SCREAMING_SNAKE_CASE.
func isValidFeatureIDFormat(s string) bool {
	// SCREAMING_SNAKE_CASE: starts with uppercase letter, followed by uppercase letters, digits, and underscores
	pattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	return pattern.MatchString(s)
}

// getAllRules returns all validation rules.
func getAllRules() []Rule {
	return []Rule{
		{
			Code:        ViolationCodeMissingFeatureID,
			Description: "Commit message is missing a Feature ID in the required format.",
			Severity:    SeverityError,
		},
		{
			Code:        ViolationCodeMultipleFeatureIDs,
			Description: "Commit message references multiple Feature IDs; only one is allowed per commit.",
			Severity:    SeverityError,
		},
		{
			Code:        ViolationCodeInvalidFeatureIDFormat,
			Description: "Feature ID does not match SCREAMING_SNAKE_CASE format.",
			Severity:    SeverityError,
		},
		{
			Code:        ViolationCodeFeatureIDNotInSpec,
			Description: "Feature ID is not defined in spec/features.yaml.",
			Severity:    SeverityError,
		},
		{
			Code:        ViolationCodeSummaryTooLong,
			Description: "Commit summary exceeds 72 characters.",
			Severity:    SeverityWarning,
		},
		{
			Code:        ViolationCodeSummaryHasTrailingPeriod,
			Description: "Commit summary ends with a period.",
			Severity:    SeverityWarning,
		},
		{
			Code:        ViolationCodeSummaryStartsWithUppercase,
			Description: "Commit summary starts with an uppercase letter.",
			Severity:    SeverityWarning,
		},
		{
			Code:        ViolationCodeInvalidFormatGeneric,
			Description: "Commit message does not match required format.",
			Severity:    SeverityError,
		},
	}
}
