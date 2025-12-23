// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import (
	"fmt"

	"github.com/bartekus/cortex/internal/docs"
	"github.com/spf13/cobra"
)

// Feature: CLI_COMMAND_FEATURES
// Spec: spec/cli/features.md

func NewFeaturesOverviewCommand() *cobra.Command {
	var (
		featuresPath string
		specRoot     string
		outPath      string
	)

	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Generate feature overview documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := docs.GenerateFeatureOverview(featuresPath, specRoot, outPath); err != nil {
				return fmt.Errorf("failed to generate feature overview: %w", err)
			}

			fmt.Printf("✓ Generated feature overview at %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&featuresPath, "features", "spec/features.yaml", "Path to features.yaml")
	cmd.Flags().StringVar(&specRoot, "spec-root", "spec", "Root directory containing spec files")
	cmd.Flags().StringVar(&outPath, "out", "docs/__generated__/features-overview.md", "Output path for overview document")

	return cmd
}
