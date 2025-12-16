// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: GOV_STATUS_ROADMAP
// Spec: spec/commands/status-roadmap.md

package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPhases_WithValidYAML(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	// Verify expected phases are detected
	expectedPhases := []string{
		"Architecture & Documentation",
		"Phase 0: Foundation",
		"Phase 1: Provider Interfaces",
		"Phase 2: Core Orchestration",
		"Phase 3: Local Development",
		"Governance",
	}

	if len(phases) != len(expectedPhases) {
		t.Errorf("DetectPhases() returned %d phases, want %d", len(phases), len(expectedPhases))
	}

	for _, expectedPhase := range expectedPhases {
		if _, exists := phases[expectedPhase]; !exists {
			t.Errorf("DetectPhases() missing expected phase %q", expectedPhase)
		}
	}
}

func TestDetectPhases_MapsFeaturesToPhases(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	// Verify specific feature mappings
	tests := []struct {
		featureID string
		phase     string
	}{
		{"ARCH_OVERVIEW", "Architecture & Documentation"},
		{"DOCS_ADR", "Architecture & Documentation"},
		{"CORE_CONFIG", "Phase 0: Foundation"},
		{"CLI_INIT", "Phase 0: Foundation"},
		{"CORE_LOGGING", "Phase 0: Foundation"},
		{"PROVIDER_BACKEND_INTERFACE", "Phase 1: Provider Interfaces"},
		{"PROVIDER_FRONTEND_INTERFACE", "Phase 1: Provider Interfaces"},
		{"PROVIDER_NETWORK_INTERFACE", "Phase 1: Provider Interfaces"},
		{"CORE_PLAN", "Phase 2: Core Orchestration"},
		{"CORE_STATE", "Phase 2: Core Orchestration"},
		{"CLI_DEPLOY", "Phase 2: Core Orchestration"},
		{"CLI_DEV", "Phase 3: Local Development"},
		{"DEV_HOSTS", "Phase 3: Local Development"},
		{"GOV_CORE", "Governance"},
		{"GOV_STATUS_ROADMAP", "Governance"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.featureID, func(t *testing.T) {
			phase, exists := phases[tt.phase]
			if !exists {
				t.Fatalf("phase %q not found", tt.phase)
			}

			found := false
			for _, feature := range phase.Features {
				if feature.ID == tt.featureID {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("feature %q not found in phase %q", tt.featureID, tt.phase)
			}
		})
	}
}

func TestDetectPhases_HandlesUncategorizedFeatures(t *testing.T) {
	t.Helper()

	// Create a temporary YAML with features before any phase comment
	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")

	yamlContent := `features:
  - id: UNCATEGORIZED_FEATURE
    title: "Uncategorized feature"
    status: todo
    spec: "test.md"
    owner: bart
    tests: []

  # Phase 0: Foundation
  - id: CATEGORIZED_FEATURE
    title: "Categorized feature"
    status: done
    spec: "test.md"
    owner: bart
    tests: []
`

	if err := os.WriteFile(featuresPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("failed to write test YAML: %v", err)
	}

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	// Verify uncategorized feature is mapped to "Uncategorized" phase
	uncategorizedPhase, exists := phases["Uncategorized"]
	if !exists {
		t.Fatal("expected 'Uncategorized' phase to exist")
	}

	found := false
	for _, feature := range uncategorizedPhase.Features {
		if feature.ID == "UNCATEGORIZED_FEATURE" {
			found = true
			break
		}
	}

	if !found {
		t.Error("uncategorized feature not found in 'Uncategorized' phase")
	}

	// Verify categorized feature is in correct phase
	phase0, exists := phases["Phase 0: Foundation"]
	if !exists {
		t.Fatal("expected 'Phase 0: Foundation' phase to exist")
	}

	found = false
	for _, feature := range phase0.Features {
		if feature.ID == "CATEGORIZED_FEATURE" {
			found = true
			break
		}
	}

	if !found {
		t.Error("categorized feature not found in 'Phase 0: Foundation' phase")
	}
}

func TestDetectPhases_HandlesMultiplePhaseComments(t *testing.T) {
	t.Helper()

	// Create a temporary YAML with multiple phase comments before a feature
	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")

	yamlContent := `features:
  # Phase 0: Foundation
  # Phase 1: Provider Interfaces
  - id: TEST_FEATURE
    title: "Test feature"
    status: done
    spec: "test.md"
    owner: bart
    tests: []
`

	if err := os.WriteFile(featuresPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("failed to write test YAML: %v", err)
	}

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	// Should use the last phase comment before the feature
	phase1, exists := phases["Phase 1: Provider Interfaces"]
	if !exists {
		t.Fatal("expected 'Phase 1: Provider Interfaces' phase to exist")
	}

	found := false
	for _, feature := range phase1.Features {
		if feature.ID == "TEST_FEATURE" {
			found = true
			break
		}
	}

	if !found {
		t.Error("feature not found in 'Phase 1: Provider Interfaces' phase")
	}
}

func TestDetectPhases_ReturnsErrorForInvalidYAML(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `features:
  - id: TEST
    title: "Test"
    status: invalid
    invalid: [unclosed
`

	if err := os.WriteFile(featuresPath, []byte(invalidYAML), 0o600); err != nil {
		t.Fatalf("failed to write invalid YAML: %v", err)
	}

	_, err := DetectPhases(featuresPath)
	if err == nil {
		t.Error("DetectPhases() expected error for invalid YAML, got nil")
	}
}

func TestDetectPhases_ReturnsErrorForMissingFile(t *testing.T) {
	t.Helper()

	_, err := DetectPhases("/nonexistent/path/features.yaml")
	if err == nil {
		t.Error("DetectPhases() expected error for missing file, got nil")
	}
}
