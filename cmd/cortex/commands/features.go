// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"github.com/spf13/cobra"
)

// Feature: CLI_COMMAND_FEATURES
// Spec: spec/cli/features.md

func NewFeaturesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "features",
		Short: "Manage feature dependency graphs and documentation",
		Long:  "Tools for visualizing, analyzing, and documenting the feature graph defined in spec/features.yaml",
	}

	cmd.AddCommand(NewFeaturesGraphCommand())
	cmd.AddCommand(NewFeaturesImpactCommand())
	cmd.AddCommand(NewFeaturesOverviewCommand())

	return cmd
}
