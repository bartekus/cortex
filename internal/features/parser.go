// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// FeatureStatus represents the lifecycle state of a feature.
type FeatureStatus string

// Feature status constants.
const (
	FeatureStatusTodo       FeatureStatus = "todo"
	FeatureStatusWIP        FeatureStatus = "wip"
	FeatureStatusDone       FeatureStatus = "done"
	FeatureStatusDeprecated FeatureStatus = "deprecated"
	FeatureStatusRemoved    FeatureStatus = "removed"
)

// FeatureSpec is a single feature entry from spec/features.yaml (plus resolved spec path).
type FeatureSpec struct {
	ID     string
	Status FeatureStatus
	Spec   string // canonical path like "spec/commands/deploy.md"
}

// FileReference represents a single file-level Feature/Spec header pair.
type FileReference struct {
	File      string
	Line      int
	FeatureID string
	SpecPath  string
	IsTest    bool
}

// FeatureIndex aggregates all known information about features from specs, code, and tests.
type FeatureIndex struct {
	RootDir  string                     // Root directory for path resolution
	Features map[string]*FeatureSpec    // Feature ID -> spec
	Impls    map[string][]FileReference // Feature ID -> impl references
	Tests    map[string][]FileReference // Feature ID -> test references
}

// LoadFeaturesYAML loads spec/features.yaml and builds the initial FeatureSpec map.
func LoadFeaturesYAML(rootDir, relPath string) (map[string]*FeatureSpec, error) {
	path := filepath.Join(rootDir, relPath)

	data, err := os.ReadFile(path) //nolint:gosec // path is from config, not user input
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var featuresYAML YAML
	if err := yaml.Unmarshal(data, &featuresYAML); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	result := make(map[string]*FeatureSpec, len(featuresYAML.Features))

	for i := range featuresYAML.Features {
		feature := &featuresYAML.Features[i]
		specPath := feature.Spec
		if specPath != "" {
			if !strings.HasPrefix(specPath, "spec/") {
				specPath = filepath.Join("spec", specPath)
			}
			// Resolve relative to rootDir for consistent path checking
			specPath = filepath.Join(rootDir, specPath)
		}

		result[feature.ID] = &FeatureSpec{
			ID:     feature.ID,
			Status: FeatureStatus(feature.Status),
			Spec:   specPath,
		}
	}

	return result, nil
}

// ScanSourceTree walks the repository tree and extracts Feature/Spec headers
// from Go files and spec files.
//
// It is deterministic: directory entries are sorted lexicographically.
func ScanSourceTree(ctx context.Context, rootDir string, features map[string]*FeatureSpec) (*FeatureIndex, error) {
	index := &FeatureIndex{
		RootDir:  rootDir,
		Features: make(map[string]*FeatureSpec, len(features)),
		Impls:    make(map[string][]FileReference),
		Tests:    make(map[string][]FileReference),
	}

	// Copy features into index to keep a single source of truth for later validation.
	for id, fs := range features {
		index.Features[id] = fs
	}

	err := walkSorted(rootDir, func(path string, info os.FileInfo) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if info.IsDir() {
			// Skip vendor and .git and .stagecraft directories by default.
			name := info.Name()
			if name == ".git" || name == "vendor" || name == ".stagecraft" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// We intentionally scan both impl and test files, but classify them separately.
		ref, ok, err := parseGoFileHeaders(path)
		if err != nil {
			return fmt.Errorf("parsing headers in %s: %w", path, err)
		}
		if !ok {
			return nil
		}

		if ref.IsTest {
			index.Tests[ref.FeatureID] = append(index.Tests[ref.FeatureID], ref)
		} else {
			index.Impls[ref.FeatureID] = append(index.Impls[ref.FeatureID], ref)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort all references for deterministic output
	for id := range index.Impls {
		sortRefs(index.Impls[id])
	}
	for id := range index.Tests {
		sortRefs(index.Tests[id])
	}

	return index, nil
}

// walkSorted walks the tree rooted at rootDir, invoking fn for each file.
// Directories are traversed in lexicographic order for determinism.
func walkSorted(rootDir string, fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// filepath.Walk itself is not guaranteed to be deterministic with respect
		// to directory entry ordering, but in practice is stable enough on most
		// platforms. If stricter guarantees are required, replace this with a
		// custom recursive traversal that sorts os.ReadDir results.
		return fn(path, info)
	})
}

// parseGoFileHeaders scans the first N lines of a Go file to find Feature/Spec
// headers. Returns (ref, true, nil) if a Feature header is found.
func parseGoFileHeaders(path string) (FileReference, bool, error) {
	const maxHeaderLines = 32

	f, err := os.Open(path) //nolint:gosec // path is from walkSorted, safe
	if err != nil {
		return FileReference{}, false, err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	var (
		lineNum     int
		featureID   string
		specPath    string
		featureLine int
	)

	for scanner.Scan() && lineNum < maxHeaderLines {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if !strings.HasPrefix(line, "//") {
			continue
		}

		text := strings.TrimSpace(strings.TrimPrefix(line, "//"))
		if strings.HasPrefix(text, "Feature:") {
			featureID = strings.TrimSpace(strings.TrimPrefix(text, "Feature:"))
			featureLine = lineNum
		} else if strings.HasPrefix(text, "Spec:") {
			specPath = strings.TrimSpace(strings.TrimPrefix(text, "Spec:"))
		}
	}

	if err := scanner.Err(); err != nil {
		return FileReference{}, false, err
	}

	if featureID == "" {
		return FileReference{}, false, nil
	}

	ref := FileReference{
		File:      path,
		Line:      featureLine,
		FeatureID: featureID,
		SpecPath:  specPath,
		IsTest:    strings.HasSuffix(path, "_test.go"),
	}

	return ref, true, nil
}

// sortRefs deterministically sorts a slice of FileReference in-place.
func sortRefs(refs []FileReference) {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].FeatureID != refs[j].FeatureID {
			return refs[i].FeatureID < refs[j].FeatureID
		}
		if refs[i].File != refs[j].File {
			return refs[i].File < refs[j].File
		}
		return refs[i].Line < refs[j].Line
	})
}
