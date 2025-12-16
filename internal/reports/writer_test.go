// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_FRONTEND_GENERIC
// Docs: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md
package reports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteJSONAtomic_CreatesFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "report.json")

	testData := map[string]string{
		"key": "value",
	}

	err := WriteJSONAtomic(targetPath, testData)
	if err != nil {
		t.Fatalf("WriteJSONAtomic failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("target file does not exist: %v", err)
	}

	// Verify temp file is removed
	tmpPath := targetPath + ".tmp"
	if _, err := os.Stat(tmpPath); err == nil {
		t.Error("temporary file should be removed after rename")
	}
}

func TestWriteJSONAtomic_ValidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "report.json")

	testData := map[string]interface{}{
		"schema_version": "1.0",
		"summary": map[string]int{
			"total_commits": 2,
		},
	}

	err := WriteJSONAtomic(targetPath, testData)
	if err != nil {
		t.Fatalf("WriteJSONAtomic failed: %v", err)
	}

	// Read and verify JSON
	data, err := os.ReadFile(targetPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("written JSON is invalid: %v", err)
	}

	if unmarshaled["schema_version"] != "1.0" {
		t.Errorf("expected schema_version=1.0, got %v", unmarshaled["schema_version"])
	}
}

func TestWriteJSONAtomic_CreatesParentDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "reports", "subdir", "report.json")

	testData := map[string]string{
		"key": "value",
	}

	err := WriteJSONAtomic(targetPath, testData)
	if err != nil {
		t.Fatalf("WriteJSONAtomic failed: %v", err)
	}

	// Verify file exists in nested directory
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("target file does not exist: %v", err)
	}
}

func TestWriteJSONAtomic_AtomicWrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "report.json")

	// Write first version
	testData1 := map[string]string{"version": "1"}
	err := WriteJSONAtomic(targetPath, testData1)
	if err != nil {
		t.Fatalf("WriteJSONAtomic failed: %v", err)
	}

	// Verify first version
	data1, _ := os.ReadFile(targetPath) //nolint:gosec // G304: test file path
	var v1 map[string]string
	_ = json.Unmarshal(data1, &v1)
	if v1["version"] != "1" {
		t.Errorf("expected version=1, got %s", v1["version"])
	}

	// Write second version
	testData2 := map[string]string{"version": "2"}
	err = WriteJSONAtomic(targetPath, testData2)
	if err != nil {
		t.Fatalf("WriteJSONAtomic failed: %v", err)
	}

	// Verify second version (should be complete, not partial)
	data2, _ := os.ReadFile(targetPath) //nolint:gosec // G304: test file path
	var v2 map[string]string
	_ = json.Unmarshal(data2, &v2)
	if v2["version"] != "2" {
		t.Errorf("expected version=2, got %s", v2["version"])
	}
}
