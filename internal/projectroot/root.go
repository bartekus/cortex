// SPDX-License-Identifier: AGPL-3.0-or-later

package projectroot

import (
	"fmt"
	"os"
	"path/filepath"
)

// Find locates the repository root by walking upwards from start, looking for markers.
// Priority order:
// 1. spec/features.yaml (Contract)
// 2. go.mod (Go project)
// 3. .git (Git root)
// 4. Agent.md (Optional/Legacy)
func Find(start string) (string, error) {
	absStart, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolving start path: %w", err)
	}

	current := absStart
	root := filepath.VolumeName(current) + string(filepath.Separator)
	if root == string(filepath.Separator) {
		root = "/"
	}

	for {
		// Check for markers in priority order
		if hasFile(current, "spec/features.yaml") {
			return current, nil
		}
		if hasFile(current, "go.mod") {
			return current, nil
		}
		if hasDir(current, ".git") {
			return current, nil
		}
		if hasFile(current, "Agent.md") {
			// Optional/Legacy marker
			return current, nil
		}

		// Stop if we've reached the filesystem root
		if current == root || current == filepath.Dir(current) {
			return "", fmt.Errorf("repository root not found (searched for spec/features.yaml, go.mod, .git, Agent.md) at %s", current)
		}

		// Move up one directory
		current = filepath.Dir(current)
	}
}

func hasFile(dir, name string) bool {
	path := filepath.Join(dir, name)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func hasDir(dir, name string) bool {
	path := filepath.Join(dir, name)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
