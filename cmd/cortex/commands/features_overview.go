// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"fmt"

	"github.com/bartekus/cortex/internal/docs"
	"github.com/spf13/cobra"
)

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
	cmd.Flags().StringVar(&outPath, "out", "docs/features/OVERVIEW.md", "Output path for overview document")

	return cmd
}
