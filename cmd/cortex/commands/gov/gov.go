// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package gov contains Cobra subcommands for the Cortex CLI.
package gov

import (
	"github.com/spf13/cobra"
)

// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md

// NewGovCommand returns the `cortex gov` command.
func NewGovCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gov",
		Short: "Governance checks for Cortex",
		Long:  "Governance commands for validating Cortex's spec, feature, and code alignment",
	}

	cmd.AddCommand(NewGovFeatureMappingCommand())
	cmd.AddCommand(NewGovSpecValidateCommand())
	cmd.AddCommand(NewGovCLIDumpJSONCommand())
	cmd.AddCommand(NewGovSpecVsCLICommand())
	cmd.AddCommand(NewGovValidateCommand())
	cmd.AddCommand(NewGovDriftCommand())

	return cmd
}
