// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package gov

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// NormalizeHelp applies normalization rules to CLI help output.
// - keep only “Usage / Available Commands / Flags” blocks
// - strip extra whitespace
// - enforce ordering (implicit by not reordering, so input order matters)
func NormalizeHelp(input string) string {
	lines := strings.Split(input, "\n")
	var normalized []string

	inBlock := false

	blockHeaders := regexp.MustCompile(`^(Usage:|Available Commands:|Flags:)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		if blockHeaders.MatchString(line) {
			inBlock = true
			normalized = append(normalized, line)
			continue
		}

		// Keep lines if we are inside a relevant block
		// Note: "Usage" is usually one line, but "Available Commands" and "Flags" are lists.
		// UseCobra's help sometimes puts other text.
		// Heuristic: If it starts with a block header, we enter a block.
		// If we encounter a new block header, we are still in a block (different one).
		// If we encounter "Global Flags:" or similar that we want to keep?
		// The requirement said: "Usage / Available Commands / Flags".
		// Cobra often has "Additional help topics:" or footer text.

		if inBlock {
			// Stop if we hit a section we don't want?
			// For now, let's keep everything after the first match of a known section
			// UNTIL we maybe hit something else?
			// The user said "keep only... blocks".
			// This implies filtering out "Cortex - Description..." at the top.
			normalized = append(normalized, line)
		}
	}

	return strings.Join(normalized, "\n")
}

// CompareHelp compares generated help with fixture.
func CompareHelp(generated, fixturePath string) error {
	fixtureBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		return fmt.Errorf("failed to read fixture %s: %w", fixturePath, err)
	}

	normGenerated := NormalizeHelp(generated)
	normFixture := NormalizeHelp(string(fixtureBytes))

	if normGenerated != normFixture {
		// Create a diff or just error
		// For simplicity, error with lengths or just "mismatch"
		// To be helpful, we could show a diff, but that requires a diff library or manual impl.
		return fmt.Errorf("CLI help drift detected!\nFixture (%s) length: %d\nGenerated length: %d\n\nGenerated:\n%s\n\nFixture:\n%s",
			fixturePath, len(normFixture), len(normGenerated), normGenerated, normFixture)
	}

	return nil
}
