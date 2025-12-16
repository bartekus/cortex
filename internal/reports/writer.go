// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package reports provides utilities for writing report files.
//
// Feature: PROVIDER_FRONTEND_GENERIC
// Docs: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md
package reports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteJSONAtomic writes a value to a JSON file atomically.
// It writes to a temporary file first, then renames it to the target path.
// This ensures the target file is either fully written or not present at all.
func WriteJSONAtomic(path string, v any) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil { //nolint:gosec // G301: output directory needs write permissions
		return fmt.Errorf("creating directory: %w", err)
	}

	// Marshal to JSON with deterministic encoding
	// Use the same encoder options as Phase 3.A/3.B golden tests
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	// Compact JSON (same as golden tests)
	var compactBuf bytes.Buffer
	if err := json.Compact(&compactBuf, data); err != nil {
		return fmt.Errorf("compacting JSON: %w", err)
	}
	buf := compactBuf.Bytes()

	// Write to temporary file
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, buf, 0o600); err != nil { //nolint:gosec // G306: output file needs read permissions
		return fmt.Errorf("writing temporary file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tmpPath)
		return fmt.Errorf("renaming temporary file: %w", err)
	}

	return nil
}
