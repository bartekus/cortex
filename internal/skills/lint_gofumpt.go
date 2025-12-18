package skills

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
)

type LintGofumpt struct{}

func (s *LintGofumpt) ID() string {
	return "lint:gofumpt"
}

func (s *LintGofumpt) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// 1. Determine files to check
	var files []string
	var err error

	if len(deps.TargetFiles) > 0 {
		for _, f := range deps.TargetFiles {
			if strings.HasSuffix(f, ".go") {
				files = append(files, f)
			}
		}
	} else {
		files, err = deps.Scanner.TrackedGoFiles(ctx)
		if err != nil {
			return runner.SkillResult{
				Skill:    s.ID(),
				Status:   runner.StatusFail,
				ExitCode: 4,
				Note:     fmt.Sprintf("Failed to list files: %v", err),
			}
		}
	}

	if len(files) == 0 {
		return runner.SkillResult{
			Skill:    s.ID(),
			Status:   runner.StatusPass,
			ExitCode: 0,
			Note:     "No Go files to check",
		}
	}

	// 2. Check if gofumpt is installed
	if _, err := exec.LookPath("gofumpt"); err != nil {
		return runner.SkillResult{
			Skill:    s.ID(),
			Status:   runner.StatusFail,
			ExitCode: 2,
			Note:     "gofumpt not found. Run: go install mvdan.cc/gofumpt@v0.6.0",
		}
	}

	// 3. Run gofumpt -l <files>
	// Chunking to avoid ARG_MAX
	const batchSize = 200
	var unformatted []string

	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		cmd := exec.CommandContext(ctx, "gofumpt", append([]string{"-l"}, batch...)...)
		out, err := cmd.Output()
		// Exit status 0 means success (found/not found doesn't change exit code for -l usually,
		// but gofumpt -l prints names of unformatted files).
		// Wait, gofumpt -l returns 0 always unless error.
		if err != nil {
			return runner.SkillResult{
				Skill:    s.ID(),
				Status:   runner.StatusFail,
				ExitCode: 4,
				Note:     fmt.Sprintf("gofumpt execution failed: %v", err),
			}
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				unformatted = append(unformatted, line)
			}
		}
	}

	if len(unformatted) > 0 {
		sort.Strings(unformatted)
		unformatted = unique(unformatted)

		var msg bytes.Buffer
		msg.WriteString("Unformatted files:\n")
		for _, f := range unformatted {
			msg.WriteString(f + "\n")
		}
		msg.WriteString("\nTo fix, run:\n  gofumpt -w " + strings.Join(unformatted, " "))

		return runner.SkillResult{
			Skill:    s.ID(),
			Status:   runner.StatusFail,
			ExitCode: 3,
			Note:     msg.String(),
		}
	}

	return runner.SkillResult{
		Skill:    s.ID(),
		Status:   runner.StatusPass,
		ExitCode: 0,
	}
}

func unique(slice []string) []string {
	if len(slice) == 0 {
		return nil
	}
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
