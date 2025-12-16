// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import (
	"os"
	"path/filepath"
	"testing"
)

// Feature: GOV_CORE
// Spec: spec/governance/GOV_CORE.md

func TestNewGraph(t *testing.T) {
	g := NewGraph()
	if g.Nodes == nil {
		t.Fatal("expected Nodes map to be initialized")
	}
	if g.Edges == nil {
		t.Fatal("expected Edges map to be initialized")
	}
}

func TestGraph_AddNode(t *testing.T) {
	g := NewGraph()
	node := &FeatureNode{
		ID:     "FEATURE1",
		Title:  "Test Feature",
		Status: "done",
	}

	g.AddNode(node)

	if g.Nodes["FEATURE1"] == nil {
		t.Fatal("expected node to be added")
	}
	if g.Nodes["FEATURE1"].ID != "FEATURE1" {
		t.Errorf("expected node ID 'FEATURE1', got %q", g.Nodes["FEATURE1"].ID)
	}
}

func TestGraph_AddEdge(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1"})
	g.AddNode(&FeatureNode{ID: "FEATURE2"})

	// FEATURE2 depends on FEATURE1
	g.AddEdge("FEATURE2", "FEATURE1")

	dependents := g.Edges["FEATURE1"]
	if len(dependents) != 1 {
		t.Fatalf("expected 1 dependent, got %d", len(dependents))
	}
	if dependents[0] != "FEATURE2" {
		t.Errorf("expected dependent 'FEATURE2', got %q", dependents[0])
	}
}

func TestGraph_AddEdge_Duplicate(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1"})
	g.AddNode(&FeatureNode{ID: "FEATURE2"})

	g.AddEdge("FEATURE2", "FEATURE1")
	g.AddEdge("FEATURE2", "FEATURE1") // Duplicate

	dependents := g.Edges["FEATURE1"]
	if len(dependents) != 1 {
		t.Fatalf("expected 1 dependent (no duplicates), got %d", len(dependents))
	}
}

