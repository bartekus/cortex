// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-services applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package mapping implements the Feature Mapping Invariant validator for GOV_CORE Phase 4.
//
// Feature: GOV_CORE
// Spec: spec/governance/GOV_CORE.md
package mapping

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bartekus/cortex/internal/features"
)

// ReportCode is a stable, string-based code for mapping violations.
// These codes are used in both text and JSON output and must be treated
// as part of the public governance contract.
type ReportCode string

const (
	// CodeMissingSpec indicates that a feature is referenced in code or tests
	// but there is no corresponding spec file.
	CodeMissingSpec ReportCode = "MISSING_SPEC"

	// CodeMissingImpl indicates that a feature has a spec entry but no
	// implementation files reference it.
	CodeMissingImpl ReportCode = "MISSING_IMPL"

	// CodeMissingTests indicates that a feature has implementation files but
	// no test files reference it.
	CodeMissingTests ReportCode = "MISSING_TESTS"

	// CodeSpecPathMismatch indicates that the Spec header in code or tests
	// does not match the canonical spec path.
	CodeSpecPathMismatch ReportCode = "SPEC_PATH_MISMATCH"

	// CodeFeatureNotListed indicates that a file references a Feature ID that
	// is not present in spec/features.yaml (dangling Feature ID).
	CodeFeatureNotListed ReportCode = "FEATURE_NOT_LISTED"

	// CodeOrphanSpec indicates that a spec file exists under spec/ but no
	// feature declares it as its canonical spec path.
	CodeOrphanSpec ReportCode = "ORPHAN_SPEC"
)

// FeatureStatus is the per-feature status summary used in reports.
type FeatureStatus string

const (
	// FeatureStatusOK indicates the feature has spec, implementation, and tests.
	FeatureStatusOK FeatureStatus = "ok"
	// FeatureStatusMissingImpl indicates the feature is missing implementation files.
	FeatureStatusMissingImpl FeatureStatus = "missing_impl"
	// FeatureStatusMissingTests indicates the feature is missing test files.
	FeatureStatusMissingTests FeatureStatus = "missing_tests"
	// FeatureStatusIncomplete indicates the feature is partially implemented.
	FeatureStatusIncomplete FeatureStatus = "incomplete"
	// FeatureStatusUnmapped indicates the feature has no spec, impl, or tests.
	FeatureStatusUnmapped FeatureStatus = "unmapped"
	// FeatureStatusSpecOnly indicates the feature has only a spec file.
	FeatureStatusSpecOnly FeatureStatus = "spec_only"
	// FeatureStatusImplementationOnly indicates the feature has implementation but no spec.
	FeatureStatusImplementationOnly FeatureStatus = "implementation_only"
)

// FeatureMapping describes the complete mapping for a single feature.
type FeatureMapping struct {
	// ID is the Feature ID, for example "CLI_INIT".
	ID string `json:"id"`

	// SpecPath is the canonical spec path for this feature, for example
	// "spec/commands/init.md". It is the path derived from spec/features.yaml
	// and the spec file frontmatter, not arbitrary file locations.
	SpecPath string `json:"spec_path"`

	// ImplFiles is the sorted list of Go implementation files that reference
	// this feature via header comments.
	ImplFiles []string `json:"impl_files"`

	// TestFiles is the sorted list of Go test files that reference this
	// feature via header comments.
	TestFiles []string `json:"test_files"`

	// Status is the derived status for this feature, for example "ok" or
	// "missing_tests". It is computed deterministically from the mapping.
	Status FeatureStatus `json:"status"`
}

// Violation describes a single invariant violation discovered during
// feature mapping analysis.
type Violation struct {
	// Code is the machine-readable violation code, for example "MISSING_SPEC".
	Code ReportCode `json:"code"`

	// Feature is the Feature ID associated with the violation, if any. It may
	// be empty when the violation cannot be attributed to a single feature.
	Feature string `json:"feature"`

	// Path is the repository-relative file path where this violation was
	// detected, when applicable.
	Path string `json:"path"`

	// Detail is a deterministic human-readable message describing the
	// violation. Multiple runs on the same repository state MUST produce
	// identical detail strings.
	Detail string `json:"detail"`
}

