// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"github.com/bartekus/cortex/cmd/cortex/commands"
	"github.com/bartekus/cortex/cmd/cortex/internal/clierr"
)

// Feature: CLI_CONTRACT
// Spec: spec/cli/contract.md

func main() {
	if err := commands.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(clierr.ExitCodeOf(err))
	}
}
