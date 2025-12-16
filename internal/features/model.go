// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

// FeatureNode represents a feature from features.yaml.
type FeatureNode struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Status      string   `yaml:"status"`
	Spec        string   `yaml:"spec"`
	Owner       string   `yaml:"owner"`
	Tests       []string `yaml:"tests"`
	DependsOn   []string `yaml:"depends_on"`
	Domain      string   `yaml:"domain"`
	Description string   `yaml:"description"`
}

// YAML represents the root structure of features.yaml.
type YAML struct {
	Features []FeatureNode `yaml:"features"`
}

// Graph represents a directed graph of features and their dependencies.
type Graph struct {
	Nodes map[string]*FeatureNode
	Edges map[string][]string // feature ID -> list of dependent feature IDs
}

// NewGraph creates a new empty graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*FeatureNode),
		Edges: make(map[string][]string),
	}
}

// AddNode adds a feature node to the graph.
func (g *Graph) AddNode(feature *FeatureNode) {
	g.Nodes[feature.ID] = feature
	// Initialize edges for this node
	if g.Edges[feature.ID] == nil {
		g.Edges[feature.ID] = []string{}
	}
}

// AddEdge adds a dependency edge: from depends on to.
// This means: if "to" changes, "from" is impacted.
func (g *Graph) AddEdge(from, to string) {
	if g.Edges[to] == nil {
		g.Edges[to] = []string{}
	}
	// Check if edge already exists
	for _, dep := range g.Edges[to] {
		if dep == from {
			return
		}
	}
	g.Edges[to] = append(g.Edges[to], from)
}
