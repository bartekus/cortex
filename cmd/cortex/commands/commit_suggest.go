// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bartekus/cortex/internal/projectroot"

	"github.com/spf13/cobra"

	"github.com/bartekus/cortex/internal/reports/commithealth"
	"github.com/bartekus/cortex/internal/reports/featuretrace"
	"github.com/bartekus/cortex/internal/reports/suggestions"
)

// Feature: CLI_COMMAND_COMMIT
// Spec: spec/cli/commit.md

// NewCommitSuggestCommand returns the `cortex commit suggest` command.
func NewCommitSuggestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit suggest",
		Short: "Generate commit discipline suggestions",
		Long:  "Reads commit health and feature traceability reports and generates actionable suggestions for improving commit discipline",
		RunE:  runCommitSuggest,
	}

	// Flags in alphabetical order for deterministic help output
	cmd.Flags().String("format", "text", "Output format: text (default) or json")
	cmd.Flags().String("severity", "info", "Minimum severity to include: info, warning, or error (default: info)")
	cmd.Flags().Int("max-suggestions", 10, "Maximum number of suggestions to display (default: 10, 0 = unlimited)")

	return cmd
}

// runCommitSuggest executes the commit suggest command.
func runCommitSuggest(cmd *cobra.Command, args []string) error {
	// 1. Get repository root
	repoPath, err := projectroot.Find(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	// 2. Read commit health report
	commitReportPath := filepath.Join(repoPath, ".cortex", "reports", "commit-health.json")
	commitReportData, err := os.ReadFile(commitReportPath) //nolint:gosec // G304: path is validated from user input
	if err != nil {
		return fmt.Errorf("reading commit health report: %w", err)
	}

	var commitReport commithealth.Report
	if err := json.Unmarshal(commitReportData, &commitReport); err != nil {
		return fmt.Errorf("parsing commit health report: %w", err)
	}

	// 3. Read feature traceability report
	featureReportPath := filepath.Join(repoPath, ".cortex", "reports", "feature-traceability.json")
	featureReportData, err := os.ReadFile(featureReportPath) //nolint:gosec // G304: path is validated from user input
	if err != nil {
		return fmt.Errorf("reading feature traceability report: %w", err)
	}

	var featureReport featuretrace.Report
	if err := json.Unmarshal(featureReportData, &featureReport); err != nil {
		return fmt.Errorf("parsing feature traceability report: %w", err)
	}

	// 4. Generate suggestions
	rawSuggestions, err := suggestions.GenerateSuggestions(&commitReport, &featureReport)
	if err != nil {
		return fmt.Errorf("generating suggestions: %w", err)
	}

	// 5. Get flags
	formatFlag, _ := cmd.Flags().GetString("format")
	severityFlag, _ := cmd.Flags().GetString("severity")
	maxSuggestionsFlag, _ := cmd.Flags().GetInt("max-suggestions")

	// 6. Parse severity
	minSeverity, err := parseSeverity(severityFlag)
	if err != nil {
		return fmt.Errorf("invalid severity: %w", err)
	}

	// 7. Prioritize and filter suggestions
	prioritized := suggestions.PrioritizeSuggestions(rawSuggestions)
	filtered := suggestions.FilterSuggestions(prioritized, minSeverity, maxSuggestionsFlag)

	// 8. Format and output
	switch formatFlag {
	case "text":
		out := suggestions.FormatSuggestionsText(filtered)
		if _, err := cmd.OutOrStdout().Write([]byte(out)); err != nil {
			return fmt.Errorf("writing text output: %w", err)
		}
		return nil

	case "json":
		// Build report structure for JSON output
		report := suggestions.BuildReport(filtered)
		jsonData, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		jsonData = append(jsonData, '\n')
		if _, err := cmd.OutOrStdout().Write(jsonData); err != nil {
			return fmt.Errorf("writing JSON output: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", formatFlag)
	}
}

// parseSeverity parses a severity string into a suggestions.Severity.
func parseSeverity(s string) (suggestions.Severity, error) {
	switch s {
	case "error":
		return suggestions.SeverityError, nil
	case "warning":
		return suggestions.SeverityWarning, nil
	case "info":
		return suggestions.SeverityInfo, nil
	default:
		return "", fmt.Errorf("unknown severity: %s (must be 'error', 'warning', or 'info')", s)
	}
}
