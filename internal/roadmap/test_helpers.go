// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CLI_COMMAND_STATUS
// Spec: spec/cli/status.md

package roadmap

import (
	"path/filepath"
	"runtime"
	"testing"
)

// testDataDir returns the testdata directory path.
func testDataDir(t *testing.T) string {
	t.Helper()

	// Get the directory where this test file is located
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Join(testDir, "testdata")
}
