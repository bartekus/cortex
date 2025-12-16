// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestCortexHelp_Golden(t *testing.T) {
	// 1. Init root command
	cmd := NewRootCommand()

	// 2. Execute help
	output, err := executeCommandForGolden(cmd, "--help")
	if err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}

	// 3. Compare/Update golden
	goldenName := "cortex_help"
	if *updateGolden {
		writeGoldenFile(t, goldenName, output)
	}

	expected := readGoldenFile(t, goldenName)
	assert.Equal(t, expected, output, "help output does not match golden file")
}
