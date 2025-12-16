// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package specschema

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidateAll validates all specs and returns an error if any validation fails.
func ValidateAll(specs []Spec) error {
	var errors []string

	for i := range specs {
		if err := ValidateSpec(&specs[i]); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", specs[i].Path, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}

// ValidateSpec validates a single spec's frontmatter.
func ValidateSpec(spec *Spec) error {
	fm := spec.Frontmatter

	// Check required fields
	if fm.Feature == "" {
		return fmt.Errorf("missing required field: feature")
	}

	if fm.Version == "" {
		return fmt.Errorf("missing required field: version")
	}

	if fm.Status == "" {
		return fmt.Errorf("missing required field: status")
	}

	if fm.Domain == "" {
		return fmt.Errorf("missing required field: domain")
	}

	// Validate status enum
	validStatuses := map[string]bool{
		"todo": true,
		"wip":  true,
		"done": true,
	}
	if !validStatuses[fm.Status] {
		return fmt.Errorf("invalid status: %s (must be one of: todo, wip, done)", fm.Status)
	}

	// Note: Feature ID validation is handled by ValidateSpecIntegrity, which checks
	// against features.yaml. We don't validate against filename here to avoid conflicts
	// when feature IDs in features.yaml don't match filenames.

	// Validate domain matches path
	expectedDomain := inferDomainFromPath(spec.Path)
	if expectedDomain != "" && fm.Domain != expectedDomain {
		return fmt.Errorf("domain mismatch: frontmatter has %q but path suggests %q", fm.Domain, expectedDomain)
	}

	// Validate version format (should be v1, v2, etc.)
	if !isValidVersion(fm.Version) {
		return fmt.Errorf("invalid version format: %q (expected v1, v2, etc.)", fm.Version)
	}

	// Validate flags if present
	for i, flag := range fm.Inputs.Flags {
		if flag.Name == "" {
			return fmt.Errorf("flag[%d]: name is required", i)
		}
		// Normalize flag name for validation
		normalizedName := strings.TrimPrefix(strings.TrimPrefix(flag.Name, "--"), "-")
		if normalizedName == "" {
			return fmt.Errorf("flag[%d]: name cannot be empty after normalization", i)
		}
	}

	// Validate exit codes if present
	for name, code := range fm.Outputs.ExitCodes {
		if code < 0 {
			return fmt.Errorf("exit code %q has negative value: %d", name, code)
		}
	}

	return nil
}

// inferDomainFromPath extracts the domain from a spec file path.
// For example, "spec/commands/build.md" -> "commands"
func inferDomainFromPath(path string) string {
	// Remove spec/ prefix if present
	relPath := strings.TrimPrefix(path, "spec/")
	relPath = strings.TrimPrefix(relPath, "spec\\") // Windows

	// Get first directory component
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}

// isValidVersion checks if version follows the expected format (v1, v2, etc.)
var versionRegex = regexp.MustCompile(`^v\d+$`)

func isValidVersion(version string) bool {
	return versionRegex.MatchString(version)
}
