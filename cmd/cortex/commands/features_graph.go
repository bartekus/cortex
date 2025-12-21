// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

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
