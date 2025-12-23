// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CLI_COMMAND_STATUS
// Spec: spec/cli/status.md

package reports

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bartekus/cortex/cmd/cortex/internal/clierr"
	"github.com/spf13/cobra"

	"github.com/bartekus/cortex/internal/projectroot"
	"github.com/bartekus/cortex/internal/roadmap"
)

const (
	defaultFeaturesPath = "spec/features.yaml"
	defaultOutputPath   = "docs/__generated__/feature-completion-analysis.md"
)

func NewStatusRoadmapCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status-roadmap",
		Short: "Generate phase-level feature completion analysis from spec/features.yaml",
		Long: `Generate a deterministic phase-level feature completion analysis document
based on spec/features.yaml and write it to docs/__generated__/feature-completion-analysis.md.

This command is part of CLI_COMMAND_STATUS and is used by governance tooling.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			featuresPath, err := cmd.Flags().GetString("features")
			if err != nil {
				return clierr.New(2, fmt.Sprintf("status roadmap: get features flag: %v", err))
			}

			outputPath, err := cmd.Flags().GetString("output")
			if err != nil {
				return clierr.New(2, fmt.Sprintf("status roadmap: get output flag: %v", err))
			}

			// Resolve paths relative to repository root
			repoRoot, err := projectroot.Find(".")
			if err != nil {
				return clierr.New(2, fmt.Sprintf("status roadmap: finding repo root: %v", err))
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
					return clierr.New(1, fmt.Sprintf("status roadmap: features file not found: %s", featuresPath))
				}
				// YAML parsing errors are validation errors (exit code 1)
				return clierr.New(1, fmt.Sprintf("status roadmap: detect phases: %v", err))
			}

			stats := roadmap.CalculateStats(phases)
			blockers := roadmap.IdentifyBlockers(phases)

			markdown := roadmap.GenerateMarkdown(stats, blockers)

			// Ensure output directory exists
			outputDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputDir, 0o750); err != nil {
				return clierr.New(1, fmt.Sprintf("status roadmap: create output directory: %v", err))
			}

			if err := os.WriteFile(outputPath, []byte(markdown), 0o600); err != nil {
				return clierr.New(1, fmt.Sprintf("status roadmap: write output %q: %v", outputPath, err))
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
