// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md
package reports

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bartekus/cortex/internal/reports/commithealth"
)

func TestParseGitLogOutput_SingleCommit(t *testing.T) {
	t.Parallel()

	output := "abc123|feat(CLI_DEPLOY): add rollback support|John Doe|john@example.com\n"

	commits, err := parseGitLogOutput(output)
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}

	commit := commits[0]
	if commit.SHA != "abc123" {
		t.Errorf("expected SHA=abc123, got %s", commit.SHA)
	}
	if commit.Message != "feat(CLI_DEPLOY): add rollback support" {
		t.Errorf("expected message='feat(CLI_DEPLOY): add rollback support', got %s", commit.Message)
	}
	if commit.AuthorName != "John Doe" {
		t.Errorf("expected author=John Doe, got %s", commit.AuthorName)
	}
	if commit.AuthorEmail != "john@example.com" {
		t.Errorf("expected email=john@example.com, got %s", commit.AuthorEmail)
	}
}

func TestParseGitLogOutput_MultipleCommits(t *testing.T) {
	t.Parallel()

	output := "abc123|feat(CLI_DEPLOY): add rollback support|John Doe|john@example.com\n" +
		"def456|fix(CORE_CONFIG): fix config loading|Jane Smith|jane@example.com\n"

	commits, err := parseGitLogOutput(output)
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	// Verify deterministic ordering (should be sorted)
	// TODO: Verify sort order matches spec (likely by SHA lexicographically)
}

func TestParseGitLogOutput_WeirdButValidMessages(t *testing.T) {
	t.Parallel()

	// Test messages with colons, brackets, multiple lines
	output := "abc123|feat(CLI_DEPLOY): add support for :special: chars|John Doe|john@example.com\n" +
		"def456|fix(CORE): handle [brackets] in messages|Jane Smith|jane@example.com\n"

	commits, err := parseGitLogOutput(output)
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
}

func TestParseGitLogOutput_InvalidLines(t *testing.T) {
	t.Parallel()

	// Test truncated or malformed lines
	output := "abc123|incomplete line\n" +
		"def456|feat(CLI_DEPLOY): valid commit|John Doe|john@example.com\n"

	_, err := parseGitLogOutput(output)
	if err == nil {
		t.Error("expected error for invalid line, got nil")
	}
}

func TestParseGitLogOutput_EmptyOutput(t *testing.T) {
	t.Parallel()

	commits, err := parseGitLogOutput("")
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if len(commits) != 0 {
		t.Errorf("expected 0 commits for empty output, got %d", len(commits))
	}
}

func TestNewHistorySource(t *testing.T) {
	t.Parallel()

	repoPath := "/some/repo/path"
	source := NewHistorySource(repoPath)

	impl, ok := source.(*HistorySourceImpl)
	if !ok {
		t.Fatalf("expected *HistorySourceImpl, got %T", source)
	}

	if impl.repoPath != repoPath {
		t.Errorf("expected repoPath=%q, got %q", repoPath, impl.repoPath)
	}
}

