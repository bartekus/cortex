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

	"github.com/bartekus/cortex/internal/features"
	"github.com/spf13/cobra"
)

// Feature: CLI_COMMAND_FEATURES
// Spec: spec/cli/features.md

func NewFeaturesImpactCommand() *cobra.Command {
	var (
		featuresPath string
		featureID    string
	)

	cmd := &cobra.Command{
		Use:   "impact [feature-id]",
		Short: "Analyze downstream impact of a feature",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				featureID = args[0]
			}
			if featureID == "" {
				return fmt.Errorf("feature-id is required")
			}

			g, err := features.LoadGraph(featuresPath)
			if err != nil {
				return fmt.Errorf("failed to load graph: %w", err)
			}

			impacted := features.Impact(g, featureID)
			if len(impacted) == 0 {
				fmt.Printf("No features depend on %s\n", featureID)
			} else {
				fmt.Printf("Features that depend on %s:\n", featureID)
				for _, id := range impacted {
					fmt.Printf("  - %s\n", id)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&featuresPath, "features", "spec/features.yaml", "Path to features.yaml")
	cmd.Flags().StringVar(&featureID, "feature", "", "Feature ID (deprecated: use arg)")

	return cmd
}
