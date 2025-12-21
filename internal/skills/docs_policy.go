package skills

import (
	"context"
	"fmt"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
)

// Feature: SKILLS_REGISTRY
// Spec: spec/skills/registry.md

type DocsPolicy struct {
	id     string
	checks []runner.Skill
}

func NewDocsPolicy() runner.Skill {
	return &DocsPolicy{
		id: "docs:policy",
		checks: []runner.Skill{
			NewDocsDocPatterns(),
			NewDocsHeaderComments(),
			NewDocsOrphanDocs(),
		},
	}
}

func (s *DocsPolicy) ID() string { return s.id }

func (s *DocsPolicy) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	var notes []string
	failed := false
	skipped := false

	for _, check := range s.checks {
		res := check.Run(ctx, deps)

		switch res.Status {
		case runner.StatusFail:
			failed = true
			notes = append(notes, fmt.Sprintf("FAIL [%s]: %s", check.ID(), oneLine(res.Note)))
		case runner.StatusSkip:
			skipped = true
			notes = append(notes, fmt.Sprintf("SKIP [%s]: %s", check.ID(), res.Note))
		case runner.StatusPass:
			// Optional: notes = append(notes, fmt.Sprintf("PASS [%s]", check.ID()))
		}
	}

	if failed {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 1,
			Note:     strings.Join(notes, "\n"),
		}
	}

	if skipped && len(notes) > 0 {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusPass,
			ExitCode: 0,
			Note:     fmt.Sprintf("Policy passed (with skips):\n%s", strings.Join(notes, "\n")),
		}
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
		Note:     "All policy checks passed.",
	}
}

func oneLine(s string) string {
	if idx := strings.Index(s, "\n"); idx != -1 {
		return s[:idx] + "..."
	}
	return s
}
