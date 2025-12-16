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

// Feature represents a single feature entry from spec/features.yaml.
type Feature struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	Spec      string   `yaml:"spec"`
	Owner     string   `yaml:"owner"`
	DependsOn []string `yaml:"depends_on"`
	Tests     []string `yaml:"tests"`
}

// featureDocument matches the top-level shape of spec/features.yaml for YAML decoding.
type featureDocument struct {
	Features []Feature `yaml:"features"`
}

// Phase groups features under a human-readable phase name.
type Phase struct {
	Name     string
	Features []Feature
}

// Stats represents overall and per-phase statistics.
type Stats struct {
	Total                int
	Done                 int
	WIP                  int
	Todo                 int
	CompletionPercentage float64
	PhaseStats           map[string]*PhaseStats
}

// PhaseStats represents statistics for a single phase.
type PhaseStats struct {
	Total                int
	Done                 int
	WIP                  int
	Todo                 int
	CompletionPercentage float64
}

// Blocker represents a feature blocked by incomplete dependencies.
type Blocker struct {
	FeatureID string
	BlockedBy []string
}