func TestLoadGraph_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")

	content := `features:
  - id: FEATURE1
    title: "Feature 1"
    status: done
    spec: "test/feature1.md"
    owner: test
    tests: []
  - id: FEATURE2
    title: "Feature 2"
    status: wip
    spec: "test/feature2.md"
    owner: test
    tests: []
    depends_on:
      - FEATURE1
`

	if err := os.WriteFile(featuresPath, []byte(content), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	graph, err := LoadGraph(featuresPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	if graph.Nodes["FEATURE1"] == nil {
		t.Fatal("expected FEATURE1 node")
	}
	if graph.Nodes["FEATURE2"] == nil {
		t.Fatal("expected FEATURE2 node")
	}

	// Check dependency edge
	dependents := graph.Edges["FEATURE1"]
	if len(dependents) != 1 {
		t.Fatalf("expected 1 dependent for FEATURE1, got %d", len(dependents))
	}
	if dependents[0] != "FEATURE2" {
		t.Errorf("expected FEATURE2 to depend on FEATURE1, got %q", dependents[0])
	}
}

func TestLoadGraph_UnknownDependency(t *testing.T) {
	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")

	content := `features:
  - id: FEATURE1
    title: "Feature 1"
    status: done
    spec: "test/feature1.md"
    owner: test
    tests: []
    depends_on:
      - UNKNOWN_FEATURE
`

	if err := os.WriteFile(featuresPath, []byte(content), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	_, err := LoadGraph(featuresPath)
	if err == nil {
		t.Fatal("expected error for unknown dependency")
	}
}

func TestValidateDAG_Acyclic(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1"})
	g.AddNode(&FeatureNode{ID: "FEATURE2"})
	g.AddNode(&FeatureNode{ID: "FEATURE3"})

	// FEATURE2 depends on FEATURE1
	g.AddEdge("FEATURE2", "FEATURE1")
	// FEATURE3 depends on FEATURE2
	g.AddEdge("FEATURE3", "FEATURE2")

	err := ValidateDAG(g)
	if err != nil {
		t.Errorf("expected no error for acyclic graph, got: %v", err)
	}
}

func TestValidateDAG_Cycle(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1", DependsOn: []string{"FEATURE2"}})
	g.AddNode(&FeatureNode{ID: "FEATURE2", DependsOn: []string{"FEATURE1"}})

	g.AddEdge("FEATURE1", "FEATURE2")
	g.AddEdge("FEATURE2", "FEATURE1")

	err := ValidateDAG(g)
	if err == nil {
		t.Fatal("expected error for cyclic graph")
	}
}

func TestValidateDAG_SelfCycle(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1", DependsOn: []string{"FEATURE1"}})

	g.AddEdge("FEATURE1", "FEATURE1")

	err := ValidateDAG(g)
	if err == nil {
		t.Fatal("expected error for self-cycle")
	}
}

func TestImpact_DirectDependency(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1"})
	g.AddNode(&FeatureNode{ID: "FEATURE2"})
	g.AddNode(&FeatureNode{ID: "FEATURE3"})

	g.AddEdge("FEATURE2", "FEATURE1")
	g.AddEdge("FEATURE3", "FEATURE1")

	impacted := Impact(g, "FEATURE1")
	if len(impacted) != 2 {
		t.Fatalf("expected 2 impacted features, got %d", len(impacted))
	}

	// Check both are present
	found2 := false
	found3 := false
	for _, id := range impacted {
		if id == "FEATURE2" {
			found2 = true
		}
		if id == "FEATURE3" {
			found3 = true
		}
	}
	if !found2 || !found3 {
		t.Errorf("expected FEATURE2 and FEATURE3 to be impacted, got %v", impacted)
	}
}

func TestImpact_TransitiveDependency(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1"})
	g.AddNode(&FeatureNode{ID: "FEATURE2"})
	g.AddNode(&FeatureNode{ID: "FEATURE3"})

	// FEATURE2 depends on FEATURE1
	g.AddEdge("FEATURE2", "FEATURE1")
	// FEATURE3 depends on FEATURE2
	g.AddEdge("FEATURE3", "FEATURE2")

	impacted := Impact(g, "FEATURE1")
	if len(impacted) != 2 {
		t.Fatalf("expected 2 impacted features (direct + transitive), got %d", len(impacted))
	}
}

func TestImpact_NoDependencies(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1"})
	g.AddNode(&FeatureNode{ID: "FEATURE2"})

	impacted := Impact(g, "FEATURE1")
	if len(impacted) != 0 {
		t.Fatalf("expected 0 impacted features, got %d", len(impacted))
	}
}

func TestImpact_UnknownFeature(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1"})

	impacted := Impact(g, "UNKNOWN")
	if len(impacted) != 0 {
		t.Fatalf("expected 0 impacted features for unknown feature, got %d", len(impacted))
	}
}

func TestToDOT_GeneratesValidDOT(t *testing.T) {
	g := NewGraph()
	g.AddNode(&FeatureNode{ID: "FEATURE1", Status: "done"})
	g.AddNode(&FeatureNode{ID: "FEATURE2", Status: "wip"})
	g.AddNode(&FeatureNode{ID: "FEATURE3", Status: "todo"})

	g.AddEdge("FEATURE2", "FEATURE1")
	g.AddEdge("FEATURE3", "FEATURE2")

	dot := ToDOT(g)

	// Basic checks
	if !contains(dot, "digraph") {
		t.Error("expected DOT to contain 'digraph'")
	}
	if !contains(dot, "FEATURE1") {
		t.Error("expected DOT to contain 'FEATURE1'")
	}
	if !contains(dot, "FEATURE2") {
		t.Error("expected DOT to contain 'FEATURE2'")
	}
	if !contains(dot, "FEATURE1") && !contains(dot, "->") {
		t.Error("expected DOT to contain edge '->'")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
