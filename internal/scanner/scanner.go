package scanner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// Scanner provides access to the repository's tracked files.
type Scanner struct {
	repoRoot string

	mu           sync.Mutex
	trackedCache []string
}

// New creates a new Scanner for the given repository root.
func New(repoRoot string) *Scanner {
	return &Scanner{
		repoRoot: repoRoot,
	}
}

// TrackedFiles returns all files tracked by git, caching the result for the instance lifetime.
// It respects .gitignore implicitly by asking git.
func (s *Scanner) TrackedFiles(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.trackedCache != nil {
		return s.trackedCache, nil
	}

	// git ls-files -z to avoid escaping issues
	cmd := exec.CommandContext(ctx, "git", "ls-files", "-z")
	cmd.Dir = s.repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}

	if len(out) == 0 {
		s.trackedCache = []string{}
		return s.trackedCache, nil
	}

	// -z separates by NUL bytes.
	// Trim trailing NUL if present
	sOut := strings.TrimSuffix(string(out), "\x00")

	files := strings.Split(sOut, "\x00")
	s.trackedCache = files
	return s.trackedCache, nil
}

// TrackedFilesFiltered returns tracked files matching the filter options.
func (s *Scanner) TrackedFilesFiltered(ctx context.Context, opts FilterOptions) ([]string, error) {
	all, err := s.TrackedFiles(ctx)
	if err != nil {
		return nil, err
	}
	return FilterFiles(all, opts), nil
}

// TrackedGoFiles returns only tracked .go files, applying default excludes.
func (s *Scanner) TrackedGoFiles(ctx context.Context) ([]string, error) {
	return s.TrackedFilesFiltered(ctx, FilterOptions{
		ExcludeDirs:       DefaultExcludeDirs(),
		IncludeExtensions: []string{".go"},
	})
}
