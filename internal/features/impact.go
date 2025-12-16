// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import "sort"

// Impact returns all features that directly or transitively depend on the given feature ID.
// This is the "impact analysis" - if feature ID changes, which features are affected?
// Results are sorted lexicographically for deterministic output.
func Impact(g *Graph, featureID string) []string {
	if _, exists := g.Nodes[featureID]; !exists {
		return nil
	}

	impacted := make(map[string]bool)
	collectImpacted(g, featureID, impacted)

	// Convert to slice and sort for deterministic output
	result := make([]string, 0, len(impacted))
	for id := range impacted {
		result = append(result, id)
	}
	sort.Strings(result)

	return result
}

// collectImpacted recursively collects all features that depend on the given feature.
func collectImpacted(g *Graph, featureID string, visited map[string]bool) {
	// Get all features that depend on this one (edges point from dependency to dependent)
	dependents := g.Edges[featureID]
	for _, dependentID := range dependents {
		if !visited[dependentID] {
			visited[dependentID] = true
			// Recursively collect features that depend on this dependent
			collectImpacted(g, dependentID, visited)
		}
	}
}
