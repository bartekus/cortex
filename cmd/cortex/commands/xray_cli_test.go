package commands

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bartekus/cortex/internal/projectroot"
)

// Feature: XRAY_CLI
// Spec: spec/xray/cli.md

func TestXRayCLI(t *testing.T) {
	root, err := projectroot.Find(".")
	if err != nil {
		t.Fatalf("failed to get project root: %v", err)
	}

	// Expect xray binary to be in bin/xray (standard local dev layout)
	xrayPath := filepath.Join(root, "bin", "xray")

	cmd := exec.Command(xrayPath, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("output: %s", string(out))
		t.Fatalf("failed to execute xray --help: %v. Make sure 'make repo' or 'make build' has been run.", err)
	}

	sOut := string(out)
	if !strings.Contains(sOut, "xray") {
		t.Errorf("expected 'xray' in help output")
	}
	if !strings.Contains(sOut, "scan") {
		t.Errorf("expected 'scan' subcommand in help output")
	}
}
