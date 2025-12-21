package skills

import (
	"context"
	"os/exec"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
)

// Feature: SKILLS_REGISTRY
// Spec: spec/skills/registry.md

type LintGolangCI struct{}

func (s *LintGolangCI) ID() string {
	return "lint:golangci"
}

func (s *LintGolangCI) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// 1. Check if golangci-lint is installed
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return runner.SkillResult{
			Skill:    s.ID(),
			Status:   runner.StatusFail,
			ExitCode: 2,
			Note:     "golangci-lint not found. Run: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2",
		}
	}

	// 2. Run golangci-lint run ./...
	cmd := exec.CommandContext(ctx, "golangci-lint", "run", "./...")
	cmd.Dir = deps.RepoRoot
	// Capture output to return in Note if failed, or just let it print to stdout?
	// The runner handles printing "Note", but for a linter, the output IS the note.
	// But golangci-lint output can be huge.
	// Scripts/run.sh just lets it run.
	// But our runner captures execution.
	// We should probably capture combined output.

	// Wait, internal/runner/runner.go prints:
	// fmt.Printf("FAIL: %s (exit %d)\n", id, res.ExitCode)
	// if res.Note != "" { fmt.Println(res.Note) }

	out, err := cmd.CombinedOutput()
	if err != nil {
		// Distinguish execution error (4) vs lint failure (3)?
		// golangci-lint returns 1 by default for issues.
		// User requested: "4 execution error", "3 lint failures (if you want mapping)".
		// But golangci-lint exit codes:
		// 1: issues found
		// 3: generic error
		// 4: timeout
		// Let's stick to simple: if err, fail.
		// If exit code is 1, maybe map to 3? Or just keep 1?
		// Existing run.sh: just returns exit code.
		// I will pass through the exit code if possible, or mapping.

		var exitCode int
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 4 // unable to run
		}

		// If it's a lint failure, we want to show the output.
		note := strings.TrimSpace(string(out))

		return runner.SkillResult{
			Skill:    s.ID(),
			Status:   runner.StatusFail,
			ExitCode: exitCode,
			Note:     note,
		}
	}

	return runner.SkillResult{
		Skill:    s.ID(),
		Status:   runner.StatusPass,
		ExitCode: 0,
	}
}
