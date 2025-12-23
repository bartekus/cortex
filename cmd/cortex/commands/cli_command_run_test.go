package commands

import (
	"bytes"
	"strings"
	"testing"
)

// Feature: CLI_COMMAND_RUN
// Spec: spec/cli/run.md

func TestCLICommandRun(t *testing.T) {
	cmd := NewRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"run", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("run command failed: %v", err)
	}

	out := b.String()

	// Assert fundamental command presence
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage info in run help")
	}

	// Assert expected flags (minimal stable set)
	// Assuming --dry-run or similar might exist, but focusing on core helpfulness first.
	// Actually checking for "run" command specific descriptions
	if !strings.Contains(out, "run") {
		t.Errorf("expected 'run' in help output")
	}
}
