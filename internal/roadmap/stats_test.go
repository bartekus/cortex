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
	"path/filepath"
	"testing"
)

func TestCalculateStats_WithTestData(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	stats := CalculateStats(phases)

	// Verify overall statistics
	// testdata/features.yaml has 15 features: 8 done, 2 wip, 5 todo
	if stats.Total != 15 {
		t.Errorf("CalculateStats() Total = %d, want 15", stats.Total)
	}

	if stats.Done != 8 {
		t.Errorf("CalculateStats() Done = %d, want 8", stats.Done)
	}

	if stats.WIP != 2 {
		t.Errorf("CalculateStats() WIP = %d, want 2", stats.WIP)
	}

	if stats.Todo != 5 {
		t.Errorf("CalculateStats() Todo = %d, want 5", stats.Todo)
	}

	expectedCompletion := float64(8) / float64(15) * 100
	if stats.CompletionPercentage < expectedCompletion-0.1 || stats.CompletionPercentage > expectedCompletion+0.1 {
		t.Errorf("CalculateStats() CompletionPercentage = %.1f, want ~%.1f", stats.CompletionPercentage, expectedCompletion)
	}
}

func TestCalculateStats_PerPhaseStatistics(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	stats := CalculateStats(phases)

	// Verify Phase 0: Foundation (3 done, 0 wip, 0 todo)
	phase0Stats := stats.PhaseStats["Phase 0: Foundation"]
	if phase0Stats.Total != 3 {
		t.Errorf("Phase 0 Total = %d, want 3", phase0Stats.Total)
	}
	if phase0Stats.Done != 3 {
		t.Errorf("Phase 0 Done = %d, want 3", phase0Stats.Done)
	}
	if phase0Stats.WIP != 0 {
		t.Errorf("Phase 0 WIP = %d, want 0", phase0Stats.WIP)
	}
	if phase0Stats.Todo != 0 {
		t.Errorf("Phase 0 Todo = %d, want 0", phase0Stats.Todo)
	}
	if phase0Stats.CompletionPercentage != 100.0 {
		t.Errorf("Phase 0 CompletionPercentage = %.1f, want 100.0", phase0Stats.CompletionPercentage)
	}

	// Verify Phase 1: Provider Interfaces (2 done, 1 wip, 0 todo)
	phase1Stats := stats.PhaseStats["Phase 1: Provider Interfaces"]
	if phase1Stats.Total != 3 {
		t.Errorf("Phase 1 Total = %d, want 3", phase1Stats.Total)
	}
	if phase1Stats.Done != 2 {
		t.Errorf("Phase 1 Done = %d, want 2", phase1Stats.Done)
	}
	if phase1Stats.WIP != 1 {
		t.Errorf("Phase 1 WIP = %d, want 1", phase1Stats.WIP)
	}
	if phase1Stats.Todo != 0 {
		t.Errorf("Phase 1 Todo = %d, want 0", phase1Stats.Todo)
	}
	expectedPhase1Completion := float64(2) / float64(3) * 100
	if phase1Stats.CompletionPercentage < expectedPhase1Completion-0.1 || phase1Stats.CompletionPercentage > expectedPhase1Completion+0.1 {
		t.Errorf("Phase 1 CompletionPercentage = %.1f, want ~%.1f", phase1Stats.CompletionPercentage, expectedPhase1Completion)
	}

	// Verify Architecture & Documentation (0 done, 0 wip, 2 todo)
	archStats := stats.PhaseStats["Architecture & Documentation"]
	if archStats.Total != 2 {
		t.Errorf("Architecture Total = %d, want 2", archStats.Total)
	}
	if archStats.Done != 0 {
		t.Errorf("Architecture Done = %d, want 0", archStats.Done)
	}
	if archStats.WIP != 0 {
		t.Errorf("Architecture WIP = %d, want 0", archStats.WIP)
	}
	if archStats.Todo != 2 {
		t.Errorf("Architecture Todo = %d, want 2", archStats.Todo)
	}
	if archStats.CompletionPercentage != 0.0 {
		t.Errorf("Architecture CompletionPercentage = %.1f, want 0.0", archStats.CompletionPercentage)
	}
}

