package skills

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/features"
	"github.com/bartekus/cortex/internal/runner"
	"github.com/bartekus/cortex/internal/scanner"
)

type DocsOrphanSpecs struct {
	id string
}

func NewDocsOrphanSpecs() runner.Skill {
	return &DocsOrphanSpecs{id: "docs:orphan-specs"}
}

func (s *DocsOrphanSpecs) ID() string { return s.id }

func (s *DocsOrphanSpecs) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// 1. Repo Sensitivity
	// If spec/features.yaml does not exist, SKIP.
	featuresPath := filepath.Join(deps.RepoRoot, "spec", "features.yaml")

	// Check existence via simple Stat, or let LoadGraph fail?
	// Skip contract: "If a Skill's prerequisites ... are missing, it MUST return StatusSkip".
	// Features.yaml is a prerequisite for knowing what specs are referenced.
	// But spec/ directory existence is also a prerequisite for HAVING orphans.

	// We'll trust LoadGraph fails if missing, but we should check if file exists first to return SKIP.
	// Actually LoadGraph returns error.
	// Let's use filter to check essential files first?
	// Or just os.Stat featuresPath

	if _, err := features.LoadGraph(featuresPath); err != nil {
		// If file doesn't exist, skip.
		// If parsing fails, fail?
		// "Repo-sensitivity: if spec/features.yaml missing -> SKIP with note."
		// LoadGraph returns error wrapped. Check error string or Stat first.
		// Using scanner is safer to check existence? No, scanner is for lists.
		// os.Stat is fine.
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   "spec/features.yaml not found (or invalid)",
		}
	}

	// Double check validity - if valid, proceed.
	graph, err := features.LoadGraph(featuresPath)
	if err != nil {
		// Should have been caught above or transient error
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Failed to load features graph: %v", err),
		}
	}

	// 2. Build referencedSpecs set
	referencedSpecs := make(map[string]bool)
	for _, node := range graph.Nodes {
		if node.Spec != "" {
			// Normalize: forward slashes, trim leading ./
			clean := filepath.ToSlash(filepath.Clean(node.Spec))
			referencedSpecs[clean] = true
		}
	}

	// 3. Scan tracked files in spec/
	opts := scanner.FilterOptions{
		IncludeExtensions: []string{".md"},
		// We can't filter by dir in options yet (ExcludeDirs only).
		// So we get all .md and filter for "spec/".
	}

	allMdFiles, err := deps.Scanner.TrackedFilesFiltered(ctx, opts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Scanner failed: %v", err),
		}
	}

	var orphans []string

	// Exemptions
	isExempt := func(path string) bool {
		// spec/README.md or spec/**/README.md
		if strings.HasSuffix(strings.ToLower(path), "readme.md") {
			return true
		}
		// spec/features.yaml (not .md, but good to note)
		return false
	}

	for _, path := range allMdFiles {
		// path is relative to RepoRoot
		// Check if it's in spec/
		if !strings.HasPrefix(path, "spec/") {
			continue
		}

		// Clean path for comparison
		clean := filepath.ToSlash(filepath.Clean(path))

		if isExempt(clean) {
			continue
		}

		if !referencedSpecs[clean] {
			orphans = append(orphans, clean)
		}
	}

	// 4. Report
	if len(orphans) > 0 {
		sort.Strings(orphans)
		lines := []string{fmt.Sprintf("Found %d orphan specs:", len(orphans))}
		for _, o := range orphans {
			lines = append(lines, fmt.Sprintf("- %s", o))
		}

		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 1,
			Note:     strings.Join(lines, "\n"),
		}
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
		Note:     "No orphan specs found.",
	}
}
