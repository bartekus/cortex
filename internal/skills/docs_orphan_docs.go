package skills

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
	"github.com/bartekus/cortex/internal/scanner"
)

type DocsOrphanDocs struct {
	id string
}

func NewDocsOrphanDocs() runner.Skill {
	return &DocsOrphanDocs{id: "docs:orphan-docs"}
}

func (s *DocsOrphanDocs) ID() string { return s.id }

// Link regex: [text](target)
// We only care about the target (submatch 2).
// Note: we intentionally keep this simple (no nested parens support).
// We use the index-based API so we can ignore image links like ![alt](img.png).
var linkRegex = regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)

func (s *DocsOrphanDocs) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// 1. Repo Sensitivity: If docs/ directory is missing, SKIP.
	docsDir := filepath.Join(deps.RepoRoot, "docs")
	if info, err := os.Stat(docsDir); err != nil || !info.IsDir() {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   "docs directory not found",
		}
	}

	// 2. Identify Candidates: tracked docs/**/*.md
	// Exclude archive, generated?
	// User said "exclude docs/archive/ or other intentional dead zones".
	candidateOpts := scanner.FilterOptions{
		IncludeExtensions: []string{".md"},
	}
	allFiles, err := deps.Scanner.TrackedFilesFiltered(ctx, candidateOpts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Scanner failed: %v", err),
		}
	}

	candidates := make(map[string]bool)
	docSources := []string{}

	// Filter candidates (only docs/**/*.md, not ignored)
	for _, p := range allFiles {
		if strings.HasPrefix(p, "docs/") {
			// Candidates must be in docs/
			// Exclude docs/archive/
			if strings.HasPrefix(p, "docs/archive/") {
				continue
			}
			candidates[p] = true
			docSources = append(docSources, p)
		} else if strings.HasPrefix(p, "spec/") {
			// Specs are sources of references too
			docSources = append(docSources, p)
		}
	}

	if len(candidates) == 0 {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusPass, // No docs to be orphans
			Note:   "No docs candidates found",
		}
	}

	// 3. Scan references
	referencedDocs := make(map[string]bool)

	for _, src := range docSources {
		srcPath := filepath.Join(deps.RepoRoot, src)
		// We need relative resolution.
		// If src is "docs/guide.md" and links to "setup.md", it means "docs/setup.md".
		srcDir := path.Dir(src) // use path (forward slash) as we cleaned paths from scanner?
		// Scanner returns result of git ls-files, mostly forward slash on Mac/Linux, but let's be safe.
		// Actually scanner.FilterFiles sorts strings.
		// Let's ensure forward slashes for math.
		srcDir = filepath.ToSlash(srcDir)

		f, err := os.Open(srcPath)
		if err != nil {
			// Warn or skip?
			// Fail for now as it's unexpected for a tracked file
			return runner.SkillResult{
				Skill:    s.id,
				Status:   runner.StatusFail,
				ExitCode: 4,
				Note:     fmt.Sprintf("Failed to read %s: %v", src, err),
			}
		}

		scn := bufio.NewScanner(f)
		// Increase buffer to handle long markdown lines (default 64K can be too small).
		scn.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		inCodeFence := false

		for scn.Scan() {
			line := scn.Text()

			// Skip fenced code blocks to avoid counting example links.
			// We treat any line starting with ``` as a fence toggle.
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inCodeFence = !inCodeFence
				continue
			}
			if inCodeFence {
				continue
			}

			// Use index-based matching so we can ignore image links.
			matches := linkRegex.FindAllStringSubmatchIndex(line, -1)
			for _, mi := range matches {
				// mi contains pairs: full match, group1, group2.
				// Ignore image links: ![alt](...)
				if mi[0] > 0 && line[mi[0]-1] == '!' {
					continue
				}

				// Extract group 2 (target)
				if len(mi) < 6 {
					continue
				}
				target := line[mi[4]:mi[5]]
				// Ignore external
				if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "/") {
					continue
				}

				// Strip anchors/queries
				if idx := strings.Index(target, "#"); idx != -1 {
					target = target[:idx]
				}
				if idx := strings.Index(target, "?"); idx != -1 {
					target = target[:idx]
				}

				target = strings.TrimSpace(target)
				if target == "" {
					continue
				}

				// Ignore non-markdown / non-file targets quickly.
				// (We only consider .md links for orphan-doc detection.)
				if !strings.HasSuffix(target, ".md") {
					continue
				}

				// Resolve path.
				// path.Join cleans and is slash-stable.
				resolved := path.Clean(path.Join(srcDir, target))

				// Only record links into docs/.
				if strings.HasPrefix(resolved, "docs/") {
					referencedDocs[resolved] = true
				}
			}
		}
		if err := scn.Err(); err != nil {
			_ = f.Close()
			return runner.SkillResult{
				Skill:    s.id,
				Status:   runner.StatusFail,
				ExitCode: 4,
				Note:     fmt.Sprintf("Failed to scan %s: %v", src, err),
			}
		}
		f.Close()
	}

	// 4. Calculate Orphans
	var orphans []string

	// Add explicit exemptions
	// READMEs in docs/ are entries, usually.
	isExempt := func(p string) bool {
		if strings.HasSuffix(strings.ToLower(p), "readme.md") {
			return true
		}
		return false
	}

	for c := range candidates {
		if isExempt(c) {
			continue
		}
		if !referencedDocs[c] {
			orphans = append(orphans, c)
		}
	}

	if len(orphans) > 0 {
		sort.Strings(orphans)
		lines := []string{fmt.Sprintf("Found %d orphan docs (not referenced by other docs/specs):", len(orphans))}
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
		Note:     "No orphan docs found.",
	}
}
