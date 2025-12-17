package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
	"github.com/bartekus/cortex/internal/scanner"
	"gopkg.in/yaml.v3"
)

type DocsYaml struct {
	id string
}

func NewDocsYaml() runner.Skill {
	return &DocsYaml{id: "docs:yaml"}
}

func (s *DocsYaml) ID() string { return s.id }

func (s *DocsYaml) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// Repo-sensitivity: Skip if spec/features.yaml is absent used as a proxy for "is this a spec repo?"
	// Or just if "spec/" absent?
	// User said "skip cleanly when spec/features.yaml is absent".

	featuresPath := filepath.Join(deps.RepoRoot, "spec", "features.yaml")
	if info, err := os.Stat(featuresPath); err != nil || info.IsDir() {
		return runner.SkillResult{
			Skill:  s.id,
			Status: runner.StatusSkip,
			Note:   "spec/features.yaml not found",
		}
	}

	// Find all YAML files in spec/ or generally?
	// "docs:yaml" in Cortex implies validating all docs/spec yamls.
	// We will restrict to tracked files in "spec/" for now, to be safe and relevant.
	// Or maybe "docs/" too if it has frontmatter?
	// Let's stick to "spec/" for this specific requirement "registry files".
	// Using scanner to be efficient.

	opts := scanner.FilterOptions{
		IncludeExtensions: []string{".yaml", ".yml"},
		// We only want files inside "spec/" for now?
		// Or maybe everything?
		// "Implement docs:yaml as a thin 'fast parse' check... catches invalid YAML or missing registry files"
		// This implies it checks the registry files specifically.
		// Let's check ALL tracked yaml files in the repo to be helpful?
		// No, might be too broad. Let's start with `spec/`.
	}

	// We need to filter scanner results by directory manually or add Dir option to scanner?
	// Scanner returns all tracked files.
	// We'll filter in loop.

	files, err := deps.Scanner.TrackedFilesFiltered(ctx, opts)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Scanning failed: %v", err),
		}
	}

	var failedFiles []string
	var checkedCount int

	for _, path := range files {
		// Only check spec/ directory?
		// And maybe root features.yaml?
		// path is relative to repo root (TrackedFilesFiltered returns relative paths).
		if !strings.HasPrefix(path, "spec/") && path != "spec/features.yaml" { // redundant check but clear
			continue
		}

		checkedCount++
		fullPath := filepath.Join(deps.RepoRoot, path)

		f, err := os.Open(fullPath)
		if err != nil {
			failedFiles = append(failedFiles, fmt.Sprintf("%s: open error: %v", path, err))
			continue
		}

		var node yaml.Node
		dec := yaml.NewDecoder(f)
		if err := dec.Decode(&node); err != nil {
			failedFiles = append(failedFiles, fmt.Sprintf("%s: invalid YAML: %v", path, err))
		}
		f.Close()
	}

	if len(failedFiles) > 0 {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 1,
			Note:     strings.Join(failedFiles, "\n"),
		}
	}

	return runner.SkillResult{
		Skill:    s.id,
		Status:   runner.StatusPass,
		ExitCode: 0,
		Note:     fmt.Sprintf("Validated %d YAML files in spec/", checkedCount),
	}
}
