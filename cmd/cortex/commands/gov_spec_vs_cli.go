// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bartekus/cortex/internal/specschema"
	"github.com/bartekus/cortex/internal/specvscli"
	"github.com/bartekus/cortex/pkg/introspect"
	"github.com/spf13/cobra"
)

// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md

func NewGovSpecVsCLICommand() *cobra.Command {
	var (
		specPath   string
		binaryPath string
		strict     bool
	)

	cmd := &cobra.Command{
		Use:   "spec-vs-cli",
		Short: "Validate alignment between CLI help output and Spec flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			specs, err := specschema.LoadAllSpecs(specPath)
			if err != nil {
				return fmt.Errorf("failed to load specs: %w", err)
			}

			// Load CLI definition from JSON file
			if binaryPath == "" {
				return fmt.Errorf("--binary-json is required")
			}

			f, err := os.Open(binaryPath)
			if err != nil {
				return fmt.Errorf("failed to open binary json file: %w", err)
			}
			defer func() { _ = f.Close() }()

			var cliCommands []introspect.CommandInfo
			if err := json.NewDecoder(f).Decode(&cliCommands); err != nil {
				return fmt.Errorf("failed to decode binary json: %w", err)
			}

			// Use CompareAllCommands from specvscli
			results := specvscli.CompareAllCommands(specs, cliCommands)

			// Report results
			hasErrors := false
			hasWarnings := false

			for _, result := range results {
				if len(result.Errors) > 0 {
					hasErrors = true
					fmt.Printf("ERROR: Command %q:\n", result.CommandName)
					for _, err := range result.Errors {
						fmt.Printf("  - %s\n", err)
					}
				}
				if len(result.Warnings) > 0 {
					hasWarnings = true
					fmt.Printf("WARNING: Command %q:\n", result.CommandName)
					for _, warn := range result.Warnings {
						fmt.Printf("  - %s\n", warn)
					}
				}
			}

			if hasErrors {
				return fmt.Errorf("CLI alignment check failed")
			}

			if hasWarnings {
				fmt.Printf("\n⚠ Flag alignment warnings (non-blocking)\n")
			} else {
				fmt.Println("✓ CLI matches Spec")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&specPath, "spec-root", "spec", "Root directory containing spec files")
	cmd.Flags().StringVar(&binaryPath, "binary-json", "", "Path to JSON output from cli-dump-json")
	cmd.Flags().BoolVar(&strict, "strict", false, "Fail on warnings (not implemented yet)")

	return cmd
}
