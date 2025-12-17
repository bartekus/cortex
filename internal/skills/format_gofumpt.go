package skills

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
)

type FormatGofumpt struct{}

func (s *FormatGofumpt) ID() string {
	return "format:gofumpt"
}

func (s *FormatGofumpt) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
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
			Note:     "No Go files to format",
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

	// 3. Run gofumpt -w <files>
	const batchSize = 200
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]
		cmd := exec.CommandContext(ctx, "gofumpt", append([]string{"-w"}, batch...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return runner.SkillResult{
				Skill:    s.ID(),
				Status:   runner.StatusFail,
				ExitCode: 4,
				Note:     fmt.Sprintf("gofumpt failed: %v\n%s", err, string(out)),
			}
		}
	}

	return runner.SkillResult{
		Skill:    s.ID(),
		Status:   runner.StatusPass,
		ExitCode: 0,
	}
}