// Report is the complete, deterministic output of the feature mapping
// analysis. It is the input to both the JSON and text renderers used by
// the CLI.
type Report struct {
	// Features is the sorted list of feature mappings. The slice MUST be
	// sorted lexicographically by Feature ID.
	Features []FeatureMapping `json:"features"`

	// Violations is the sorted list of violations. The slice MUST be sorted
	// lexicographically by (Code, Feature, Path).
	Violations []Violation `json:"violations"`
}

// Options controls how the feature mapping analysis runs.
//
// v1 keeps this intentionally minimal. Additional options (filters, severity
// toggles, etc.) should be added in a backwards compatible way.
type Options struct {
	// RootDir is the repository root to analyse. For normal CLI usage this
	// will be ".", but tests may override it to point at a fixture tree.
	RootDir string
}

// DefaultOptions returns a deterministic set of default options.
func DefaultOptions() Options {
	return Options{
		RootDir: ".",
	}
}

// Analyze runs the feature mapping analysis over the repository located at
// opts.RootDir and returns a deterministic Report.
//
// Behavioural guarantees for v1:
//
//   - No network access.
//   - No environment-variable-dependent behaviour.
//   - No timestamps or random values.
//   - All slices in the returned Report are lexicographically sorted.
func Analyze(opts Options) (Report, error) {
	root := filepath.Clean(opts.RootDir)

	// Load features from spec/features.yaml
	featuresMap, err := features.LoadFeaturesYAML(root, "spec/features.yaml")
	if err != nil {
		return Report{}, fmt.Errorf("loading features: %w", err)
	}

	// Convert to internal featureMeta format
	metas := make([]featureMeta, 0, len(featuresMap))
	for id, fs := range featuresMap {
		// Convert absolute spec path back to relative for consistency
		specPath := fs.Spec
		if filepath.IsAbs(specPath) {
			rel, err := filepath.Rel(root, specPath)
			if err == nil {
				specPath = rel
			}
		}
		metas = append(metas, featureMeta{
			ID:       id,
			Status:   string(fs.Status),
			SpecPath: specPath,
		})
	}

	return analyzeWithFeatures(root, metas)
}

// featureMeta is an internal helper describing what we know about a feature
// from spec/features.yaml. In production this is populated from the
// existing feature parsing tooling. Tests construct it directly.
type featureMeta struct {
	ID       string
	Status   string // "todo", "wip", "done"
	SpecPath string
}

