// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package specschema provides tools for loading and validating spec file frontmatter.
package specschema

import (
	"path/filepath"
)

// SpecFrontmatter represents the YAML frontmatter structure for spec files.
type SpecFrontmatter struct {
	Feature string                 `yaml:"feature"`
	Version string                 `yaml:"version"`
	Status  string                 `yaml:"status"`
	Domain  string                 `yaml:"domain"`
	Inputs  SpecInputs             `yaml:"inputs"`
	Outputs SpecOutputs            `yaml:"outputs"`
	Extra   map[string]interface{} `yaml:",inline"`
}

// SpecInputs represents the inputs section of spec frontmatter.
type SpecInputs struct {
	Flags []CliFlag `yaml:"flags"`
}

// SpecOutputs represents the outputs section of spec frontmatter.
type SpecOutputs struct {
	ExitCodes map[string]int `yaml:"exit_codes"`
}

// CliFlag represents a CLI flag definition in spec frontmatter.
type CliFlag struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Default     string `yaml:"default"`
	Description string `yaml:"description"`
}

// Spec represents a loaded spec file with its frontmatter and path.
type Spec struct {
	Path        string
	Frontmatter SpecFrontmatter
}

// ExpectedFeatureIDFromPath extracts the expected feature ID from a spec file path.
// For example, "spec/governance/CLI_COMMAND_GOV.md" -> "CLI_COMMAND_GOV".
func ExpectedFeatureIDFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}
