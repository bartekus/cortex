// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package features provides tools for managing feature dependency graphs.
package features

import (
	"fmt"
	"sort"
	"strings"
)

// ToDOT generates a DOT format representation of the feature dependency graph.
func ToDOT(g *Graph) string {
	var sb strings.Builder
	sb.WriteString("digraph feature_dependencies {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box];\n\n")

	// Sort node IDs for deterministic output
	nodeIDs := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Strings(nodeIDs)

	// Add nodes with status-based colors
	for _, id := range nodeIDs {
		node := g.Nodes[id]
		color := getStatusColor(node.Status)
		label := fmt.Sprintf("%s\\n[%s]", id, node.Status)
		sb.WriteString(fmt.Sprintf("  %q [label=%q fillcolor=%q style=filled];\n",
			id, label, color))
	}

	sb.WriteString("\n")

	// Add edges (dependencies) - sort for deterministic output
	for _, id := range nodeIDs {
		node := g.Nodes[id]
		// Sort dependencies for deterministic edge ordering
		deps := make([]string, len(node.DependsOn))
		copy(deps, node.DependsOn)
		sort.Strings(deps)
		for _, depID := range deps {
			sb.WriteString(fmt.Sprintf("  %q -> %q;\n", depID, id))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

// getStatusColor returns a color for a feature status.
func getStatusColor(status string) string {
	switch status {
	case "done":
		return "lightgreen"
	case "wip":
		return "lightyellow"
	case "todo":
		return "lightgray"
	default:
		return "white"
	}
}