func TestCalculateStats_EmptyPhases(t *testing.T) {
	t.Helper()

	phases := make(map[string]*Phase)
	stats := CalculateStats(phases)

	if stats.Total != 0 {
		t.Errorf("CalculateStats() with empty phases Total = %d, want 0", stats.Total)
	}

	if stats.Done != 0 {
		t.Errorf("CalculateStats() with empty phases Done = %d, want 0", stats.Done)
	}

	if stats.CompletionPercentage != 0.0 {
		t.Errorf("CalculateStats() with empty phases CompletionPercentage = %.1f, want 0.0", stats.CompletionPercentage)
	}
}

func TestIdentifyBlockers_WithDependencies(t *testing.T) {
	t.Helper()

	testDataDir := testDataDir(t)
	featuresPath := filepath.Join(testDataDir, "features.yaml")

	phases, err := DetectPhases(featuresPath)
	if err != nil {
		t.Fatalf("DetectPhases() failed: %v", err)
	}

	blockers := IdentifyBlockers(phases)

	// DEV_HOSTS depends on CLI_DEV (wip), so DEV_HOSTS should be blocked
	found := false
	for _, blocker := range blockers {
		if blocker.FeatureID != "DEV_HOSTS" {
			continue
		}
		found = true
		if len(blocker.BlockedBy) == 0 {
			t.Error("DEV_HOSTS blocker missing BlockedBy dependencies")
		}
		hasCLIDev := false
		for _, dep := range blocker.BlockedBy {
			if dep == "CLI_DEV" {
				hasCLIDev = true
				break
			}
		}
		if !hasCLIDev {
			t.Error("DEV_HOSTS blocker missing CLI_DEV in BlockedBy")
		}
		break
	}

	if !found {
		t.Error("DEV_HOSTS not identified as blocker")
	}

	// CLI_DEPLOY depends on CORE_PLAN and CORE_STATE (both done), so CLI_DEPLOY should NOT be blocked
	for _, blocker := range blockers {
		if blocker.FeatureID == "CLI_DEPLOY" {
			t.Error("CLI_DEPLOY should not be blocked (all dependencies are done)")
			break
		}
	}

	// Features with done dependencies should not be blockers
	// GOV_STATUS_ROADMAP depends on GOV_CORE (done), so should not be blocked
	for _, blocker := range blockers {
		if blocker.FeatureID == "GOV_STATUS_ROADMAP" {
			t.Error("GOV_STATUS_ROADMAP should not be blocked (dependency is done)")
		}
	}
}

func TestIdentifyBlockers_NoBlockers(t *testing.T) {
	t.Helper()

	// Create phases with all features done
	phases := make(map[string]*Phase)
	phases["Test Phase"] = &Phase{
		Name: "Test Phase",
		Features: []Feature{
			{ID: "FEATURE_1", Status: "done"},
			{ID: "FEATURE_2", Status: "done", DependsOn: []string{"FEATURE_1"}},
		},
	}

	blockers := IdentifyBlockers(phases)

	if len(blockers) != 0 {
		t.Errorf("IdentifyBlockers() with all done features returned %d blockers, want 0", len(blockers))
	}
}

func TestIdentifyBlockers_AllBlocked(t *testing.T) {
	t.Helper()

	// Create phases with circular or chained dependencies
	phases := make(map[string]*Phase)
	phases["Test Phase"] = &Phase{
		Name: "Test Phase",
		Features: []Feature{
			{ID: "FEATURE_1", Status: "todo"},
			{ID: "FEATURE_2", Status: "todo", DependsOn: []string{"FEATURE_1"}},
			{ID: "FEATURE_3", Status: "todo", DependsOn: []string{"FEATURE_2"}},
		},
	}

	blockers := IdentifyBlockers(phases)

	// FEATURE_2 should be blocked by FEATURE_1
	// FEATURE_3 should be blocked by FEATURE_2
	if len(blockers) < 2 {
		t.Errorf("IdentifyBlockers() returned %d blockers, want at least 2", len(blockers))
	}

	foundFeature2 := false
	foundFeature3 := false
	for _, blocker := range blockers {
		if blocker.FeatureID == "FEATURE_2" {
			foundFeature2 = true
		}
		if blocker.FeatureID == "FEATURE_3" {
			foundFeature3 = true
		}
	}

	if !foundFeature2 {
		t.Error("FEATURE_2 not identified as blocker")
	}
	if !foundFeature3 {
		t.Error("FEATURE_3 not identified as blocker")
	}
}
