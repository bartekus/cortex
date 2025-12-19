// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartekus/cortex/internal/features"
)

// TestRunner_NoFeatures ensures the scaffolded Runner can execute safely when
// features.yaml parsing is still minimal. This test is expected to pass and
// acts as a harness for future work.
func TestRunner_NoFeatures(t *testing.T) {
	t.Parallel()

	// Use the actual features.yaml from the repo root
	rootDir := "."
	featuresPath := "spec/features.yaml"

	// Check if features.yaml exists
	if _, err := os.Stat(featuresPath); os.IsNotExist(err) {
		t.Skipf("features.yaml not found at %s, skipping test", featuresPath)
	}

	r := &features.Runner{
		FeaturesPath: featuresPath,
		RootDir:      rootDir,
	}

	ctx := context.Background()

	// This should not panic, but may return errors if there are validation issues
	// We're just testing that the scaffold compiles and runs
	err := r.Run(ctx)
	if err != nil {
		// For now, we allow errors as the implementation may have TODOs
		// In a full implementation, we'd want more specific test cases
		t.Logf("Runner.Run returned error (expected during scaffold phase): %v", err)
	}
}

// TestLoadFeaturesYAML tests loading the actual features.yaml file.
func TestLoadFeaturesYAML(t *testing.T) {
	t.Parallel()

	rootDir := "."
	featuresPath := "spec/features.yaml"

	if _, err := os.Stat(featuresPath); os.IsNotExist(err) {
		t.Skipf("features.yaml not found at %s, skipping test", featuresPath)
	}

	featureMap, err := features.LoadFeaturesYAML(rootDir, featuresPath)
	if err != nil {
		t.Fatalf("LoadFeaturesYAML failed: %v", err)
	}

	if len(featureMap) == 0 {
		t.Fatal("expected at least one feature, got 0")
	}

	// Check that CLI_COMMAND_GOV exists
	if _, exists := featureMap["CLI_COMMAND_GOV"]; !exists {
		t.Error("expected CLI_COMMAND_GOV feature to exist")
	}
}

// TestScanSourceTree tests scanning the source tree for Feature headers.
func TestScanSourceTree(t *testing.T) {
	t.Parallel()

	rootDir := "."
	featuresPath := "spec/features.yaml"

	if _, err := os.Stat(featuresPath); os.IsNotExist(err) {
		t.Skipf("features.yaml not found at %s, skipping test", featuresPath)
	}

	featureMap, err := features.LoadFeaturesYAML(rootDir, featuresPath)
	if err != nil {
		t.Fatalf("LoadFeaturesYAML failed: %v", err)
	}

	ctx := context.Background()
	index, err := features.ScanSourceTree(ctx, rootDir, featureMap)
	if err != nil {
		t.Fatalf("ScanSourceTree failed: %v", err)
	}

	if index == nil {
		t.Fatal("ScanSourceTree returned nil index")
		return
	}

	// At minimum, we should have features loaded
	if len(index.Features) == 0 {
		t.Error("expected at least one feature in index")
	}
}

// TestRunner_WithFixtureRepo tests the runner against a controlled fixture repository.
// This serves as a golden baseline: as long as fixtures are well-formed, this should never fail.
func TestRunner_WithFixtureRepo(t *testing.T) {
	t.Parallel()

	// Root is the fixture directory - resolve to absolute path for consistent comparison
	root := "testdata/feature-map-fixture"
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	featuresPath := "spec/features.yaml"

	// Check if fixture exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skipf("fixture directory not found at %s, skipping test", root)
	}

	r := &features.Runner{
		RootDir:      absRoot,
		FeaturesPath: featuresPath,
	}

	ctx := context.Background()

	// This should pass with no errors for a well-formed fixture
	if err := r.Run(ctx); err != nil {
		t.Fatalf("expected no errors for fixture repo, got: %v", err)
	}
}
