// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bartekus/cortex/internal/featureindex"
	"github.com/bartekus/cortex/internal/projectmeta"
	"github.com/bartekus/cortex/internal/projectroot"
	"github.com/bartekus/cortex/internal/git"
	"github.com/bartekus/cortex/internal/reports"
	"github.com/bartekus/cortex/internal/reports/commithealth"
)

// Feature: PROVIDER_FRONTEND_GENERIC
// Docs: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md

var newHistorySource = git.NewHistorySource

// NewCommitReportCommand returns the `stagecraft commit report` command.
func NewCommitReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit report",
		Short: "Generate commit health report",
		Long:  "Generates a commit health report analyzing commit message discipline",
		RunE:  runCommitReport,
	}

	// Flags in alphabetical order for deterministic help output
	cmd.Flags().String("from", "origin/main", "Start of commit range (default: origin/main)")
	cmd.Flags().String("to", "HEAD", "End of commit range (default: HEAD)")

	return cmd
}

// runCommitReport executes the commit report command.
func runCommitReport(cmd *cobra.Command, args []string) error {
	// 1. Get repository root
	repoPath, err := projectroot.Find(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	// 2. Get commit range flags
	fromFlag, _ := cmd.Flags().GetString("from")
	toFlag, _ := cmd.Flags().GetString("to")

	// 3. Load feature registry
	knownFeatures, err := featureindex.Load(repoPath)
	if err != nil {
		return fmt.Errorf("loading feature registry: %w", err)
	}

	// 4. Get commit history via git adapter
	historySource := newHistorySource(repoPath)
	commits, err := historySource.Commits()
	if err != nil {
		return fmt.Errorf("retrieving commit history: %w", err)
	}

	// 5. Determine repo info
	repoInfo := commithealth.RepoInfo{
		Name:          projectmeta.DetermineRepoName(repoPath),
		DefaultBranch: "main", // TODO: Detect from git config
	}

	// 6. Build commit range info
	rangeInfo := commithealth.CommitRange{
		From:        fromFlag,
		To:          toFlag,
		Description: fmt.Sprintf("%s..%s", fromFlag, toFlag),
	}

	// 7. Generate report using Phase 3.B generator
	report, err := commithealth.GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		return fmt.Errorf("generating commit health report: %w", err)
	}

	// 8. Write report atomically
	reportPath := filepath.Join(repoPath, ".stagecraft", "reports", "commit-health.json")
	if err := reports.WriteJSONAtomic(reportPath, report); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	return nil
}

// loadFeatureRegistry loads `spec/features.yaml` from the given repo root and returns a set of feature IDs.
//
// This is intentionally lightweight and dependency-free (no YAML parser) because the only field we need
// for CLI validation and reporting is the `id` value.
//
// Supported patterns (after TrimSpace):
// - "- id: FOO"
// - "id: FOO"
func loadFeatureRegistry(root string) (map[string]bool, error) {
	path := filepath.Join(root, "spec", "features.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	out := make(map[string]bool)
	lines := strings.Split(string(b), "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)

		// Match either "- id:" or "id:".
		var v string
		switch {
		case strings.HasPrefix(line, "- id:"):
			v = strings.TrimSpace(strings.TrimPrefix(line, "- id:"))
		case strings.HasPrefix(line, "id:"):
			v = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		default:
			continue
		}

		// Strip optional quotes
		v = strings.Trim(v, "\"'")
		if v == "" {
			continue
		}
		out[v] = true
	}

	return out, nil
}
