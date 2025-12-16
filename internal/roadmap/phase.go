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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// DetectPhases parses spec/features.yaml and returns a mapping of phase name
// to Phase, where phase names are derived from YAML comments immediately
// preceding feature entries.
//
// Rules (as exercised by tests):
//   - Phase names come from comment lines, e.g. "# Phase 0: Foundation".
//   - The last phase comment before a "- id:" line is used for that feature.
//   - If a feature appears before any phase comment, it is assigned to the
//     "Uncategorized" phase.
//   - Multiple phase comments before a feature: the last one wins.
//   - Invalid YAML must return an error.
//   - A missing file must return an error.
func DetectPhases(featuresPath string) (map[string]*Phase, error) {
	//nolint:gosec // G304: file path is from user input via CLI flag, validated
	data, err := os.ReadFile(featuresPath)
	if err != nil {
		return nil, fmt.Errorf("roadmap: read features file: %w", err)
	}

	// First, validate and decode the YAML into Features.
	var doc featureDocument
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("roadmap: parse features yaml: %w", err)
	}

	// Build a featureID -> phaseName mapping by scanning the raw file and
	// using comment lines as phase markers.
	featurePhase := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	currentPhase := "Uncategorized"

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Comment line: update current phase if non-empty after '#'.
		if strings.HasPrefix(trimmed, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if comment != "" {
				// Use the full comment text as the phase name.
				currentPhase = comment
			}
			continue
		}

		// Feature ID line: "- id: FOO_BAR"
		if strings.HasPrefix(trimmed, "- id:") {
			id := strings.TrimSpace(strings.TrimPrefix(trimmed, "- id:"))
			if id != "" {
				featurePhase[id] = currentPhase
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("roadmap: scan features yaml: %w", err)
	}

	// Now aggregate features into phases.
	phases := make(map[string]*Phase)

	for i := range doc.Features {
		f := &doc.Features[i]
		phaseName, ok := featurePhase[f.ID]
		if !ok || phaseName == "" {
			phaseName = "Uncategorized"
		}

		p, exists := phases[phaseName]
		if !exists {
			p = &Phase{Name: phaseName}
			phases[phaseName] = p
		}

		p.Features = append(p.Features, *f)
	}

	return phases, nil
}
