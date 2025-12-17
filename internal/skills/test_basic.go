package skills

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
)

// Generic Exec Skill
type ExecSkill struct {
	id   string
	args []string
	// Env etc
}

func (s *ExecSkill) ID() string { return s.id }

func (s *ExecSkill) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	cmd := exec.CommandContext(ctx, s.args[0], s.args[1:]...)
	cmd.Dir = deps.RepoRoot

	out, err := cmd.CombinedOutput()
	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}

		// Capture last N lines of output for note?
		// Or all of it?
		// User said: "failures become StatusFail with a useful note (include last N lines of stderr)"

		output := string(out)
		lines := strings.Split(output, "\n")
		// Keep last 20 lines
		if len(lines) > 20 {
			lines = lines[len(lines)-20:]
			output = "...(truncated)...\n" + strings.Join(lines, "\n")
		}

		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: exitCode,
			Note:     strings.TrimSpace(output),
		}
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
	}
}

func NewTestBuild() runner.Skill {
	return &ExecSkill{
		id:   "test:build",
		args: []string{"go", "build", "./..."},
	}
}

func NewTestBinary() runner.Skill {
	// Logic:
	// if cmd/cortex exists -> build it (bin/cortex)
	// else if cmd/cortex exists -> build it (bin/cortex)
	// else SKIP
	return &SmartBinarySkill{id: "test:binary"}
}

type SmartBinarySkill struct {
	id string
}

func (s *SmartBinarySkill) ID() string { return s.id }

func (s *SmartBinarySkill) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// Check for Cortex
	// We need absolute path? deps.RepoRoot is absolute.
	// But we use relative paths for go build typically?
	// go build -o bin/foo ./cmd/foo works from module root.

	// Check existence
	// Check existence
	cortexPath := "cmd/cortex"

	// We need to check if directory exists relative to RepoRoot
	// We can use os.Stat

	hasCortex := false
	if info, err := os.Stat(deps.RepoRoot + "/" + cortexPath); err == nil && info.IsDir() {
		hasCortex = true
	}

	var args []string

	if hasCortex {
		args = []string{"go", "build", "-o", "bin/cortex", "./cmd/cortex"}
	} else {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   "No cmd/cortex found",
		}
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = deps.RepoRoot

	out, err := cmd.CombinedOutput()
	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}

		output := string(out)
		lines := strings.Split(output, "\n")
		// Keep last 20 lines
		if len(lines) > 20 {
			lines = lines[len(lines)-20:]
			output = "...(truncated)...\n" + strings.Join(lines, "\n")
		}

		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: exitCode,
			Note:     strings.TrimSpace(output),
		}
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
		Note:     "Built " + args[3],
	}
}

func NewTestGo() runner.Skill {
	return &ExecSkill{
		id:   "test:go",
		args: []string{"go", "test", "./..."},
	}
}
