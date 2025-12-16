// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package specschema

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadAllSpecs walks the spec directory and loads all .md files with frontmatter.
func LoadAllSpecs(root string) ([]Spec, error) {
	var specs []Spec

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		spec, err := LoadSpec(path)
		if err != nil {
			return fmt.Errorf("failed to load spec %s: %w", path, err)
		}

		specs = append(specs, *spec)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort specs by path for deterministic output
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].Path < specs[j].Path
	})

	return specs, nil
}

// LoadSpec loads a single spec file and extracts its frontmatter.
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is from config, not user input
	if err != nil {
		return nil, err
	}

	frontmatter, err := ExtractFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	return &Spec{
		Path:        path,
		Frontmatter: *frontmatter,
	}, nil
}

// ExtractFrontmatter extracts YAML frontmatter from markdown content.
// Frontmatter is delimited by --- at the start of the file.
func ExtractFrontmatter(content string) (*SpecFrontmatter, error) {
	// Check if file starts with frontmatter delimiter
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("spec file does not start with YAML frontmatter (---)")
	}

	// Find the end of frontmatter
	lines := strings.Split(content, "\n")
	var frontmatterLines []string
	var foundEnd bool

	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			foundEnd = true
			break
		}
		frontmatterLines = append(frontmatterLines, lines[i])
	}

	if !foundEnd {
		return nil, fmt.Errorf("spec file frontmatter is not properly closed (missing ---)")
	}

	frontmatterYAML := strings.Join(frontmatterLines, "\n")

	var frontmatter SpecFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter YAML: %w", err)
	}

	return &frontmatter, nil
}
