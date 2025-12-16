// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: GOV_STATUS_ROADMAP
// Spec: spec/commands/status-roadmap.md

// Package roadmap provides tools for analyzing feature completion and generating
// roadmap status reports from spec/features.yaml.
package roadmap

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateMarkdown generates a deterministic markdown document from statistics and blockers.
func GenerateMarkdown(stats *Stats, blockers []*Blocker) string {
	var b strings.Builder

	// Top-level heading
	b.WriteString("# Feature Completion Analysis\n\n")
	b.WriteString("> **Source**: Generated from `spec/features.yaml` by `stagecraft status roadmap`\n")
	b.WriteString("> **Last Updated**: See `spec/features.yaml` for the source of truth\n>\n")
	b.WriteString("> **Note**: This document is automatically generated. To regenerate, run `stagecraft status roadmap`.\n\n")
	b.WriteString("⸻\n\n")

	// Executive Summary
	b.WriteString("## Executive Summary\n\n")
	fmt.Fprintf(&b, "- **Total Features**: %d\n", stats.Total)
	if stats.Total > 0 {
		fmt.Fprintf(&b, "- **Completed**: %d (%.1f%%)\n", stats.Done, float64(stats.Done)/float64(stats.Total)*100.0)
		fmt.Fprintf(&b, "- **In Progress**: %d (%.1f%%)\n", stats.WIP, float64(stats.WIP)/float64(stats.Total)*100.0)
		fmt.Fprintf(&b, "- **Planned**: %d (%.1f%%)\n", stats.Todo, float64(stats.Todo)/float64(stats.Total)*100.0)
	} else {
		b.WriteString("- **Completed**: 0\n")
		b.WriteString("- **In Progress**: 0\n")
		b.WriteString("- **Planned**: 0\n")
	}

	b.WriteString("\n⸻\n\n")

	// Phase-by-Phase Completion
	b.WriteString("## Phase-by-Phase Completion\n\n")
	b.WriteString("| Phase | Features | Done | WIP | Todo | Completion | Status |\n")
	b.WriteString("|-------|----------|------|-----|------|------------|--------|\n")

	phaseNames := sortedPhaseNames(stats.PhaseStats)
	for _, name := range phaseNames {
		ps := stats.PhaseStats[name]
		status := phaseStatusLabel(ps)
		fmt.Fprintf(
			&b,
			"| **%s** | %d | %d | %d | %d | %.0f%% | %s |\n",
			name,
			ps.Total,
			ps.Done,
			ps.WIP,
			ps.Todo,
			ps.CompletionPercentage,
			status,
		)
	}

	b.WriteString("\n⸻\n\n")

	// Roadmap Alignment
	b.WriteString("## Roadmap Alignment\n\n")
	b.WriteString("### Strong Progress\n\n")
	// Identify phases with good progress
	for _, name := range phaseNames {
		ps := stats.PhaseStats[name]
		if ps.CompletionPercentage >= 100.0 {
			fmt.Fprintf(&b, "- ✅ **%s Complete**: All features done (%d/%d)\n", name, ps.Done, ps.Total)
		} else if ps.CompletionPercentage >= 50.0 && ps.Done > 0 {
			fmt.Fprintf(&b, "- 🔄 **%s In Progress**: %.1f%% complete (%d/%d done", name, ps.CompletionPercentage, ps.Done, ps.Total)
			if ps.WIP > 0 {
				fmt.Fprintf(&b, ", %d wip", ps.WIP)
			}
			b.WriteString(")\n")
		}
	}

	b.WriteString("\n### Critical Gaps\n\n")
	// Identify phases with no progress
	for _, name := range phaseNames {
		ps := stats.PhaseStats[name]
		if ps.CompletionPercentage == 0.0 && ps.Total > 0 {
			fmt.Fprintf(&b, "- ⚠️ **%s**: 0%% complete — not started\n", name)
		}
	}

	b.WriteString("\n⸻\n\n")

	// Priority Recommendations
	b.WriteString("## Priority Recommendations\n\n")
	if stats.Total == 0 {
		b.WriteString("No features are defined in `spec/features.yaml`.\n\n")
	} else {
		b.WriteString("### 🔥 Immediate (Unblocks Other Work)\n\n")
		// List blockers that block other features
		if len(blockers) > 0 {
			for _, blk := range blockers {
				fmt.Fprintf(&b, "1. Complete `%s` to unblock dependent features\n", blk.FeatureID)
			}
		} else {
			b.WriteString("No immediate blockers detected.\n")
		}
		b.WriteString("\n")
	}

	// Detailed Phase Analysis
	b.WriteString("## Detailed Phase Analysis\n\n")
	for _, name := range phaseNames {
		ps := stats.PhaseStats[name]
		fmt.Fprintf(&b, "### %s\n\n", name)

		if ps.Total == 0 {
			b.WriteString("No features defined for this phase.\n\n")
			continue
		}

		fmt.Fprintf(
			&b,
			"- Features: %d (Done: %d, WIP: %d, Todo: %d)\n",
			ps.Total,
			ps.Done,
			ps.WIP,
			ps.Todo,
		)
		fmt.Fprintf(&b, "- Completion: %.1f%%\n\n", ps.CompletionPercentage)
	}

	// Critical Path Analysis
	b.WriteString("⸻\n\n")
	b.WriteString("## Critical Path Analysis\n\n")
	if len(blockers) == 0 {
		b.WriteString("No blocked features detected. All dependencies for non-done features are satisfied.\n\n")
	} else {
		b.WriteString("The following features are blocked by incomplete dependencies:\n\n")
		for _, blk := range blockers {
			fmt.Fprintf(&b, "- `%s` blocked by: %s\n", blk.FeatureID, strings.Join(blk.BlockedBy, ", "))
		}
		b.WriteString("\n")
	}

	// Next Steps
	b.WriteString("## Next Steps\n\n")
	b.WriteString("1. Use `stagecraft status roadmap` to regenerate this document whenever `spec/features.yaml` changes.\n")
	b.WriteString("2. Prioritize unblocking critical-path features.\n")
	b.WriteString("3. Complete partially implemented phases before starting new ones.\n")

	return b.String()
}

