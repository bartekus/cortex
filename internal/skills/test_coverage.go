package skills

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bartekus/cortex/internal/runner"
)

type TestCoverage struct {
	id string
}

func NewTestCoverage() runner.Skill {
	return &TestCoverage{id: "test:coverage"}
}

func (s *TestCoverage) ID() string { return s.id }

func (s *TestCoverage) Run(ctx context.Context, deps *runner.Deps) runner.SkillResult {
	// 1. Prepare coverage file path
	if deps.StateDir == "" {
		return runner.SkillResult{Skill: s.id, Status: runner.StatusFail, ExitCode: 4, Note: "StateDir not set"}
	}
	coverProfile := filepath.Join(deps.StateDir, "coverage.out")

	// 2. Run go test
	cmd := exec.CommandContext(ctx, "go", "test", "./...", "-coverprofile="+coverProfile, "-covermode=atomic")
	cmd.Dir = deps.RepoRoot

	if out, err := cmd.CombinedOutput(); err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		// Capture output for diagnosis
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: exitCode,
			Note:     strings.TrimSpace(string(out)),
		}
	}

	// 3. Parse Coverage
	//   a) Overall via go tool cover -func
	totalCov, err := getOverallCoverage(ctx, deps.RepoRoot, coverProfile)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Failed to parse coverage: %v", err),
		}
	}

	//   b) Core packages coverage (manual parse)
	corePkgs := []string{"pkg/config", "internal/core"}
	coreCov, err := getCoreCoverage(coverProfile, corePkgs)
	if err != nil {
		return runner.SkillResult{
			Skill:    s.id,
			Status:   runner.StatusFail,
			ExitCode: 4,
			Note:     fmt.Sprintf("Failed to parse core coverage: %v", err),
		}
	}

	// 4. Evaluate Thresholds
	// Overall: < 50 FAIL, < 60 WARN
	// Core: < 80 FAIL

	status := runner.StatusPass
	exitCode := 0
	var notes []string

	notes = append(notes, fmt.Sprintf("Overall: %.1f%%", totalCov))

	if totalCov < 50.0 {
		status = runner.StatusFail
		exitCode = 3
		notes = append(notes, "FAIL: Overall coverage < 50%")
	} else if totalCov < 60.0 {
		msg := "WARNING: Overall coverage < 60%"
		notes = append(notes, msg)
		if deps.FailOnWarning {
			status = runner.StatusFail
			exitCode = 3
			notes = append(notes, "(Fail on warning)")
		}
	}

	notes = append(notes, "Core packages:")
	for pkg, cov := range coreCov {
		// cov is float64. -1 means missing/skipped.
		if cov == -1 {
			notes = append(notes, fmt.Sprintf("  %s: skipped (missing)", pkg))
			continue
		}

		statusStr := "OK"
		if cov < 80.0 {
			status = runner.StatusFail
			exitCode = 3
			statusStr = "FAIL (< 80%)"
		}
		notes = append(notes, fmt.Sprintf("  %s: %.1f%% %s", pkg, cov, statusStr))
	}

	notes = append(notes, fmt.Sprintf("Coverage file: %s", coverProfile))

	return runner.SkillResult{
		Skill:    s.id,
		Status:   status,
		ExitCode: exitCode,
		Note:     strings.Join(notes, "\n"),
	}
}

func getOverallCoverage(ctx context.Context, dir, profile string) (float64, error) {
	cmd := exec.CommandContext(ctx, "go", "tool", "cover", "-func="+profile)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	// Look for line "total:\t\t(statements)\t\tXX.X%"
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				last := parts[len(parts)-1]
				last = strings.TrimSuffix(last, "%")
				return strconv.ParseFloat(last, 64)
			}
		}
	}
	return 0, fmt.Errorf("total coverage line not found")
}

func getCoreCoverage(profile string, packages []string) (map[string]float64, error) {
	// Parse profile manually
	// mode: atomic
	// github.com/foo/bar/file.go:1.1,1.2 1 1

	f, err := os.Open(profile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// Map pkg -> {covered, total}
	stats := make(map[string]struct{ covered, total int64 })

	// Map package suffix to full path?
	// The profile has full paths.
	// We only want to match if it CONTAINS "/pkg/config/" etc.
	// Or typically "module/pkg/config/file.go"

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		fileRange := parts[0]
		numStmts, _ := strconv.ParseInt(parts[1], 10, 64)
		count, _ := strconv.ParseInt(parts[2], 10, 64)

		filePath := strings.Split(fileRange, ":")[0]

		for _, pkg := range packages {
			// Check if file belongs to package
			// Pattern: any/path/<pkg>/file.go
			// Use simple substring check for now, robust enough for "pkg/config" vs "internal/core"
			// Ensure it ends with / or is the directory
			if strings.Contains(filePath, "/"+pkg+"/") || strings.HasPrefix(filePath, pkg+"/") {
				s := stats[pkg]
				s.total += numStmts
				if count > 0 {
					s.covered += numStmts
				}
				stats[pkg] = s
			}
		}
	}

	results := make(map[string]float64)
	for _, pkg := range packages {
		s, ok := stats[pkg]
		if !ok || s.total == 0 {
			results[pkg] = -1 // Missing or no statements
		} else {
			results[pkg] = float64(s.covered) / float64(s.total) * 100.0
		}
	}

	return results, nil
}
