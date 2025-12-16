// SPDX-License-Identifier: AGPL-3.0-or-later

package featureindex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Registry contains the set of known Feature IDs.
type Registry map[string]bool

// Load reads the feature registry from the repository root.
// It expects to find `spec/features.yaml` relative to rootPath.
func Load(rootPath string) (Registry, error) {
	path := filepath.Join(rootPath, "spec", "features.yaml")

	//nolint:gosec // G304: path is constructed from rootPath + contract path
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading feature registry from %s: %w", path, err)
	}

	registry := make(Registry)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Handle list items: "- id: FEATURE" or "- FEATURE"
		if strings.HasPrefix(trimmed, "-") {
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		}

		// Handle key-value pairs: "id: FEATURE"
		if strings.HasPrefix(trimmed, "id:") {
			id := strings.TrimSpace(strings.TrimPrefix(trimmed, "id:"))
			if id != "" {
				registry[id] = true
			}
		}
	}

	return registry, nil
}
