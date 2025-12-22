package commands

import (
	"bytes"
	"strings"
	"testing"
)

// Feature: CLI_CONTRACT
// Spec: spec/cli/contract.md

func TestCLIContract(t *testing.T) {
	cmd := NewRootCommand()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command failed: %v", err)
	}

	out := b.String()

	// Assert top-level commands that are part of the core contract
	requiredCommands := []string{
		"gov",
		"context",
		"features",
		"run",
		"version",
	}

	for _, c := range requiredCommands {
		if !strings.Contains(out, c) {
			t.Errorf("expected top-level command %q in root help", c)
		}
	}
}