func TestHistorySourceImpl_Commits(t *testing.T) {
	t.Parallel()

	// Create a temporary git repository for testing
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(repoPath, 0o750); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to initialize git repo: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to configure git user.name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to configure git user.email: %v", err)
	}

	// Create and commit a file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "feat(CLI_DEPLOY): initial commit", "--author", "Test User <test@example.com>")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git commit: %v", err)
	}

	// Test Commits method
	source := NewHistorySource(repoPath)
	commits, err := source.Commits()
	if err != nil {
		t.Fatalf("Commits() failed: %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}

	commit := commits[0]
	if commit.Message != "feat(CLI_DEPLOY): initial commit" {
		t.Errorf("expected message='feat(CLI_DEPLOY): initial commit', got %q", commit.Message)
	}
	if commit.AuthorName != "Test User" {
		t.Errorf("expected author='Test User', got %q", commit.AuthorName)
	}
	if commit.AuthorEmail != "test@example.com" {
		t.Errorf("expected email='test@example.com', got %q", commit.AuthorEmail)
	}
	if commit.SHA == "" {
		t.Error("expected non-empty SHA")
	}
}

func TestHistorySourceImpl_Commits_EmptyRepo(t *testing.T) {
	t.Parallel()

	// Create a temporary git repository with no commits
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "empty-repo")
	if err := os.MkdirAll(repoPath, 0o750); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to initialize git repo: %v", err)
	}

	source := NewHistorySource(repoPath)
	commits, err := source.Commits()
	// Empty repo (no commits) causes git log to fail with exit status 128
	// This is expected behavior - the function returns an error
	if err == nil {
		t.Error("expected error for empty repo (no commits), got nil")
		return
	}

	// Verify error is properly wrapped
	if !strings.Contains(err.Error(), "git log") {
		t.Errorf("expected error to mention 'git log', got: %v", err)
	}

	// Commits should be nil when there's an error
	if commits != nil {
		t.Errorf("expected nil commits on error, got %v", commits)
	}
}

func TestHistorySourceImpl_Commits_NonExistentRepo(t *testing.T) {
	t.Parallel()

	nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist")
	source := NewHistorySource(nonExistentPath)

	_, err := source.Commits()
	if err == nil {
		t.Error("expected error for non-existent repo, got nil")
	}

	if !strings.Contains(err.Error(), "git log") {
		t.Errorf("expected error to mention 'git log', got: %v", err)
	}
}

func TestRunGitLog(t *testing.T) {
	t.Parallel()

	// Create a temporary git repository
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	if err := os.MkdirAll(repoPath, 0o750); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to initialize git repo: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to configure git user.name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to configure git user.email: %v", err)
	}

	// Create and commit a file
	testFile := filepath.Join(repoPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "test commit")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git commit: %v", err)
	}

	// Test runGitLog
	ctx := context.Background()
	output, err := runGitLog(ctx, repoPath)
	if err != nil {
		t.Fatalf("runGitLog failed: %v", err)
	}

	if output == "" {
		t.Error("expected non-empty output from runGitLog")
	}

	// Verify output format (should contain pipe-separated values)
	if !strings.Contains(output, "|") {
		t.Errorf("expected pipe-separated format, got: %q", output)
	}

	// Verify it contains expected fields
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 1 {
		t.Error("expected at least one line of output")
	}

	parts := strings.Split(lines[0], "|")
	if len(parts) != 4 {
		t.Errorf("expected 4 parts separated by |, got %d: %v", len(parts), parts)
	}
}

func TestRunGitLog_NonExistentRepo(t *testing.T) {
	t.Parallel()

	nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist")
	ctx := context.Background()

	_, err := runGitLog(ctx, nonExistentPath)
	if err == nil {
		t.Error("expected error for non-existent repo, got nil")
	}

	if !strings.Contains(err.Error(), "running git log") {
		t.Errorf("expected error to mention 'running git log', got: %v", err)
	}
}

func TestRunGitLog_InvalidRepo(t *testing.T) {
	t.Parallel()

	// Create a directory that is not a git repository
	tmpDir := t.TempDir()
	ctx := context.Background()

	_, err := runGitLog(ctx, tmpDir)
	if err == nil {
		t.Error("expected error for non-git directory, got nil")
	}

	if !strings.Contains(err.Error(), "running git log") {
		t.Errorf("expected error to mention 'running git log', got: %v", err)
	}
}

func TestFakeHistorySource(t *testing.T) {
	t.Parallel()

	fake := &fakeHistorySource{
		commits: []commithealth.CommitMetadata{
			{
				SHA:         "abc123",
				Message:     "feat(CLI_DEPLOY): add rollback support",
				AuthorName:  "John Doe",
				AuthorEmail: "john@example.com",
			},
		},
	}

	commits, err := fake.Commits()
	if err != nil {
		t.Fatalf("fake.Commits() failed: %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
}
