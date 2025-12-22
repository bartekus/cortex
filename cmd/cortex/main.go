// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"github.com/bartekus/cortex/cmd/cortex/commands"
)

// Feature: CLI_CONTRACT
// Spec: spec/cli/contract.md

func main() {
	if err := commands.NewRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