// analyzeWithFeatures contains the actual deterministic analysis logic.
// Tests call this directly with an in-memory feature set.
func analyzeWithFeatures(root string, metas []featureMeta) (Report, error) {
	root = filepath.Clean(root)

	// Canonical feature metadata map.
	metaByID := make(map[string]featureMeta, len(metas))
	specOwner := make(map[string]string, len(metas)) // specPath -> featureID
	for _, m := range metas {
		metaByID[m.ID] = m
		if m.SpecPath != "" {
			normalized := pathClean(m.SpecPath)
			if owner, exists := specOwner[normalized]; exists && owner != m.ID {
				// Duplicate mapping of spec path to features could be handled here
				// with a dedicated violation type if desired.
			} else {
				specOwner[normalized] = m.ID
			}
		}
	}

	// Mapping state built from scanning the tree.
	mappingByID := make(map[string]*FeatureMapping, len(metas))
	violations := make([]Violation, 0)

	// Track all spec files we observe, for orphan detection.
	specFiles := make(map[string]bool)

	// Walk the repository tree deterministically.
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Skip .git, vendor, and testdata directories.
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "testdata" || name == ".stagecraft" {
				return filepath.SkipDir
			}
			return nil
		}

		// Track spec files.
		if strings.HasPrefix(rel, "spec"+string(filepath.Separator)) && strings.HasSuffix(rel, ".md") {
			specFiles[pathClean(rel)] = true
			return nil
		}

		// Only care about Go files for mapping.
		if !strings.HasSuffix(rel, ".go") {
			return nil
		}

		isTestFile := strings.HasSuffix(rel, "_test.go")

		fileFeature, fileSpec, err := parseHeaders(path)
		if err != nil {
			return fmt.Errorf("parse headers in %s: %w", rel, err)
		}

		if fileFeature == "" {
			// No feature header, nothing to map.
			return nil
		}

		// Normalise relative path for reporting.
		rel = pathClean(rel)

		meta, known := metaByID[fileFeature]

		if !known {
			// Feature used in code but not listed - dangling ID.
			violations = append(violations, Violation{
				Code:    CodeFeatureNotListed,
				Feature: fileFeature,
				Path:    rel,
				Detail:  fmt.Sprintf("feature %s referenced in %s but not listed in feature set", fileFeature, rel),
			})
			return nil
		}

		fm := mappingByID[fileFeature]
		if fm == nil {
			fm = &FeatureMapping{
				ID:       fileFeature,
				SpecPath: pathClean(meta.SpecPath),
			}
			mappingByID[fileFeature] = fm
		}

		if isTestFile {
			fm.TestFiles = append(fm.TestFiles, rel)
		} else {
			fm.ImplFiles = append(fm.ImplFiles, rel)
		}

		// Check Spec header correctness when we know the canonical path.
		canonical := pathClean(meta.SpecPath)
		if canonical != "" && fileSpec != "" && canonical != pathClean(fileSpec) {
			violations = append(violations, Violation{
				Code:    CodeSpecPathMismatch,
				Feature: fileFeature,
				Path:    rel,
				Detail:  fmt.Sprintf("file declares Spec %q but canonical spec path is %q", fileSpec, canonical),
			})
		}

		return nil
	})
	if err != nil {
		return Report{}, err
	}

	// Build FeatureMapping entries for all known features, even if they had no files.
	for _, m := range metas {
		fm := mappingByID[m.ID]
		if fm == nil {
			fm = &FeatureMapping{
				ID:       m.ID,
				SpecPath: pathClean(m.SpecPath),
			}
			mappingByID[m.ID] = fm
		}

		// Determine whether the spec file exists.
		specExists := false
		if m.SpecPath != "" {
			specExists = specFiles[pathClean(m.SpecPath)]
		}

		// Sort file lists before deriving status or violations.
		sort.Strings(fm.ImplFiles)
		sort.Strings(fm.TestFiles)

		switch m.Status {
		case "todo":
			switch {
			case !specExists && len(fm.ImplFiles) == 0 && len(fm.TestFiles) == 0:
				fm.Status = FeatureStatusUnmapped
			case specExists && len(fm.ImplFiles) == 0 && len(fm.TestFiles) == 0:
				fm.Status = FeatureStatusSpecOnly
			case !specExists && (len(fm.ImplFiles) > 0 || len(fm.TestFiles) > 0):
				fm.Status = FeatureStatusImplementationOnly
			default:
				fm.Status = FeatureStatusIncomplete
			}

		case "wip":
			if !specExists {
				violations = append(violations, Violation{
					Code:    CodeMissingSpec,
					Feature: m.ID,
					Path:    pathClean(m.SpecPath),
					Detail:  fmt.Sprintf("wip feature %s has no spec file at %s", m.ID, m.SpecPath),
				})
			}
			if len(fm.ImplFiles) == 0 && len(fm.TestFiles) == 0 {
				violations = append(violations, Violation{
					Code:    CodeMissingImpl,
					Feature: m.ID,
					Path:    "",
					Detail:  fmt.Sprintf("wip feature %s has no implementation or tests", m.ID),
				})
				fm.Status = FeatureStatusMissingImpl
			} else if specExists {
				fm.Status = FeatureStatusOK
			} else {
				fm.Status = FeatureStatusIncomplete
			}

		case "done":
			if !specExists {
				violations = append(violations, Violation{
					Code:    CodeMissingSpec,
					Feature: m.ID,
					Path:    pathClean(m.SpecPath),
					Detail:  fmt.Sprintf("done feature %s has no spec file at %s", m.ID, m.SpecPath),
				})
			}
			if len(fm.ImplFiles) == 0 && len(fm.TestFiles) == 0 {
				violations = append(violations, Violation{
					Code:    CodeMissingImpl,
					Feature: m.ID,
					Path:    "",
					Detail:  fmt.Sprintf("done feature %s has no implementation or tests", m.ID),
				})
				fm.Status = FeatureStatusMissingImpl
			} else if len(fm.TestFiles) == 0 {
				violations = append(violations, Violation{
					Code:    CodeMissingTests,
					Feature: m.ID,
					Path:    "",
					Detail:  fmt.Sprintf("done feature %s has no tests", m.ID),
				})
				fm.Status = FeatureStatusMissingTests
			} else if specExists {
				fm.Status = FeatureStatusOK
			} else {
				fm.Status = FeatureStatusIncomplete
			}

		default:
			// Unknown status - treat similar to todo.
			if !specExists && len(fm.ImplFiles) == 0 && len(fm.TestFiles) == 0 {
				fm.Status = FeatureStatusUnmapped
			} else {
				fm.Status = FeatureStatusIncomplete
			}
		}
	}

	// Detect orphan specs: any spec file that is not the canonical spec of any feature.
	// Exclude ADR files (Architecture Decision Records) as they are not feature specs.
	for specPath := range specFiles {
		// Skip ADR files - they are architectural decisions, not feature specs.
		if strings.HasPrefix(specPath, "spec/adr/") {
			continue
		}
		if _, ok := specOwner[specPath]; !ok {
			violations = append(violations, Violation{
				Code:    CodeOrphanSpec,
				Feature: "",
				Path:    specPath,
				Detail:  fmt.Sprintf("spec file %s is not referenced by any feature", specPath),
			})
		}
	}

	// Build deterministic slices.
	featuresList := make([]FeatureMapping, 0, len(mappingByID))
	for _, fm := range mappingByID {
		featuresList = append(featuresList, *fm)
	}

	sort.Slice(featuresList, func(i, j int) bool {
		return featuresList[i].ID < featuresList[j].ID
	})

	sort.Slice(violations, func(i, j int) bool {
		return lessViolation(violations[i], violations[j])
	})

	return Report{
		Features:   featuresList,
		Violations: violations,
	}, nil
}

