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

type Purity struct {
	id string
}

func NewPurity() runner.Skill {
	return &Purity{id: "purity"}
}

func (s *Purity) ID() string { return s.id }

// Banned imports configuration
// Map of "banned package" -> list of allowed directories prefixes
var bannedImports = map[string][]string{
	"os/exec": {
		"cmd/",                // Allowed in commands
		"internal/skills/",    // Allowed in skills
		"internal/scanner/",   // Allowed in scanner (git ls-files)
		"internal/runner/",    // Allowed if runner needs to exec (maybe?)
		"internal/git/",       // Allowed for git operations
		"pkg/executil/",       // Allowed: core exec utility
		"test/e2e/",           // Allowed: e2e tests
		"internal/dev/",       // Allowed: dev tooling
		"internal/providers/", // Allowed: local providers
		// "tools/", // If we had tools
	},
	// Add others if needed: "syscall", "unsafe"
}

func (s *Purity) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// 1. Scan tracked Go files
	goOpts := scanner.FilterOptions{
		IncludeExtensions: []string{".go"},
		// Exclude tests? Usually logic purity applies to tests too, but maybe less strict.
		// For now, scan all.
	}
	files, err := deps.Scanner.TrackedFilesFiltered(ctx, goOpts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Scanner failed: %v", err),
		}
	}

	if len(files) == 0 {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   "No Go files found",
		}
	}

	var violations []string

	for _, p := range files {
		// Clean path
		p = filepath.ToSlash(p) // normalized

		imports, err := scanImports(filepath.Join(deps.RepoRoot, p))
		if err != nil {
			violations = append(violations, fmt.Sprintf("%s: failed to scan imports: %v", p, err))
			continue
		}

		for imp := range imports {
			if allowedDirs, banned := bannedImports[imp]; banned {
				// Check if P is in allowedDirs
				allowed := false
				for _, dir := range allowedDirs {
					if strings.HasPrefix(p, dir) {
						allowed = true
						break
					}
				}
				if !allowed {
					violations = append(violations, fmt.Sprintf("%s: banned import %q", p, imp))
				}
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 1,
			Note:     strings.Join(violations, "\n"),
		}
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
		Note:     "No banned imports found.",
	}
}

// scanImports scans a file for import "..." lines.
// Does a simple text scan.
// Handles:
// import "fmt"
// import (
//
//	"fmt"
//	alias "fmt"
//	. "fmt"
//
// )
func scanImports(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	imports := make(map[string]bool)
	scanner := bufio.NewScanner(f)

	inImportBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Single line import
		if strings.HasPrefix(line, "import ") && !strings.Contains(line, "(") {
			// import "fmt"
			// import alias "fmt"
			// Extract string quote
			if pkg := extractImport(line); pkg != "" {
				imports[pkg] = true
			}
			continue
		}

		// Block import start
		if strings.HasPrefix(line, "import (") {
			inImportBlock = true
			continue
		}

		// Block import end
		if inImportBlock && strings.HasPrefix(line, ")") {
			inImportBlock = false
			continue
		}

		// Inside block
		if inImportBlock {
			if pkg := extractImport(line); pkg != "" {
				imports[pkg] = true
			}
		}
	}

	return imports, scanner.Err()
}

func extractImport(line string) string {
	// Simple extractor: find content between first and last quotes
	mq := strings.Index(line, "\"")
	if mq == -1 {
		return ""
	}
	lastrem := strings.LastIndex(line, "\"")
	if lastrem <= mq {
		return ""
	}
	return line[mq+1 : lastrem]
}
