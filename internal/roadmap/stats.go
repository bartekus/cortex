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

import "sort"

// CalculateStats calculates overall and per-phase statistics from the given phases.
func CalculateStats(phases map[string]*Phase) *Stats {
	stats := &Stats{
		PhaseStats: make(map[string]*PhaseStats),
	}

	for phaseName, phase := range phases {
		ps := &PhaseStats{}

		for i := range phase.Features {
			f := &phase.Features[i]
			ps.Total++
			stats.Total++

			switch f.Status {
			case "done":
				ps.Done++
				stats.Done++
			case "wip":
				ps.WIP++
				stats.WIP++
			default:
				// Treat any non-done, non-wip status as todo.
				ps.Todo++
				stats.Todo++
			}
		}

		if ps.Total > 0 {
			ps.CompletionPercentage = float64(ps.Done) / float64(ps.Total) * 100.0
		}

		stats.PhaseStats[phaseName] = ps
	}

	if stats.Total > 0 {
		stats.CompletionPercentage = float64(stats.Done) / float64(stats.Total) * 100.0
	}

	return stats
}

// IdentifyBlockers identifies features that are blocked by incomplete dependencies.
// A feature is considered blocked if:
//   - Its own status is not "done", and
//   - At least one of its dependencies is either missing or not "done".
func IdentifyBlockers(phases map[string]*Phase) []*Blocker {
	// Index all features by ID for quick lookup.
	index := make(map[string]Feature)
	for _, phase := range phases {
		for i := range phase.Features {
			f := &phase.Features[i]
			index[f.ID] = *f
		}
	}

	var blockers []*Blocker

	for featureID := range index {
		f := index[featureID]
		if f.Status == "done" {
			continue
		}

		if len(f.DependsOn) == 0 {
			continue
		}

		var blockedBy []string
		for _, depID := range f.DependsOn {
			dep, ok := index[depID]
			// If dependency is missing or not done, treat it as blocking.
			if !ok || dep.Status != "done" {
				blockedBy = append(blockedBy, depID)
			}
		}

		if len(blockedBy) > 0 {
			blockers = append(blockers, &Blocker{
				FeatureID: f.ID,
				BlockedBy: blockedBy,
			})
		}
	}

	// Deterministic ordering of blockers.
	sort.Slice(blockers, func(i, j int) bool {
		return blockers[i].FeatureID < blockers[j].FeatureID
	})

	return blockers
}
