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
		ID:     "FEATURE1",
		Title:  "Feature 1",
		Status: "done",
	})
	g.AddNode(&features.FeatureNode{
		ID:        "FEATURE2",
		Title:     "Feature 2",
		Status:    "wip",
		DependsOn: []string{"FEATURE1"},
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

func TestGenerateImplementationStatus_CreatesFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")
	outPath := filepath.Join(tmpDir, "implementation-status.md")

	// Create minimal features.yaml
	content := `features:
  - id: CLI_DEPLOY
    title: "Deploy command"
    status: done
    spec: "commands/deploy.md"
    owner: test
    tests: ["internal/cli/commands/deploy_test.go"]
`

	if err := os.WriteFile(featuresPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	err := GenerateImplementationStatus(featuresPath, outPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected implementation status file to be created, got error: %v", err)
	}

	// Check content
	data, err := os.ReadFile(outPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("failed to read implementation status file: %v", err)
	}

	contentStr := string(data)
	if !strings.Contains(contentStr, "# Implementation Status") {
		t.Error("expected implementation status to contain title")
	}
	if !strings.Contains(contentStr, "CLI_DEPLOY") {
		t.Error("expected implementation status to contain CLI_DEPLOY")
	}
}

func TestGenerateImplementationStatus_WithMultipleFeatures(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")
	outPath := filepath.Join(tmpDir, "implementation-status.md")

	// Create features.yaml with multiple features in different categories
	content := `features:
  - id: CLI_DEPLOY
    title: "Deploy command"
    status: done
    spec: "commands/deploy.md"
    owner: test
    tests: ["internal/cli/commands/deploy_test.go"]
  - id: CORE_CONFIG
    title: "Configuration system"
    status: done
    spec: "core/config.md"
    owner: test
    tests: ["pkg/config/config_test.go"]
  - id: PROVIDER_BACKEND_GENERIC
    title: "Generic backend provider"
    status: wip
    spec: "providers/backend/generic.md"
    owner: test
    tests: []
`

	if err := os.WriteFile(featuresPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	err := GenerateImplementationStatus(featuresPath, outPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Check content
	data, err := os.ReadFile(outPath) //nolint:gosec // G304: test file path
	if err != nil {
		t.Fatalf("failed to read implementation status file: %v", err)
	}

	contentStr := string(data)
	if !strings.Contains(contentStr, "CLI_DEPLOY") {
		t.Error("expected implementation status to contain CLI_DEPLOY")
	}
	if !strings.Contains(contentStr, "CORE_CONFIG") {
		t.Error("expected implementation status to contain CORE_CONFIG")
	}
	if !strings.Contains(contentStr, "PROVIDER_BACKEND_GENERIC") {
		t.Error("expected implementation status to contain PROVIDER_BACKEND_GENERIC")
	}
	if !strings.Contains(contentStr, "CLI Commands") {
		t.Error("expected implementation status to contain CLI Commands category")
	}
	if !strings.Contains(contentStr, "Core Functionality") {
		t.Error("expected implementation status to contain Core Functionality category")
	}
	if !strings.Contains(contentStr, "Providers") {
		t.Error("expected implementation status to contain Providers category")
	}
}

func TestGenerateImplementationStatus_MissingFeaturesFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.yaml")
	outPath := filepath.Join(tmpDir, "implementation-status.md")

	err := GenerateImplementationStatus(nonExistentPath, outPath)
	if err == nil {
		t.Error("expected error for missing features file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load features") {
		t.Errorf("expected error to mention 'failed to load features', got: %v", err)
	}
}

func TestGenerateImplementationStatus_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	featuresPath := filepath.Join(tmpDir, "features.yaml")
	outPath := filepath.Join(tmpDir, "implementation-status.md")

	// Create invalid YAML
	content := `features:
  - id: FEATURE1
    title: "Feature 1"
    status: done
    invalid: [unclosed bracket
`

	if err := os.WriteFile(featuresPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	err := GenerateImplementationStatus(featuresPath, outPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestGenerateImplementationStatusMarkdown_IncludesAllSections(t *testing.T) {
	t.Parallel()

	g := features.NewGraph()
	g.AddNode(&features.FeatureNode{
		ID:     "CLI_DEPLOY",
		Title:  "Deploy command",
		Status: "done",
		Spec:   "commands/deploy.md",
		Owner:  "test",
		Tests:  []string{"internal/cli/commands/deploy_test.go"},
	})
	g.AddNode(&features.FeatureNode{
		ID:     "CORE_CONFIG",
		Title:  "Configuration system",
		Status: "wip",
		Spec:   "core/config.md",
		Owner:  "test",
		Tests:  []string{},
	})

	markdown := generateImplementationStatusMarkdown(g)

	if !strings.Contains(markdown, "# Implementation Status") {
		t.Error("expected title")
	}
	if !strings.Contains(markdown, "## Feature Status Legend") {
		t.Error("expected status legend section")
	}
	if !strings.Contains(markdown, "## Features") {
		t.Error("expected features section")
	}
	if !strings.Contains(markdown, "## Implementation Notes") {
		t.Error("expected implementation notes section")
	}
	if !strings.Contains(markdown, "## Coverage Status") {
		t.Error("expected coverage status section")
	}
	if !strings.Contains(markdown, "CLI_DEPLOY") {
		t.Error("expected CLI_DEPLOY in output")
	}
	if !strings.Contains(markdown, "CORE_CONFIG") {
		t.Error("expected CORE_CONFIG in output")
	}
}

func TestGenerateImplementationStatusMarkdown_StatusGrouping(t *testing.T) {
	t.Parallel()

	g := features.NewGraph()
	g.AddNode(&features.FeatureNode{
		ID:     "FEATURE1",
		Title:  "Feature 1",
		Status: "done",
	})
	g.AddNode(&features.FeatureNode{
		ID:     "FEATURE2",
		Title:  "Feature 2",
		Status: "wip",
	})
	g.AddNode(&features.FeatureNode{
		ID:     "FEATURE3",
		Title:  "Feature 3",
		Status: "todo",
	})

	markdown := generateImplementationStatusMarkdown(g)

	if !strings.Contains(markdown, "### Completed Features") {
		t.Error("expected Completed Features section")
	}
	if !strings.Contains(markdown, "### In Progress") {
		t.Error("expected In Progress section")
	}
	if !strings.Contains(markdown, "FEATURE1") {
		t.Error("expected FEATURE1 in completed section")
	}
	if !strings.Contains(markdown, "FEATURE2") {
		t.Error("expected FEATURE2 in in-progress section")
	}
}

func TestInferCategory_CLI(t *testing.T) {
	t.Parallel()

	node := &features.FeatureNode{
		ID: "CLI_DEPLOY",
	}

	category := inferCategory(node)
	if category != "CLI Commands" {
		t.Errorf("expected category 'CLI Commands', got %q", category)
	}
}

func TestInferCategory_CORE(t *testing.T) {
	t.Parallel()

	node := &features.FeatureNode{
		ID: "CORE_CONFIG",
	}

	category := inferCategory(node)
	if category != "Core Functionality" {
		t.Errorf("expected category 'Core Functionality', got %q", category)
	}
}

func TestInferCategory_PROVIDER(t *testing.T) {
	t.Parallel()

	node := &features.FeatureNode{
		ID: "PROVIDER_BACKEND_GENERIC",
	}

	category := inferCategory(node)
	if category != "Providers" {
		t.Errorf("expected category 'Providers', got %q", category)
	}
}

func TestInferCategory_ARCH(t *testing.T) {
	t.Parallel()

	node := &features.FeatureNode{
		ID: "ARCH_DESIGN",
	}

	category := inferCategory(node)
	if category != "Architecture & Core" {
		t.Errorf("expected category 'Architecture & Core', got %q", category)
	}
}

func TestInferCategory_Default(t *testing.T) {
	t.Parallel()

	node := &features.FeatureNode{
		ID: "UNKNOWN_FEATURE",
	}

	category := inferCategory(node)
	if category != "Other" {
		t.Errorf("expected category 'Other', got %q", category)
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
