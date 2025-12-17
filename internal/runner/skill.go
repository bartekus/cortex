package runner

import (
	"context"

	"github.com/bartekus/cortex/internal/scanner"
)

// Deps contains dependencies injected into skills.
type Deps struct {
	RepoRoot      string
	StateDir      string
	Scanner       *scanner.Scanner
	FailOnWarning bool
	TargetFiles   []string // Files to process (if empty, process all tracked files)
	// Add other deps like Registry later
}

// Skill defines a unit of work in the migration runner.
type Skill interface {
	// ID returns the unique identifier (e.g. "lint:gofumpt").
	ID() string

	// Run executes the skill.
	Run(ctx context.Context, deps *Deps) SkillResult
}
