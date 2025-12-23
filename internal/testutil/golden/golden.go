package golden

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestdataDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(filename), "testdata")
}

var Update = flag.Bool("update", false, "update golden files")

func Read(t *testing.T, testdataDir, name string) string {
	t.Helper()
	safeName(t, name)

	path := filepath.Join(testdataDir, name+".golden")
	data, err := os.ReadFile(path) //nolint:gosec // testdata path controlled by test
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read golden %s: %v", path, err)
	}
	return string(data)
}

func Write(t *testing.T, testdataDir, name, content string) {
	t.Helper()
	safeName(t, name)

	if err := os.MkdirAll(testdataDir, 0o750); err != nil {
		t.Fatalf("mkdir testdata: %v", err)
	}
	path := filepath.Join(testdataDir, name+".golden")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write golden %s: %v", path, err)
	}
}

func safeName(t *testing.T, name string) {
	t.Helper()
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		t.Fatalf("invalid golden name %q", name)
	}
}
