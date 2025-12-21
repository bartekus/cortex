package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
	"github.com/bartekus/cortex/internal/specschema"
)

// Feature: SKILLS_REGISTRY
// Spec: spec/skills/registry.md

type DocsValidateSpec struct {
	id string
}

func NewDocsValidateSpec() runner.Skill {
	return &DocsValidateSpec{id: "docs:validate-spec"}
}

func (s *DocsValidateSpec) ID() string { return s.id }

func (s *DocsValidateSpec) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// Reuse logic from gov_spec_validate.go
	// Since we are running as a skill, we assume standard locations unless configured otherwise.
	// But skills should generally work on the standard structure.

	// Root defaults to "spec" in repo root.
	rootPath := filepath.Join(deps.RepoRoot, "spec")

	// Check if spec dir exists
	// We can't rely on LoadAllSpecs to handle absence gracefully if it just fails on root.
	if info, err := os.Stat(rootPath); err != nil || !info.IsDir() {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   "No spec directory found",
		}
	}

	specs, err := specschema.LoadAllSpecs(rootPath)
	if err != nil {
		// Repo-sensitive skip?
		// If "spec" dir doesn't exist?
		// specschema.LoadAllSpecs traverses.
		// If rootPath doesn't exist, it might fail.
		// Let's check existence first?
		// The command just says "no spec files found" if empty.
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4, // Execution error
			Note:     fmt.Sprintf("Failed to load specs: %v", err),
		}
	}

	if len(specs) == 0 {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   fmt.Sprintf("No spec files found in %s", rootPath),
		}
	}

	var notes []string
	status := runner.StatusPass
	exitCode := 0

	if err := specschema.ValidateAll(specs); err != nil {
		status = runner.StatusFail
		exitCode = 1
		notes = append(notes, fmt.Sprintf("Spec validation failed: %v", err))
	} else {
		notes = append(notes, fmt.Sprintf("Validated %d spec file(s)", len(specs)))
	}

	// Also validate integrity? The command has a flag --check-integrity.
	// The user said "docs:validate-spec" implementation.
	// The command `spec-validate` does integrity ONLY if requested.
	// But `docs:validate-spec` implies validating the specs.
	// Phase 3 plan lists `feature-integrity` as a separate check/skill?
	// Task list: "Implement docs:feature-integrity" is separate item #31.
	// So `docs:validate-spec` should probably JUST do schema validation.

	// Wait, the command `gov spec validate` is the umbrella?
	// No, the command `gov spec validate` corresponds to `docs:validate-spec`.
	// I will include integrity ONLY if it's the standard for this skill,
	// but based on "Implement docs:feature-integrity" being separate, I will skip integrity here.

	return runner.SkillResult{
		Skill:    s.id,
		Status:   status,
		ExitCode: exitCode,
		Note:     strings.Join(notes, "\n"),
	}
}