// sortedPhaseNames returns phase names in deterministic, roadmap-aligned order:
//
//  1. "Architecture & Documentation"
//  2. "Phase 0".."Phase 10" (numeric order)
//  3. "Governance"
//  4. All other phases in lexicographical order
func sortedPhaseNames(phases map[string]*PhaseStats) []string {
	names := make([]string, 0, len(phases))
	for name := range phases {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		gi, ni, si := phaseSortKey(names[i])
		gj, nj, sj := phaseSortKey(names[j])

		if gi != gj {
			return gi < gj
		}
		if ni != nj {
			return ni < nj
		}
		return si < sj
	})

	return names
}

// phaseSortKey assigns a sort group and numeric index (for Phase N) to a phase name.
func phaseSortKey(name string) (group, num int, s string) {
	switch {
	case strings.HasPrefix(name, "Architecture"):
		return 0, 0, name
	case strings.HasPrefix(name, "Phase "):
		// Attempt to parse the phase number between "Phase " and ":".
		rest := strings.TrimPrefix(name, "Phase ")
		colonIdx := strings.Index(rest, ":")
		if colonIdx >= 0 {
			rest = rest[:colonIdx]
		}
		var n int
		_, _ = fmt.Sscanf(rest, "%d", &n) // best-effort; n defaults to 0 on failure
		return 1, n, name
	case strings.HasPrefix(name, "Governance"):
		return 2, 0, name
	default:
		return 3, 0, name
	}
}

// phaseStatusLabel derives a human-readable status label for a phase.
func phaseStatusLabel(ps *PhaseStats) string {
	switch {
	case ps.Total == 0:
		return "⚠️ Not started"
	case ps.CompletionPercentage >= 100.0:
		return "✅ Complete"
	case ps.CompletionPercentage > 0:
		return "🔄 In progress"
	default:
		return "⚠️ Not started"
	}
}
