package commands

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/bartekus/cortex/pkg/gov"
	"github.com/spf13/cobra"
)

func NewGovDriftCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drift",
		Short: "Detect drift between implementation and fixtures",
	}

	cmd.AddCommand(newDriftHelpCommand())
	cmd.AddCommand(newDriftXrayCommand())

	return cmd
}

func newDriftHelpCommand() *cobra.Command {
	var (
		binaryPath  string
		fixturePath string
	)

	cmd := &cobra.Command{
		Use:   "help",
		Short: "Check for CLI help output drift",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Run the binary to get help output
			c := exec.Command(binaryPath, "--help")
			var out bytes.Buffer
			c.Stdout = &out
			c.Stderr = &out // Capture stderr too just in case
			err := c.Run()
			if err != nil {
				return fmt.Errorf("failed to run %s --help: %w\nOutput:\n%s", binaryPath, err, out.String())
			}

			if err := gov.CompareHelp(out.String(), fixturePath); err != nil {
				return err
			}

			fmt.Println("✓ CLI help matches fixture")
			return nil
		},
	}

	cmd.Flags().StringVar(&binaryPath, "binary", "bin/cortex", "Path to cortex binary")
	cmd.Flags().StringVar(&fixturePath, "fixture", "spec/fixtures/cli/help.sample.txt", "Path to help fixture")

	return cmd
}

func newDriftXrayCommand() *cobra.Command {
	var fixturePath string

	cmd := &cobra.Command{
		Use:   "xray",
		Short: "Check for XRAY index fixture drift",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := gov.CheckXrayDrift(fixturePath); err != nil {
				return err
			}
			fmt.Println("✓ XRAY index fixture is valid")
			return nil
		},
	}

	cmd.Flags().StringVar(&fixturePath, "fixture", "spec/fixtures/xray/index.sample.json", "Path to XRAY index fixture")

	return cmd
}
