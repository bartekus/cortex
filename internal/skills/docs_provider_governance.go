package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
	"github.com/bartekus/cortex/internal/scanner"
)

type DocsProviderGovernance struct {
	id string
}

func NewDocsProviderGovernance() runner.Skill {
	return &DocsProviderGovernance{id: "docs:provider-governance"}
}

func (s *DocsProviderGovernance) ID() string { return s.id }

func (s *DocsProviderGovernance) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// 1. Repo Sensitivity
	// If spec/providers does not exist, SKIP.
	providersSpecDir := filepath.Join(deps.RepoRoot, "spec", "providers")
	if info, err := os.Stat(providersSpecDir); err != nil || !info.IsDir() {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   "spec/providers directory not found",
		}
	}

	// 2. Scan for provider specs
	opts := scanner.FilterOptions{
		IncludeExtensions: []string{".md"},
	}
	allFiles, err := deps.Scanner.TrackedFilesFiltered(ctx, opts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Scanner failed: %v", err),
		}
	}

	var missingDocs []string

	for _, p := range allFiles {
		// Only check spec/providers/*.md
		if !strings.HasPrefix(p, "spec/providers/") {
			continue
		}

		// Exclude README.md
		if strings.HasSuffix(strings.ToLower(p), "readme.md") {
			continue
		}

		// Extract relative path from spec/providers/
		// e.g. spec/providers/aws.md -> aws.md
		relPath := p[len("spec/providers/"):]

		// Check for corresponding doc
		// Expected: docs/providers/<relPath>
		expectedDoc := filepath.Join("docs", "providers", relPath)

		// Verify existence
		// We can use os.Stat (simple existence) or check if it's in allFiles (tracked existence).
		// Governance usually implies tracked existence.
		// Since we have allFiles, let's use that if possible, but searching a slice is O(N).
		// Let's use os.Stat for simplicity, or build a map if N is large. Map is safer.
		// Actually, let's assume if it exists on disk it's fine, but tracked is better.
		// I'll assume tracked. Let's create a map of docs files.

		// Wait, I can just os.Stat(filepath.Join(deps.RepoRoot, expectedDoc)).
		// If it exists but untracked, another check (orphan-docs) might complain or not.
		// Let's rely on os.Stat for existence.
		if _, err := os.Stat(filepath.Join(deps.RepoRoot, expectedDoc)); os.IsNotExist(err) {
			missingDocs = append(missingDocs, fmt.Sprintf("Provider spec %s missing doc %s", p, expectedDoc))
		}
	}

	if len(missingDocs) > 0 {
		sort.Strings(missingDocs)
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 1,
			Note:     strings.Join(missingDocs, "\n"),
		}
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
		Note:     "All provider specs have matching docs.",
	}
}
