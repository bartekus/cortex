package skills

import (
	"context"

	"github.com/bartekus/cortex/internal/runner"
)

// Registry defines the canonical order of skills.
// Even if not implemented, they are listed here to maintain stable interface.

// Stub skills validation or placeholders could go here if needed.
// For now we just put the implemented ones.
// The user provided a specific list, let's implement placeholders for them so "run list" matches the plan?
// The user said: "Even if some are not implemented yet, keep them in the list and return skip with a note".

// We need a generic placeholder skill then.

type PlaceholderSkill struct {
	id string
}

func (s *PlaceholderSkill) ID() string { return s.id }
func (s *PlaceholderSkill) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	return runner.SkillResult{
		Skill:  s.id,
		Status: runner.StatusSkip,
		Note:   "Not implemented yet",
	}
}

func newPlaceholder(id string) runner.Skill {
	return &PlaceholderSkill{id: id}
}

// Re-defining All with real implementations where available
var Registry = []runner.Skill{
	&LintGofumpt{},
	&LintGolangCI{},
	NewTestBuild(),
	NewTestBinary(),
	NewTestGo(),
	NewTestCoverage(),
	NewDocsYaml(),
	NewDocsValidateSpec(),
	newPlaceholder("docs:spec-reference-check"),
	NewDocsOrphanSpecs(),
	NewDocsOrphanDocs(),
	NewDocsDocPatterns(),

	newPlaceholder("docs:required-tests"),
	NewDocsHeaderComments(),
	NewPurity(),
	NewDocsPolicy(),
	NewDocsProviderGovernance(),
}
