// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package reports contains Cobra subcommands for the Cortex CLI.
package reports

import (
	"github.com/spf13/cobra"
)

// NewReportsCommand returns the `cortex git` command.
func NewReportsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reports",
		Short: "Report generators for Cortex",
		Long:  "Report commands for Cortex's commit discipline & health, feature traceability and status roadmap analysis",
	}

	cmd.AddCommand(NewCommitReportCommand())
	cmd.AddCommand(NewCommitSuggestCommand())
	cmd.AddCommand(NewFeatureTraceabilityCommand())
	cmd.AddCommand(NewStatusRoadmapCommand())

	return cmd
}
