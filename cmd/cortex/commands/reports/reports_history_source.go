// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package git provides a deterministic git adapter for commit history retrieval.
//
// Feature: XRAY_INDEX_FORMAT
// Spec: spec/xray/index-format.md
package reports

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/reports/commithealth"
)

// HistorySourceImpl implements HistorySource using git commands.
type HistorySourceImpl struct {
	repoPath string
}

// NewHistorySource creates a new HistorySource that reads from the given repository path.
func NewHistorySource(repoPath string) commithealth.HistorySource {
	return &HistorySourceImpl{
		repoPath: repoPath,
	}
}

// Commits retrieves commit history from git, sorted deterministically.
func (h *HistorySourceImpl) Commits() ([]commithealth.CommitMetadata, error) {
	ctx := context.Background()
	output, err := runGitLog(ctx, h.repoPath)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	commits, err := parseGitLogOutput(output)
	if err != nil {
		return nil, fmt.Errorf("parsing git log output: %w", err)
	}

	return commits, nil
}

// runGitLog executes git log and returns the raw output.
// This function shells out to git with explicit environment variables.
func runGitLog(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", `--format=%H|%s|%an|%ae`, "--reverse")
	cmd.Dir = repoPath
	// Explicit, minimal environment - no implicit inheritance.
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"LANG=C",
		"LC_ALL=C",
	}

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running git log: %w", err)
	}

	return string(out), nil
}

// parseGitLogOutput parses git log output into CommitMetadata slices.
// This is a pure function that can be tested without shelling out to git.
func parseGitLogOutput(output string) ([]commithealth.CommitMetadata, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil, nil
	}

	lines := strings.Split(trimmed, "\n")
	commits := make([]commithealth.CommitMetadata, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			return nil, fmt.Errorf("malformed git log line: %q", line)
		}

		commits = append(commits, commithealth.CommitMetadata{
			SHA:         parts[0],
			Message:     parts[1],
			AuthorName:  parts[2],
			AuthorEmail: parts[3],
		})
	}

	// Deterministic ordering regardless of git log ordering
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].SHA < commits[j].SHA
	})

	return commits, nil
}
