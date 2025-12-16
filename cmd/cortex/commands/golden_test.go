// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// updateGolden is a flag to update golden files during development.
// Usage: go test -update ./ai.agent/cmd/cortex/commands
var updateGolden = flag.Bool("update", false, "update golden files")

// readGoldenFile reads a golden file, or returns empty string if it doesn't exist.
func readGoldenFile(t *testing.T, name string) string {
	t.Helper()

	// Defensive: avoid path traversal or separators in golden names.
	if strings.Contains(name, "..") || strings.ContainsRune(name, os.PathSeparator) {
		t.Fatalf("invalid golden file name %q", name)
	}

	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	path := filepath.Join(testDir, "testdata", name+".golden")

	//nolint:gosec // G304: golden path is safe
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	return string(data)
}

// writeGoldenFile writes a golden file.
func writeGoldenFile(t *testing.T, name, content string) {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	dir := filepath.Join(testDir, "testdata")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("failed to create testdata directory: %v", err)
	}
	path := filepath.Join(dir, name+".golden")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write golden file %s: %v", path, err)
	}
}

// executeCommandForGolden executes a command and returns its output for golden file comparison.
func executeCommandForGolden(cmd *cobra.Command, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}
