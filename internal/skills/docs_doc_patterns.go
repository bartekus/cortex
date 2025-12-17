package skills

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
	"github.com/bartekus/cortex/internal/scanner"
)

type DocsDocPatterns struct {
	id string
}

func NewDocsDocPatterns() runner.Skill {
	return &DocsDocPatterns{id: "docs:doc-patterns"}
}

func (s *DocsDocPatterns) ID() string { return s.id }

var fileNameRegex = regexp.MustCompile(`^[a-z0-9\-_]+\.md$`)

func (s *DocsDocPatterns) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// Scan docs/
	opts := scanner.FilterOptions{
		IncludeExtensions: []string{".md"},
	}
	files, err := deps.Scanner.TrackedFilesFiltered(ctx, opts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Scanner failed: %v", err),
		}
	}

	// Also check if docs/ exists? Or do we care?
	// If files is empty, maybe docs doesn't exist?
	// Just processing files found is safe.

	var failures []string

	// Updated Regex to allow Uppercase
	fileNameRegex := regexp.MustCompile(`^[A-Za-z0-9\-_]+\.md$`)

	for _, p := range files {
		// Only check docs/
		if !strings.HasPrefix(p, "docs/") {
			continue
		}

		// Check exclusions
		// 1. Hidden directories (segment starts with .)
		parts := strings.Split(p, string(filepath.Separator))
		isHidden := false
		for _, part := range parts {
			if strings.HasPrefix(part, ".") && part != "." && part != ".." {
				isHidden = true
				break
			}
		}
		if isHidden {
			continue
		}

		// 2. Specific ignores
		if strings.Contains(p, "/__generated__/") || strings.Contains(p, "/archive/") {
			continue
		}

		// 1. Filename naming
		base := filepath.Base(p)
		if !fileNameRegex.MatchString(base) {
			failures = append(failures, fmt.Sprintf("%s: invalid filename (must match [A-Za-z0-9-_]+\\.md)", p))
		}

		// 2. No spaces in path
		if strings.Contains(p, " ") {
			failures = append(failures, fmt.Sprintf("%s: path contains spaces", p))
		}

		// 3. No Untitled
		if strings.Contains(strings.ToLower(p), "untitled") {
			failures = append(failures, fmt.Sprintf("%s: filename contains 'untitled'", p))
		}

		// 4. No docs/docs/
		if strings.Contains(p, "docs/docs/") {
			failures = append(failures, fmt.Sprintf("%s: double nesting 'docs/docs/'", p))
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
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
		Note:     "No doc pattern violations found.",
	}
}
