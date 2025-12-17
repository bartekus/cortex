// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/features"
)

// GenerateImplementationStatus generates docs/engine/status/implementation-status.md from features.yaml.
func GenerateImplementationStatus(featuresPath, outPath string) error {
	// Load features
	graph, err := features.LoadGraph(featuresPath)
	if err != nil {
		return fmt.Errorf("failed to load features: %w", err)
	}

	// Generate markdown
	content := generateImplementationStatusMarkdown(graph)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil { //nolint:gosec // output directory needs write permissions
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil { //nolint:gosec // output file needs read permissions
		return fmt.Errorf("failed to write implementation status file: %w", err)
	}

	return nil
}

// generateImplementationStatusMarkdown generates the markdown content for implementation status.
func generateImplementationStatusMarkdown(graph *features.Graph) string {
	var sb strings.Builder

	sb.WriteString("# Implementation Status\n\n")
	sb.WriteString("> **⚠️ Note**: This document is a snapshot view. For the complete, up-to-date feature list, see [`spec/features.yaml`](../../../spec/features.yaml).\n>\n")
	sb.WriteString("> This document shows a subset of features for quick reference. The full feature catalog with 61+ features organized by phase is available in [`docs/narrative/implementation-roadmap.md`](../../narrative/implementation-roadmap.md).\n\n")
	sb.WriteString("This document tracks the implementation status of Cortex features. It should be regenerated from `spec/features.yaml` when needed.\n\n")
	sb.WriteString("> **Last Updated**: See `spec/features.yaml` for the source of truth.\n\n")
	sb.WriteString("## Feature Status Legend\n\n")
	sb.WriteString("- **done** - Feature is complete with tests and documentation\n")
	sb.WriteString("- **wip** - Feature is in progress\n")
	sb.WriteString("- **todo** - Feature is planned but not started\n")
	sb.WriteString("- **blocked** - Feature is blocked by dependencies\n\n")
	sb.WriteString("## Features\n\n")

	// Group features by category
	byCategory := make(map[string][]*features.FeatureNode)
	for _, node := range graph.Nodes {
		category := inferCategory(node)
		byCategory[category] = append(byCategory[category], node)
	}

	// Sort categories
	categories := make([]string, 0, len(byCategory))
	for category := range byCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	// Generate tables for each category
	for _, category := range categories {
		nodes := byCategory[category]
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].ID < nodes[j].ID
		})

		sb.WriteString(fmt.Sprintf("### %s\n\n", category))
		sb.WriteString("| ID | Title | Status | Owner | Spec | Tests |\n")
		sb.WriteString("|----|-------|--------|-------|------|-------|\n")

		for _, node := range nodes {
			specLink := "-"
			if node.Spec != "" {
				specName := filepath.Base(node.Spec)
				specLink = fmt.Sprintf("[%s](../../../spec/%s)", specName, node.Spec)
			}

			testsLink := "-"
			if len(node.Tests) > 0 {
				var testLinks []string
				for _, test := range node.Tests {
					testName := filepath.Base(test)
					testLinks = append(testLinks, fmt.Sprintf("[%s](../../../%s)", testName, test))
				}
				testsLink = strings.Join(testLinks, ", ")
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
				node.ID, node.Title, node.Status, node.Owner, specLink, testsLink))
		}

		sb.WriteString("\n")
	}

	// Implementation Notes section
	sb.WriteString("## Implementation Notes\n\n")

	// Group by status
	byStatus := make(map[string][]*features.FeatureNode)
	for _, node := range graph.Nodes {
		byStatus[node.Status] = append(byStatus[node.Status], node)
	}

	// Completed Features
	if done, ok := byStatus["done"]; ok {
		sb.WriteString("### Completed Features\n\n")
		sort.Slice(done, func(i, j int) bool {
			return done[i].ID < done[j].ID
		})
		for _, node := range done {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", node.ID, node.Title))
		}
		sb.WriteString("\n")
	}

	// In Progress
	if wip, ok := byStatus["wip"]; ok {
		sb.WriteString("### In Progress\n\n")
		sort.Slice(wip, func(i, j int) bool {
			return wip[i].ID < wip[j].ID
		})
		for _, node := range wip {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", node.ID, node.Title))
		}
		sb.WriteString("\n")
	}

	// Coverage Status
	sb.WriteString("## Coverage Status\n\n")
	sb.WriteString("Current test coverage targets:\n")
	sb.WriteString("- **Core packages** (`pkg/config`, `internal/core`): Target 80%+\n")
	sb.WriteString("- **CLI layer** (`internal/cli`): Target 70%+\n")
	sb.WriteString("- **Drivers** (`internal/drivers`): Target 70%+\n")
	sb.WriteString("- **Overall**: Target 60%+ (increasing to 80% as project matures)\n\n")

	// How to Update
	sb.WriteString("## How to Update\n\n")
	sb.WriteString("This file should be regenerated when `spec/features.yaml` changes. To update:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Run the generator script\n")
	sb.WriteString("./scripts/generate-implementation-status.sh\n")
	sb.WriteString("```\n\n")
	sb.WriteString("For detailed feature specifications, see the individual spec files referenced in the table above.\n")

	return sb.String()
}

// inferCategory tries to infer the category from feature ID or spec path.
func inferCategory(node *features.FeatureNode) string {
	id := strings.ToUpper(node.ID)

	// Architecture & Documentation
	if strings.HasPrefix(id, "ARCH_") || strings.HasPrefix(id, "DOCS_") {
		return "Architecture & Core"
	}

	// Core functionality
	if strings.HasPrefix(id, "CORE_") {
		return "Core Functionality"
	}

	// CLI commands
	if strings.HasPrefix(id, "CLI_") {
		return "CLI Commands"
	}

	// Providers
	if strings.HasPrefix(id, "PROVIDER_") {
		return "Providers"
	}

	// Migration engines
	if strings.HasPrefix(id, "MIGRATION_") {
		return "Migration Engines"
	}

	// Drivers
	if strings.HasPrefix(id, "DRIVER_") {
		return "Drivers"
	}

	// Default
	return "Other"
}
