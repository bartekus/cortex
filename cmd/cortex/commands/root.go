package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
	cmd.AddCommand(NewContextCommand())
	cmd.AddCommand(NewContextXrayCommand()) // Promoted to top-level
	cmd.AddCommand(NewCommitReportCommand())
	cmd.AddCommand(NewCommitSuggestCommand())
	cmd.AddCommand(NewFeatureTraceabilityCommand())
	cmd.AddCommand(NewGovCommand())
	cmd.AddCommand(NewStatusCommand())
	cmd.AddCommand(NewFeaturesCommand())
	cmd.AddCommand(GetRunCmd())

	return cmd
}
