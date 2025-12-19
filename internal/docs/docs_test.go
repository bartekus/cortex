// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package docs provides documentation generation tools.
package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bartekus/cortex/internal/features"
	"github.com/bartekus/cortex/internal/specschema"
)

// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md

func TestGenerateFeatureOverview_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")
	outPath := filepath.Join(tmpDir, "OVERVIEW.md")

	// Create minimal features.yaml
	content := `features:
  - id: FEATURE1
    title: "Feature 1"
    status: done
    spec: "test/feature1.md"
    owner: test
    tests: []
`

	if err := os.WriteFile(featuresPath, []byte(content), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	err := GenerateFeatureOverview(featuresPath, tmpDir, outPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected overview file to be created, got error: %v", err)
	}

	// Check content
	data, err := os.ReadFile(outPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("failed to read overview file: %v", err)
	}

	contentStr := string(data)
	if !strings.Contains(contentStr, "Cortex Features Overview") {
		t.Error("expected overview to contain title")
	}
	if !strings.Contains(contentStr, "FEATURE1") {
		t.Error("expected overview to contain FEATURE1")
	}
}

func TestGenerateFeatureOverview_WithSpecFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create spec dir: %v", err)
	}
	outPath := filepath.Join(tmpDir, "OVERVIEW.md")

	// Create features.yaml
	featuresContent := `features:
  - id: FEATURE1
    title: "Feature 1"
    status: done
    spec: "test/feature1.md"
    owner: test
    tests: []
`

	if err := os.WriteFile(featuresPath, []byte(featuresContent), 0o644); err != nil { //nolint:gosec // test file
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Create spec with frontmatter
	specSubDir := filepath.Join(specDir, "test")
	if err := os.MkdirAll(specSubDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create spec subdir: %v", err)
	}
	specPath := filepath.Join(specSubDir, "FEATURE1.md")
	specContent := `---
feature: FEATURE1
version: v1
status: done
domain: test
---
# Feature 1
`

	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil { //nolint:gosec // test file
		t.Fatalf("failed to write spec: %v", err)
	}

	err := GenerateFeatureOverview(featuresPath, specDir, outPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check domain is inferred from spec
	data, err := os.ReadFile(outPath) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("failed to read overview: %v", err)
	}

	contentStr := string(data)
	if !strings.Contains(contentStr, "test") {
		t.Error("expected overview to contain domain from spec frontmatter")
	}
}

func TestInferDomain_FromSpecFrontmatter(t *testing.T) {
	node := &features.FeatureNode{
		ID:   "FEATURE1",
		Spec: "test/feature1.md",
	}

	specMap := map[string]specschema.SpecFrontmatter{
		"FEATURE1": {
			Domain: "commands",
		},
	}

	domain := inferDomain(node, specMap)
	if domain != "commands" {
		t.Errorf("expected domain 'commands', got %q", domain)
	}
}

func TestInferDomain_FromSpecPath(t *testing.T) {
	node := &features.FeatureNode{
		ID:   "FEATURE1",
		Spec: "commands/build.md",
	}

	specMap := map[string]specschema.SpecFrontmatter{}

	domain := inferDomain(node, specMap)
	if domain != "commands" {
		t.Errorf("expected domain 'commands' from path, got %q", domain)
	}
}

func TestInferDomain_Default(t *testing.T) {
	node := &features.FeatureNode{
		ID:   "FEATURE1",
		Spec: "",
	}

	specMap := map[string]specschema.SpecFrontmatter{}

	domain := inferDomain(node, specMap)
	if domain != "unknown" {
		t.Errorf("expected domain 'unknown', got %q", domain)
	}
}

func TestGenerateMarkdown_IncludesAllSections(t *testing.T) {
	g := features.NewGraph()
	g.AddNode(&features.FeatureNode{
		ID:             "FEATURE1",
		Title:          "Feature 1",
		Implementation: "done",
	})
	g.AddNode(&features.FeatureNode{
		ID:             "FEATURE2",
		Title:          "Feature 2",
		Implementation: "wip",
		DependsOn:      []string{"FEATURE1"},
	})

	specMap := map[string]specschema.SpecFrontmatter{}

	markdown := generateMarkdown(g, specMap)

	if !strings.Contains(markdown, "# Cortex Features Overview") {
		t.Error("expected title")
	}
	if !strings.Contains(markdown, "## Features by Domain") {
		t.Error("expected features table section")
	}
	if !strings.Contains(markdown, "## Dependency Graph") {
		t.Error("expected dependency graph section")
	}
	if !strings.Contains(markdown, "## Status Summary") {
		t.Error("expected status summary section")
	}
	if !strings.Contains(markdown, "FEATURE1") {
		t.Error("expected FEATURE1 in output")
	}
	if !strings.Contains(markdown, "FEATURE2") {
		t.Error("expected FEATURE2 in output")
	}
}

func TestGenerateFeatureOverview_MissingFeaturesFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.yaml")
	outPath := filepath.Join(tmpDir, "OVERVIEW.md")

	err := GenerateFeatureOverview(nonExistentPath, tmpDir, outPath)
	if err == nil {
		t.Error("expected error for missing features file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load features") {
		t.Errorf("expected error to mention 'failed to load features', got: %v", err)
	}
}

func TestGenerateFeatureOverview_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")
	outPath := filepath.Join(tmpDir, "OVERVIEW.md")

	// Create invalid YAML
	content := `features:
  - id: FEATURE1
    title: "Feature 1"
    invalid: [unclosed bracket
`

	if err := os.WriteFile(featuresPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	err := GenerateFeatureOverview(featuresPath, tmpDir, outPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
