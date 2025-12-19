package gov

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateTraceability checks if spec files exist and reference the feature ID.
func (r *Registry) ValidateTraceability(rootDir string) error {
	for _, f := range r.Features {
		fullPath := filepath.Join(rootDir, f.Spec)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("feature %s spec file not found: %s", f.ID, f.Spec)
			}
			return fmt.Errorf("feature %s failed to read spec file: %w", f.ID, err)
		}

		// Check if feature ID is referenced inside the spec
		if !strings.Contains(string(content), f.ID) {
			return fmt.Errorf("feature %s spec file %s does not reference feature ID %s", f.ID, f.Spec, f.ID)
		}
	}
	return nil
}

// ValidateDependencies checks for missing dependencies and cycles.
func (r *Registry) ValidateDependencies() error {
	idMap := make(map[string]bool)
	for _, f := range r.Features {
		idMap[f.ID] = true
	}

	for _, f := range r.Features {
		for _, depID := range f.DependsOn {
			if !idMap[depID] {
				return fmt.Errorf("feature %s depends on unknown feature ID: %s", f.ID, depID)
			}
		}
	}

	// Cycle detection using DFS
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var detectCycle func(currentID string) error
	detectCycle = func(currentID string) error {
		visited[currentID] = true
		recursionStack[currentID] = true

		// Find the feature object to get its dependencies
		var currentFeature *Feature
		for i := range r.Features {
			if r.Features[i].ID == currentID {
				currentFeature = &r.Features[i]
				break
			}
		}

		if currentFeature != nil {
			for _, depID := range currentFeature.DependsOn {
				if !visited[depID] {
					if err := detectCycle(depID); err != nil {
						return err
					}
				} else if recursionStack[depID] {
					return fmt.Errorf("dependency cycle detected involving: %s -> %s", currentID, depID)
				}
			}
		}

		recursionStack[currentID] = false
		return nil
	}

	for _, f := range r.Features {
		if !visited[f.ID] {
			if err := detectCycle(f.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
