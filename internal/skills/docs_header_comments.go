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

type DocsHeaderComments struct {
	id string
}

func NewDocsHeaderComments() runner.Skill {
	return &DocsHeaderComments{id: "docs:header-comments"}
}

func (s *DocsHeaderComments) ID() string { return s.id }

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

	// 2. Check Go SPDX headers
	for _, p := range goFiles {
		// "Required a line containing SPDX-License-Identifier: (exact prefix recommended)"
		// First ~5 lines.
		fullPath := filepath.Join(deps.RepoRoot, p)
		if err := checkSPDX(fullPath); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", p, err))
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

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
		Note:     "All checked files have correct headers.",
	}
}

func checkSPDX(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineCount := 0
	found := false
	for scanner.Scan() {
		lineCount++
		if lineCount > 5 {
			break
		}
		if strings.Contains(scanner.Text(), "SPDX-License-Identifier:") {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("missing SPDX-License-Identifier header")
	}
	return nil
}

func checkFrontmatter(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

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
