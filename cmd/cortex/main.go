// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bartekus/cortex/cmd/cortex/commands"
)

// Feature: CORTEX_CLI
// Spec: ai.agent/cortex/README.md (Phase 2)

func main() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

// NewRootCommand constructs the Cortex root Cobra command.
func NewRootCommand() *cobra.Command {
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
	cmd.AddCommand(commands.NewContextCommand())
	cmd.AddCommand(commands.NewContextXrayCommand()) // Promoted to top-level
	cmd.AddCommand(commands.NewCommitReportCommand())
	cmd.AddCommand(commands.NewCommitSuggestCommand())
	cmd.AddCommand(commands.NewFeatureTraceabilityCommand())
	cmd.AddCommand(commands.NewGovCommand())
	cmd.AddCommand(commands.NewStatusCommand())
	cmd.AddCommand(commands.NewFeaturesCommand())
	cmd.AddCommand(commands.GetRunCmd())

	return cmd
}
