// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package commands contains Cobra subcommands for the Cortex CLI.
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bartekus/cortex/pkg/introspect"
	"github.com/spf13/cobra"
)

func NewGovCLIDumpJSONCommand() *cobra.Command {
	var out string

	cmd := &cobra.Command{
		Use:   "cli-dump-json",
		Short: "Dump the CLI command tree (commands + flags) to JSON for spec-vs-cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == "" {
				return fmt.Errorf("--out is required")
			}
			root := cmd.Root()
			if root == nil {
				return fmt.Errorf("failed to resolve root command")
			}
			tree := introspect.Introspect(root)

			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return fmt.Errorf("failed to create output dir: %w", err)
			}

			f, err := os.Create(out)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer func() { _ = f.Close() }()

			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			if err := enc.Encode(tree); err != nil {
				return fmt.Errorf("failed to encode json: %w", err)
			}

			fmt.Printf("✓ Wrote CLI JSON to %s\n", out)
			return nil
		},
	}

	cmd.Flags().StringVar(&out, "out", ".cortex/data/cli.json", "Output path for CLI JSON")
	return cmd
}
