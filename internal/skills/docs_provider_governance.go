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
		// e.g. spec/providers/backend/encore-ts.md -> docs/providers/backend/encore-ts.md

		// Fallbacks:
		// 1. docs/providers/<relPath>
		// 2. docs/providers/<category>.md (if <relPath> is <category>/<name>.md -> NO, wait)
		//    User said: "docs/providers/<category>.md (category index)"
		//    If spec is spec/providers/backend/encore-ts.md, category is backend.
		// 3. docs/providers/<category>/<name>/README.md
		// 4. docs/providers/<category>/<name>/index.md

		candidates := []string{}

		// 1. Direct mirror
		candidates = append(candidates, filepath.Join("docs", "providers", relPath))

		// Logic to extract category and name
		// relPath = backend/encore-ts.md
		dir := filepath.Dir(relPath)               // backend
		nameExt := filepath.Base(relPath)          // encore-ts.md
		name := strings.TrimSuffix(nameExt, ".md") // encore-ts

		// 2. Category Index? Only if it's a "provider" inside a "category".
		// If spec is spec/providers/aws.md, dir=".", name="aws". Category index docs/providers.md? Maybe.
		// User example: spec/providers/backend/encore-ts.md -> docs/providers/backend.md
		if dir != "." {
			categoryDoc := filepath.Join("docs", "providers", dir+".md")
			candidates = append(candidates, categoryDoc)
		}

		// 3. Folder README
		// docs/providers/backend/encore-ts/README.md
		candidates = append(candidates, filepath.Join("docs", "providers", dir, name, "README.md"))

		// 4. Folder Index
		// docs/providers/backend/encore-ts/index.md
		candidates = append(candidates, filepath.Join("docs", "providers", dir, name, "index.md"))

		// 5. Root Name Match
		// spec/providers/integration/terraform.md -> docs/providers/terraform.md
		candidates = append(candidates, filepath.Join("docs", "providers", name+".md"))

		found := false
		for _, c := range candidates {
			if _, err := os.Stat(filepath.Join(deps.RepoRoot, c)); err == nil {
				found = true
				break
			}
		}

		if !found {
			missingDocs = append(missingDocs, fmt.Sprintf("Provider spec %s missing doc. Tried: %v", p, candidates))
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
