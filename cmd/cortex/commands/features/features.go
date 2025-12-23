// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package features contains Cobra subcommands for the Cortex CLI.
package features

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
