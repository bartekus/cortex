// SPDX-License-Identifier: AGPL-3.0-or-later
package projection

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "out", "file.txt")
	content := []byte("hello world")

	if err := AtomicWrite(target, content); err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(got) != string(content) {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	keys := SortedKeys(m)
	if len(keys) != 3 {
		t.Fatalf("got %d keys, want 3", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("got %v, want [a b c]", keys)
	}
}
