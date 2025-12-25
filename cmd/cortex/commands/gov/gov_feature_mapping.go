// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package gov

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/bartekus/cortex/cmd/cortex/internal/clierr"
	"github.com/bartekus/cortex/internal/mapping"

	"github.com/spf13/cobra"
)

func NewGovFeatureMappingCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "feature-mapping",
		Short: "Validate feature/spec/code/test mapping",
		Long:  "Validates the Feature Mapping Invariant across specs, features.yaml, implementation code, and tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := mapping.DefaultOptions()

			report, err := mapping.Analyze(opts)
			if err != nil {
				// Internal error – use exit code 2.
				return clierr.New(2, fmt.Sprintf("governance feature mapping failed: %v", err))
			}

			switch format {
			case "json":
				if err := renderReportJSON(cmd, report); err != nil {
					// Rendering issues are internal errors.
					return clierr.New(2, fmt.Sprintf("render mapping report (json): %v", err))
				}
			case "text", "":
				if err := renderReportText(cmd, report); err != nil {
					// Rendering issues are internal errors.
					return clierr.New(2, fmt.Sprintf("render mapping report (text): %v", err))
				}
			default:
				return clierr.New(2, fmt.Sprintf("unsupported format %q (expected text or json)", format))
			}

			// After rendering, decide exit code based on violations.
			if len(report.Violations) > 0 {
				return clierr.New(1, fmt.Sprintf("feature mapping validation failed with %d violation(s)", len(report.Violations)))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")

	return cmd
}

func renderReportJSON(cmd *cobra.Command, report mapping.Report) error {
	// The mapping.Report implementation is responsible for sorting its slices.
	// This function must not introduce additional non-determinism.
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("encode mapping report as JSON: %w", err)
	}

	return nil
}

func renderReportText(cmd *cobra.Command, report mapping.Report) error {
	out := cmd.OutOrStdout()

	features := append([]mapping.FeatureMapping(nil), report.Features...)
	sort.Slice(features, func(i, j int) bool {
		return features[i].ID < features[j].ID
	})

	violations := append([]mapping.Violation(nil), report.Violations...)
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Code != violations[j].Code {
			return violations[i].Code < violations[j].Code
		}
		if violations[i].Feature != violations[j].Feature {
			return violations[i].Feature < violations[j].Feature
		}
		return violations[i].Path < violations[j].Path
	})

	// Summary line.
	if _, err := fmt.Fprintf(out, "Feature mapping: %d features, %d violations\n", len(features), len(violations)); err != nil {
		return fmt.Errorf("write mapping summary: %w", err)
	}

	// Fast path: no violations – clearly state success.
	if len(violations) == 0 {
		if _, err := fmt.Fprintln(out, "\nNo violations found. Feature Mapping Invariant holds."); err != nil {
			return fmt.Errorf("write success message: %w", err)
		}
		return nil
	}

	// Optional: feature status summary for quick governance overview.
	statusCounts := make(map[mapping.FeatureStatus]int)
	for _, f := range features {
		statusCounts[f.Status]++
	}

	if len(statusCounts) > 0 {
		if _, err := fmt.Fprintln(out, "\nFeature status summary:"); err != nil {
			return fmt.Errorf("write status summary header: %w", err)
		}
		orderedStatuses := []mapping.FeatureStatus{
			mapping.FeatureStatusOK,
			mapping.FeatureStatusMissingImpl,
			mapping.FeatureStatusMissingTests,
			mapping.FeatureStatusIncomplete,
			mapping.FeatureStatusSpecOnly,
			mapping.FeatureStatusImplementationOnly,
			mapping.FeatureStatusUnmapped,
		}
		for _, st := range orderedStatuses {
			if count, ok := statusCounts[st]; ok && count > 0 {
				if _, err := fmt.Fprintf(out, "  %s: %d\n", st, count); err != nil {
					return fmt.Errorf("write status line: %w", err)
				}
			}
		}
	}

	// Violations grouped by code.
	if _, err := fmt.Fprintln(out, "\nViolations by code:"); err != nil {
		return fmt.Errorf("write violations header: %w", err)
	}

	currentCode := mapping.ReportCode("")
	for _, v := range violations {
		if v.Code != currentCode {
			currentCode = v.Code
			if _, err := fmt.Fprintf(out, "\n  %s:\n", currentCode); err != nil {
				return fmt.Errorf("write violation code header: %w", err)
			}
		}

		path := v.Path
		if path == "" {
			path = "<no path>"
		}
		feature := v.Feature
		if feature == "" {
			feature = "<no feature>"
		}
		if _, err := fmt.Fprintf(out, "    - %s (%s): %s\n", feature, path, v.Detail); err != nil {
			return fmt.Errorf("write violation line: %w", err)
		}
	}

	return nil
}
