// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"fmt"
	"os"

	"github.com/bartekus/cortex/cmd/cortex/commands/reports"
	"github.com/spf13/cobra"

	"github.com/bartekus/cortex/cmd/cortex/commands/context"
	"github.com/bartekus/cortex/cmd/cortex/commands/features"
	"github.com/bartekus/cortex/cmd/cortex/commands/gov"
)

// NewRootCmd constructs the Cortex root Cobra command.
func NewRootCmd() *cobra.Command {
	version := os.Getenv("CORTEX_VERSION")
	if version == "" {
		version = "0.0.0-dev"
	}

	cmd := &cobra.Command{
		Use:           "cortex",
		Short:         "Cortex - Developer & Governance Tooling for Cortex",
		Long:          "Cortex provides repository scanning, governance checks, and AI context generation tools.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	cmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")

	// Version command
	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of Cortex",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cortex version %s\n", version)
		},
	})

	// Register existing context commands
	// Note: We register NewContextCommand which provides subcommands like build, docs, xray.
	cmd.AddCommand(context.NewContextCommand())
	cmd.AddCommand(features.NewFeaturesCommand())
	cmd.AddCommand(reports.NewReportsCommand())
	cmd.AddCommand(gov.NewGovCommand())
	cmd.AddCommand(GetRunCmd())

	return cmd
}
