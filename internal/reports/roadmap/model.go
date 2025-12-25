// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CLI_COMMAND_STATUS
// Spec: spec/cli/status.md

package roadmap

// Feature represents a single feature entry from spec/features.yaml.
type Feature struct {
	ID             string   `yaml:"id"`
	Title          string   `yaml:"title"`
	Governance     string   `yaml:"governance"`
	Implementation string   `yaml:"implementation"`
	Spec           string   `yaml:"spec"`
	Owner          string   `yaml:"owner"`
	DependsOn      []string `yaml:"depends_on"`
	Tests          []string `yaml:"tests"`
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
