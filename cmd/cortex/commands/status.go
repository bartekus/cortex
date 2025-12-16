// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: GOV_STATUS_ROADMAP
// Spec: spec/commands/status-roadmap.md

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bartekus/cortex/internal/projectroot"
	"github.com/bartekus/cortex/internal/roadmap"
)

const (
	defaultFeaturesPath = "spec/features.yaml"
	defaultOutputPath   = "docs/engine/status/feature-completion-analysis.md"
)

// NewStatusCommand returns the `stagecraft status` command with subcommands
// such as `stagecraft status roadmap`.
func NewStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Stagecraft project status and roadmap analysis",
	}

	cmd.AddCommand(newStatusRoadmapCommand())

	return cmd
}

func newStatusRoadmapCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roadmap",
		Short: "Generate phase-level feature completion analysis from spec/features.yaml",
		Long: `Generate a deterministic phase-level feature completion analysis document
based on spec/features.yaml and write it to docs/engine/status/feature-completion-analysis.md.

This command is part of GOV_STATUS_ROADMAP and is used by governance tooling.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			featuresPath, err := cmd.Flags().GetString("features")
			if err != nil {
				return newExitError(2, fmt.Sprintf("status roadmap: get features flag: %v", err))
			}

			outputPath, err := cmd.Flags().GetString("output")
			if err != nil {
				return newExitError(2, fmt.Sprintf("status roadmap: get output flag: %v", err))
			}

			// Resolve paths relative to repository root
			repoRoot, err := projectroot.Find(".")
			if err != nil {
				return newExitError(2, fmt.Sprintf("status roadmap: finding repo root: %v", err))
			}

			if !filepath.IsAbs(featuresPath) {
				featuresPath = filepath.Join(repoRoot, featuresPath)
			}
			if !filepath.IsAbs(outputPath) {
				outputPath = filepath.Join(repoRoot, outputPath)
			}

			phases, err := roadmap.DetectPhases(featuresPath)
			if err != nil {
				// Check if it's a file not found error
				if os.IsNotExist(err) {
					return newExitError(1, fmt.Sprintf("status roadmap: features file not found: %s", featuresPath))
				}
				// YAML parsing errors are validation errors (exit code 1)
				return newExitError(1, fmt.Sprintf("status roadmap: detect phases: %v", err))
			}

			stats := roadmap.CalculateStats(phases)
			blockers := roadmap.IdentifyBlockers(phases)

			markdown := roadmap.GenerateMarkdown(stats, blockers)

			// Ensure output directory exists
			outputDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputDir, 0o750); err != nil {
				return newExitError(1, fmt.Sprintf("status roadmap: create output directory: %v", err))
			}

			if err := os.WriteFile(outputPath, []byte(markdown), 0o600); err != nil {
				return newExitError(1, fmt.Sprintf("status roadmap: write output %q: %v", outputPath, err))
			}

			return nil
		},
	}

	cmd.Flags().String(
		"features",
		defaultFeaturesPath,
		"path to spec/features.yaml",
	)
	cmd.Flags().String(
		"output",
		defaultOutputPath,
		"path to write the generated feature completion analysis",
	)

	return cmd
}
