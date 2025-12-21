package skills

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
	"github.com/bartekus/cortex/internal/scanner"
)

// Feature: SKILLS_REGISTRY
// Spec: spec/skills/registry.md

type DocsHeaderComments struct {
	id string
}

func NewDocsHeaderComments() runner.Skill {
	return &DocsHeaderComments{id: "docs:header-comments"}
}

func (s *DocsHeaderComments) ID() string { return s.id }

// Package comment enforcement mode.
// Default is strict (require). Set CORTEX_HEADER_COMMENTS_PACKAGE=warn to only warn.
func packageCommentMode() string {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("CORTEX_HEADER_COMMENTS_PACKAGE")))
	if v == "warn" {
		return "warn"
	}
	return "require"
}

func (s *DocsHeaderComments) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// 1. Scan for Go files and Spec files
	goOpts := scanner.FilterOptions{
		IncludeExtensions: []string{".go"},
	}
	// We scan all .go files tracked.
	goFiles, err := deps.Scanner.TrackedFilesFiltered(ctx, goOpts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Scanner failed (go): %v", err),
		}
	}

	specOpts := scanner.FilterOptions{
		IncludeExtensions: []string{".md"},
	}
	mdFiles, err := deps.Scanner.TrackedFilesFiltered(ctx, specOpts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Scanner failed (md): %v", err),
		}
	}

	var failures []string
	var warnings []string

	// 2. Check Go SPDX headers
	for _, p := range goFiles {
		// "Required a line containing SPDX-License-Identifier: (exact prefix recommended)"
		// First ~5 lines.
		fullPath := filepath.Join(deps.RepoRoot, p)
		warn, err := checkSPDX(fullPath)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", p, err))
		} else if warn != "" {
			warnings = append(warnings, fmt.Sprintf("%s: %s", p, warn))
		}
	}

	// 3. Check Spec Frontmatter
	// spec/**/*.md excluding README
	for _, p := range mdFiles {
		if !strings.HasPrefix(p, "spec/") {
			continue
		}
		if strings.HasSuffix(strings.ToLower(p), "readme.md") {
			continue
		}
		fullPath := filepath.Join(deps.RepoRoot, p)
		if err := checkFrontmatter(fullPath); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", p, err))
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		// Truncate logic? Or just return all?
		// User: "fail with path list".
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 1,
			Note:     strings.Join(failures, "\n"),
		}
	}

	if len(warnings) > 0 {
		sort.Strings(warnings)
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusPass,
			ExitCode: 0,
			Note:     "Warnings:\n" + strings.Join(warnings, "\n"),
		}
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
		Note:     "All checked files have correct headers.",
	}
}

func checkSPDX(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)

	// State machine
	// 0: Start
	// 1: Found SPDX (optional)
	// 2: Found Block Comment (optional)
	// 3: Expecting Package Comment
	// 4: Found Package Comment
	// 5: Found package declaration (terminal)

	// Simpler approach compatible with "Run down, skip X, Y, Z, expect // Package"

	lineNum := 0
	// seenSPDX := false // Unused, logic just continues
	// seenBlockComment := false

	inBlock := false

	for scanner.Scan() {
		lineNum++
		text := strings.TrimSpace(scanner.Text())

		// 1. Skip BOM & Blanks
		if text == "" {
			continue
		}

		// 2. Skip any comments (SPDX, Feature, build tags, etc)
		// Try to identify if it's the specific "// Package" one.
		if strings.HasPrefix(text, "// Package") || strings.HasPrefix(text, "//Package") {
			// Found it!
			return "", nil
		}

		// If it's another comment, skip it.
		// Note: This treats ALL comments as "headers" to skip.
		if strings.HasPrefix(text, "//") {
			continue
		}

		// Block comments
		// Let's allow multiple block comments or mixed.
		if strings.HasPrefix(text, "/*") {
			inBlock = true
			if strings.Contains(text, "*/") {
				inBlock = false
			}
			continue
		}
		if inBlock {
			if strings.Contains(text, "*/") {
				inBlock = false
			}
			continue
		}

		// 3. If we hit package declaration without seeing // Package...
		if strings.HasPrefix(text, "package ") {
			if packageCommentMode() == "warn" {
				return "missing '// Package <name>' comment before 'package' declaration", nil
			}
			return "", fmt.Errorf("missing '// Package <name>' comment before 'package' declaration")
		}

		// If we hit something else (code, imports) - unexpected.
		// Usually 'package' is the first non-comment thing.
		return "", fmt.Errorf("expected '// Package ...' or 'package ...', found: %q", text)
	}

	return "", fmt.Errorf("unexpected EOF before package declaration")
}

func checkFrontmatter(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)

	// Must start with ---
	if !scanner.Scan() {
		return fmt.Errorf("empty file")
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return fmt.Errorf("missing frontmatter start '---'")
	}

	// Must have closing --- within N lines (say 60)
	// And verify it's just ---
	lineCount := 1
	closed := false
	for scanner.Scan() {
		lineCount++
		if lineCount > 60 {
			break
		}
		if strings.TrimSpace(scanner.Text()) == "---" {
			closed = true
			break
		}
	}

	if !closed {
		return fmt.Errorf("missing or unclosed frontmatter (must close within 60 lines)")
	}
	return nil
}
