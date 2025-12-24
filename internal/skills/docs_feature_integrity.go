package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartekus/cortex/internal/mapping"
	"github.com/bartekus/cortex/internal/runner"
)

// Feature: SKILLS_REGISTRY
// Spec: spec/skills/registry.md

type DocsFeatureIntegrity struct {
	id string
}

func NewDocsFeatureIntegrity() runner.Skill {
	return &DocsFeatureIntegrity{id: "docs:feature-integrity"}
}

func (s *DocsFeatureIntegrity) ID() string { return s.id }

func (s *DocsFeatureIntegrity) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// Repo-sensitivity: Skip if spec/features.yaml is absent
	featuresPath := filepath.Join(deps.RepoRoot, "spec", "features.yaml")
	if info, err := os.Stat(featuresPath); err != nil || info.IsDir() {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   "spec/features.yaml not found",
		}
	}

	opts := mapping.DefaultOptions()
	opts.RootDir = deps.RepoRoot

	report, err := mapping.Analyze(opts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Analysis failed: %v", err),
		}
	}

	// Deterministic output generation
	var lines []string
	lines = append(lines, fmt.Sprintf("Features analyzed: %d", len(report.Features)))
	lines = append(lines, fmt.Sprintf("Violations found: %d", len(report.Violations)))

	status := runner.StatusPass
	exitCode := 0

	if len(report.Violations) > 0 {
		status = runner.StatusFail
		exitCode = 1

		lines = append(lines, "")
		lines = append(lines, "Violations:")

		// Sort just in case (report should be sorted, but let's be safe)
		// actually report.Violations IS sorted by mapping.Analyze contract.

		for _, v := range report.Violations {
			// Format: [CODE] Feature (Path): Detail
			// or just Detail?
			// The cli uses: "    - Feature (Path): Detail" under header.
			// Let's use a flat list for skill note.

			loc := v.Path
			if loc == "" {
				loc = "<global>"
			}
			feat := v.Feature
			if feat == "" {
				feat = "<global>"
			}

			line := fmt.Sprintf("[%s] %s (%s): %s", v.Code, feat, loc, v.Detail)
			lines = append(lines, line)
		}
	} else {
		lines = append(lines, "Invariant holds.")
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   status,
		ExitCode: exitCode,
		Note:     strings.Join(lines, "\n"),
	}
}
