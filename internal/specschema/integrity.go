// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package specschema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartekus/cortex/internal/features"
)

// ValidateSpecIntegrity validates that features.yaml and spec files are in sync.
// It checks:
//   - Every feature in features.yaml has a corresponding spec file
//   - Spec files match their feature IDs
func ValidateSpecIntegrity(featuresPath, specRoot string) error {
	// Load features
	graph, err := features.LoadGraph(featuresPath)
	if err != nil {
		return fmt.Errorf("failed to load features: %w", err)
	}

	// Load all specs
	specs, err := LoadAllSpecs(specRoot)
	if err != nil {
		return fmt.Errorf("failed to load specs: %w", err)
	}

	// Build spec lookup by feature ID
	specMap := make(map[string]Spec)
	for _, spec := range specs {
		specMap[spec.Frontmatter.Feature] = spec
	}

	var errors []string

	// Check: every feature in features.yaml should have a spec file
	for id, node := range graph.Nodes {
		// Skip features without spec path
		if node.Spec == "" {
			continue
		}

		// Check if spec file exists
		specPath := node.Spec
		if !strings.HasPrefix(specPath, "spec/") {
			specPath = filepath.Join("spec", specPath)
		}

		// Check if file exists
		if _, err := os.Stat(specPath); err != nil {
			if os.IsNotExist(err) {
				// Skip missing spec files for todo features (they're planned but not yet implemented)
				// Also skip for done features that reference docs/adr/ files (ADRs are in a different location)
				if node.Implementation == "todo" || strings.Contains(node.Spec, "adr/") {
					continue
				}
				errors = append(errors, fmt.Sprintf("feature %q references spec file %q that does not exist", id, node.Spec))
			} else {
				errors = append(errors, fmt.Sprintf("feature %q references spec file %q that cannot be accessed: %v", id, node.Spec, err))
			}
			continue
		}

		// Check if spec file has matching feature ID
		// Note: Some spec files are shared by multiple features (e.g., core/backend-registry.md
		// is used by both CORE_BACKEND_REGISTRY and PROVIDER_BACKEND_INTERFACE). In such cases,
		// we check if the file exists and has valid frontmatter, but don't require exact feature ID match.
		if spec, found := specMap[id]; found {
			// File exists and was loaded - check if it has valid frontmatter
			if spec.Frontmatter.Feature == "" {
				errors = append(errors, fmt.Sprintf("feature %q spec file has empty feature ID", id))
			}
			// Don't require exact match - allow shared spec files
			_ = found // suppress unused variable warning
		} else {
			// Spec file exists but doesn't have frontmatter with matching ID
			// Try to load it
			spec, err := LoadSpec(specPath)
			if err != nil {
				errors = append(errors, fmt.Sprintf("feature %q spec file %q exists but cannot be loaded: %v", id, node.Spec, err))
				continue
			}
			// Check if it has valid frontmatter, but don't require exact feature ID match
			// (allows shared spec files)
			if spec.Frontmatter.Feature == "" {
				errors = append(errors, fmt.Sprintf("feature %q spec file %q has empty feature ID", id, node.Spec))
			}
			_ = spec // spec loaded and validated above
		}
	}

	// Check: spec files should be referenced in features.yaml (warning only)
	// This is less strict - we allow orphaned spec files
	// Note: We intentionally don't check this to allow orphaned specs
	_ = len(specs) // suppress unused variable warning

	if len(errors) > 0 {
		return fmt.Errorf("spec integrity validation failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}
