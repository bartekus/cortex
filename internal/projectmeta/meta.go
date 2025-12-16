// SPDX-License-Identifier: AGPL-3.0-or-later

package projectmeta

import (
	"path/filepath"
	"strings"
)

// DetermineRepoName determines the repository name from the root path.
// It uses a simple heuristic: the base name of the root directory.
func DetermineRepoName(rootPath string) string {
	base := filepath.Base(rootPath)
	if base == "" || base == "." || base == "/" {
		// Fallback if something is weird with the path
		return "project"
	}
	return strings.TrimSpace(base)
}
