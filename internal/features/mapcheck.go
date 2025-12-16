// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package features contains governance tooling for feature-level analysis.
//
// Feature: GOV_CORE
// Spec: spec/governance/GOV_CORE.md
//
// This package implements Phase 4 of GOV_CORE: multi-feature cross-validation
// and enforcement of the Feature Mapping Invariant across specs, code, and tests.
package features

import (
	"context"
	"fmt"
	"io"
	"os"
)

// Runner coordinates loading specs, scanning source files, validating the
// feature mappings, and reporting any issues in a deterministic way.
type Runner struct {
	// FeaturesPath is the path to spec/features.yaml.
	FeaturesPath string

	// RootDir is the root directory to scan for Go files and specs.
	RootDir string

	// Out is where human-readable reports are written. Defaults to os.Stdout
	// if nil.
	Out io.Writer
}

// Run executes the feature mapping analysis. It returns an error if any
// validation error (not just a warning) is encountered.
func (r *Runner) Run(ctx context.Context) error {
	if r.Out == nil {
		r.Out = os.Stdout
	}

	if r.FeaturesPath == "" {
		r.FeaturesPath = "spec/features.yaml"
	}
	if r.RootDir == "" {
		r.RootDir = "."
	}

	features, err := LoadFeaturesYAML(r.RootDir, r.FeaturesPath)
	if err != nil {
		return fmt.Errorf("loading features.yaml: %w", err)
	}

	index, err := ScanSourceTree(ctx, r.RootDir, features)
	if err != nil {
		return fmt.Errorf("scanning source tree: %w", err)
	}

	issues, err := ValidateFeatureIndex(index)
	if err != nil {
		return fmt.Errorf("validating feature index: %w", err)
	}

	// Deterministic sorting is handled inside ValidateFeatureIndex or here
	// before printing. Issues are already expected to be sorted by
	// FeatureID, Severity, File, Line.
	if len(issues) == 0 {
		_, _ = fmt.Fprintln(r.Out, "feature-map-check: OK")
		return nil
	}

	printIssues(r.Out, issues)

	// Exit as error if any issue is severity Error, otherwise success with warnings.
	if hasErrors(issues) {
		return fmt.Errorf("feature mapping validation failed with %d error(s)", countErrors(issues))
	}

	return nil
}

func printIssues(w io.Writer, issues []ValidationIssue) {
	for _, iss := range issues {
		// Example format:
		// ERROR [CLI_DEPLOY] internal/cli/commands/deploy.go:42: missing Spec header
		_, _ = fmt.Fprintf(
			w,
			"%s [%s] %s:%d: %s\n",
			iss.Severity,
			iss.FeatureID,
			iss.File,
			iss.Line,
			iss.Message,
		)
	}
}

func hasErrors(issues []ValidationIssue) bool {
	for _, iss := range issues {
		if iss.Severity == SeverityError {
			return true
		}
	}
	return false
}

func countErrors(issues []ValidationIssue) int {
	var n int
	for _, iss := range issues {
		if iss.Severity == SeverityError {
			n++
		}
	}
	return n
}
