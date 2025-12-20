// SPDX-License-Identifier: AGPL-3.0-or-later
package builder_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartekus/cortex/internal/builder"
	"github.com/bartekus/cortex/internal/xray"
)

func TestBuildContext_Golden(t *testing.T) {
	// Setup temp workspace
	tempDir := t.TempDir()

	// Create a mock XRAY index
	index := &xray.Index{
		SchemaVersion: "1.0.0",
		Root:          tempDir,
		Files: []xray.FileNode{
			{Path: "B.txt", Hash: "sha256:bbb", Size: 20},
			{Path: "A.txt", Hash: "sha256:aaa", Size: 10}, // Unsorted to test sorting
		},
	}

	// Create dummy files (required for chunking)
	if err := os.WriteFile(filepath.Join(tempDir, "A.txt"), []byte("content A"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "B.txt"), []byte("content B"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create expected structure in tempDir (simulating repo root)
	if err := builder.BuildContext(tempDir, index); err != nil {
		t.Fatalf("BuildContext failed: %v", err)
	}

	// Verify meta.json
	metaPath := filepath.Join(tempDir, ".cortex", "meta.json")
	assertFileExists(t, metaPath)
	// Optionally check content

	// Verify manifest.json
	manifestPath := filepath.Join(tempDir, ".cortex", "files", "manifest.json")
	assertFileExists(t, manifestPath)

	bytes, _ := os.ReadFile(manifestPath)
	var manifest []builder.ManifestEntry
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	if len(manifest) != 2 {
		t.Errorf("Expected 2 manifest entries, got %d", len(manifest))
	}
	if manifest[0].Path != "A.txt" {
		t.Errorf("Manifest not sorted. First entry is %s", manifest[0].Path)
	}

	// Verify digest.txt
	digestPath := filepath.Join(tempDir, ".cortex", "digest.txt")
	assertFileExists(t, digestPath)

	dBytes, _ := os.ReadFile(digestPath)
	if len(dBytes) < 64 {
		t.Errorf("Digest seems too short: %s", string(dBytes))
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file at %s, but missing", path)
	}
}
