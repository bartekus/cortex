package commands

import (
	"fmt"
	"os"

	"github.com/bartekus/cortex/pkg/gov"
	"github.com/spf13/cobra"
)

// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md

func NewGovValidateCommand() *cobra.Command {
	var (
		registryPath string
		rootDir      string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the feature registry and spec integrity",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Parse and Validate Registry
			reg, err := gov.LoadRegistry(registryPath)
			if err != nil {
				return fmt.Errorf("failed to load registry from %s: %w", registryPath, err)
			}

			if err := reg.Validate(); err != nil {
				return fmt.Errorf("registry validation failed: %w", err)
			}
			fmt.Println("✓ Registry structure valid (governance + implementation)")

			// 2. Traceability Checks
			if err := reg.ValidateTraceability(rootDir); err != nil {
				return fmt.Errorf("traceability check failed: %w", err)
			}
			fmt.Println("✓ Traceability checks passed (spec files exist and reference IDs)")

			// 3. Dependency Graph
			if err := reg.ValidateDependencies(); err != nil {
				return fmt.Errorf("dependency graph check failed: %w", err)
			}
			fmt.Println("✓ Dependency graph valid (all dependencies exist, no cycles)")

			return nil
		},
	}

	// Default to assume running from repo root
	cwd, _ := os.Getwd()
	cmd.Flags().StringVar(&registryPath, "registry", "spec/features.yaml", "Path to features.yaml")
	cmd.Flags().StringVar(&rootDir, "root", cwd, "Root directory of the repository")

	return cmd
}
