package commands

import (
	"bytes"
	"strings"
	"testing"
)

// Feature: CLI_COMMAND_FEATURES
// Spec: spec/cli/features.md

func TestCLICommandFeatures(t *testing.T) {
	cmd := NewRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"features", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("features command failed: %v", err)
	}

	out := b.String()

	// Assert fundamental command presence
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage info in features help")
	}

	// Assert subcommands expected by contract
	expectedSubs := []string{
		"overview",
		"graph",
		"impact",
	}

	for _, sub := range expectedSubs {
		if !strings.Contains(out, sub) {
			t.Errorf("expected subcommand %q in features help output", sub)
		}
	}
}