// pathClean normalises a path to use forward slashes for portability and
// removes leading "./" segments.
func pathClean(p string) string {
	if p == "" {
		return ""
	}
	p = filepath.ToSlash(filepath.Clean(p))
	if strings.HasPrefix(p, "./") {
		return p[2:]
	}
	return p
}

// parseHeaders reads a Go file and extracts the first Feature and Spec headers
// from line comments of the form:
//
//	// Feature: FEATURE_ID
//	// Spec: spec/path/to/file.md
//
// The match is case-insensitive on the keys.
func parseHeaders(path string) (featureID, specPath string, err error) {
	f, err := os.Open(path) //nolint:gosec // path is from filepath.WalkDir, safe
	if err != nil {
		return "", "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// Log close error but don't fail parsing if file was already read
			_ = closeErr
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "//") {
			continue
		}

		body := strings.TrimSpace(strings.TrimPrefix(line, "//"))
		lower := strings.ToLower(body)
		switch {
		case strings.HasPrefix(lower, "feature:"):
			if featureID == "" {
				featureID = strings.TrimSpace(body[len("Feature:"):])
			}
		case strings.HasPrefix(lower, "spec:"):
			if specPath == "" {
				specPath = strings.TrimSpace(body[len("Spec:"):])
			}
		}

		if featureID != "" && specPath != "" {
			// We have both headers, no need to keep scanning.
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	return featureID, specPath, nil
}

// lessViolation defines the deterministic ordering for violations.
func lessViolation(a, b Violation) bool {
	if a.Code != b.Code {
		return a.Code < b.Code
	}
	if a.Feature != b.Feature {
		return a.Feature < b.Feature
	}
	return a.Path < b.Path
}
