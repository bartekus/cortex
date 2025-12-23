// SPDX-License-Identifier: AGPL-3.0-or-later

package gov

import (
	"fmt"

	"github.com/bartekus/cortex/internal/specschema"
	"github.com/spf13/cobra"
)

// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md

func NewGovSpecValidateCommand() *cobra.Command {
	var (
		rootPath       string
		featuresPath   string
		checkIntegrity bool
	)

	cmd := &cobra.Command{
		Use:   "spec-validate",
		Short: "Validate spec file frontmatter",
		RunE: func(cmd *cobra.Command, args []string) error {
			specs, err := specschema.LoadAllSpecs(rootPath)
			if err != nil {
				return fmt.Errorf("failed to load specs: %w", err)
			}

			if len(specs) == 0 {
				fmt.Printf("warning: no spec files found in %s\n", rootPath)
				return nil
			}

			if err := specschema.ValidateAll(specs); err != nil {
				return fmt.Errorf("spec validation failed: %w", err)
			}

			if checkIntegrity {
				if err := specschema.ValidateSpecIntegrity(featuresPath, rootPath); err != nil {
					return fmt.Errorf("spec integrity validation failed: %w", err)
				}
				fmt.Printf("✓ Spec integrity check passed\n")
			}

			fmt.Printf("✓ Validated %d spec file(s)\n", len(specs))
			return nil
		},
	}

	cmd.Flags().StringVar(&rootPath, "root", "spec", "Root directory containing spec files")
	cmd.Flags().StringVar(&featuresPath, "features", "spec/features.yaml", "Path to features.yaml")
	cmd.Flags().BoolVar(&checkIntegrity, "check-integrity", false, "Also validate features.yaml ↔ spec file integrity")

	return cmd
}
