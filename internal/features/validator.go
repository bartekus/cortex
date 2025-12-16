// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import (
	"fmt"
)

// ValidateDAG validates that the graph is acyclic (no dependency cycles).
func ValidateDAG(g *Graph) error {
	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var cycle []string

	for id := range g.Nodes {
		if !visited[id] {
			if hasCycle(g, id, visited, recStack, &cycle) {
				return fmt.Errorf("dependency cycle detected: %v", cycle)
			}
		}
	}

	return nil
}

// hasCycle performs DFS to detect cycles.
func hasCycle(g *Graph, nodeID string, visited, recStack map[string]bool, cycle *[]string) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	// Check dependencies (what this node depends on)
	node := g.Nodes[nodeID]
	for _, depID := range node.DependsOn {
		if !visited[depID] {
			if hasCycle(g, depID, visited, recStack, cycle) {
				*cycle = append([]string{nodeID}, *cycle...)
				return true
			}
		} else if recStack[depID] {
			// Found a back edge - cycle detected
			*cycle = []string{depID, nodeID}
			return true
		}
	}

	recStack[nodeID] = false
	return false
}
