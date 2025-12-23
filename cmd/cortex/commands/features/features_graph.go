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

func NewFeaturesGraphCommand() *cobra.Command {
	var (
		featuresPath string
		dot          bool
	)

	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Validate and visualize the feature DAG",
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := features.LoadGraph(featuresPath)
			if err != nil {
				return fmt.Errorf("failed to load graph: %w", err)
			}

			if err := features.ValidateDAG(g); err != nil {
				return fmt.Errorf("feature DAG invalid: %w", err)
			}

			if dot {
				fmt.Println(features.ToDOT(g))
			} else {
				fmt.Printf("✓ Feature dependency graph is valid (acyclic)\n")
				fmt.Printf("  Total features: %d\n", len(g.Nodes))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&featuresPath, "features", "spec/features.yaml", "Path to features.yaml")
	cmd.Flags().BoolVar(&dot, "dot", false, "Output in DOT format")

	return cmd
}

// Use:   "features",
// Short: "Manage feature dependency graphs and documentation",
// Long:  "Tools for visualizing, analyzing, and documenting the feature graph defined in spec/features.yaml",
// }
