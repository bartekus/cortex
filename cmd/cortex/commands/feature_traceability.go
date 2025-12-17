// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bartekus/cortex/internal/projectroot"
	"github.com/bartekus/cortex/internal/reports"
	"github.com/bartekus/cortex/internal/reports/featuretrace"
)

// Feature: PROVIDER_FRONTEND_GENERIC
// Docs: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md

// NewFeatureTraceabilityCommand returns the `cortex feature traceability` command.
func NewFeatureTraceabilityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feature traceability",
		Short: "Generate feature traceability report",
		Long:  "Generates a feature traceability report analyzing feature presence across spec, implementation, tests, and commits",
		RunE:  runFeatureTraceability,
	}

	// No flags for now (can add --root-dir later if needed)

	return cmd
}

// runFeatureTraceability executes the feature traceability command.
func runFeatureTraceability(cmd *cobra.Command, args []string) error {
	// 1. Get repository root
	repoPath, err := projectroot.Find(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	// 2. Scan repository for feature presence
	scanConfig := featuretrace.ScanConfig{
		RootDir: repoPath,
	}

	features, err := featuretrace.ScanFeaturePresence(scanConfig)
	if err != nil {
		return fmt.Errorf("scanning repository: %w", err)
	}

	// 3. Generate report using Phase 3.B generator
	report, err := featuretrace.GenerateFeatureTraceabilityReport(features)
	if err != nil {
		return fmt.Errorf("generating feature traceability report: %w", err)
	}

	// 4. Write report atomically
	reportPath := filepath.Join(repoPath, ".cortex", "reports", "feature-traceability.json")
	if err := reports.WriteJSONAtomic(reportPath, report); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	return nil
}
