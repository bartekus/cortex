// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package featuretrace defines the data model for feature traceability reports.
//
// Feature: PROVIDER_FRONTEND_GENERIC
// Docs: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md
package featuretrace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ScanConfig contains configuration for scanning feature presence.
type ScanConfig struct {
	RootDir string
}

// walkDir walks the directory tree from root, calling fn for each file.
// Directories are traversed in lexicographical order to ensure determinism.
func walkDir(root string, fn func(path string, info os.FileInfo) error) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("reading directory %s: %w", root, err)
	}

	// Sort entries lexicographically by name
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(root, name)

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("getting info for %s: %w", fullPath, err)
		}

		if info.IsDir() {
			// Skip hidden dirs, .git, and testdata
			if strings.HasPrefix(name, ".") || name == "testdata" || name == ".git" {
				continue
			}
			if err := walkDir(fullPath, fn); err != nil {
				return err
			}
			continue
		}

		if err := fn(fullPath, info); err != nil {
			return err
		}
	}

	return nil
}

// ScanFeaturePresence scans the repository tree and returns feature presence information.
// The result is sorted lexicographically by FeatureID.
func ScanFeaturePresence(cfg ScanConfig) ([]FeaturePresence, error) {
	presenceMap := make(map[string]*FeaturePresence)

	err := walkDir(cfg.RootDir, func(path string, info os.FileInfo) error {
		featureID, err := extractFeatureIDFromFile(path)
		if err != nil {
			return fmt.Errorf("extracting feature ID from %s: %w", path, err)
		}
		if featureID == "" {
			return nil
		}

		rel, err := filepath.Rel(cfg.RootDir, path)
		if err != nil {
			return fmt.Errorf("making relative path for %s: %w", path, err)
		}
		rel = filepath.ToSlash(rel)

		fp, ok := presenceMap[featureID]
		if !ok {
			fp = &FeaturePresence{
				FeatureID: featureID,
			}
			presenceMap[featureID] = fp
		}

		switch {
		case isSpecFile(rel):
			fp.HasSpec = true
			if fp.SpecPath == "" {
				fp.SpecPath = rel
			}
		case isTestFile(rel):
			fp.TestFiles = append(fp.TestFiles, rel)
		case isImplementationFile(rel):
			fp.ImplementationFiles = append(fp.ImplementationFiles, rel)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning repository: %w", err)
	}

	// Normalize lists and convert to sorted slice
	result := make([]FeaturePresence, 0, len(presenceMap))
	for _, fp := range presenceMap {
		sort.Strings(fp.ImplementationFiles)
		sort.Strings(fp.TestFiles)
		result = append(result, *fp)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].FeatureID < result[j].FeatureID
	})

	return result, nil
}

// extractFeatureIDFromFile extracts Feature ID from a file's header comment.
// Returns empty string if no Feature ID is found.
func extractFeatureIDFromFile(filePath string) (string, error) {
	f, err := os.Open(filePath) //nolint:gosec // G304: file path is from repository scan, not user input
	if err != nil {
		return "", fmt.Errorf("opening file %s: %w", filePath, err)
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Stop when we hit a non-comment line - only header comments count.
		if !strings.HasPrefix(trimmed, "//") {
			break
		}

		if strings.HasPrefix(trimmed, "// Feature:") {
			rest := strings.TrimPrefix(trimmed, "// Feature:")
			id := strings.TrimSpace(rest)
			return id, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanning file %s: %w", filePath, err)
	}

	return "", nil
}

// isSpecFile determines if a file path represents a spec file.
func isSpecFile(path string) bool {
	if filepath.Ext(path) != ".md" {
		return false
	}
	p := filepath.ToSlash(path)
	return strings.HasPrefix(p, "spec/") || strings.Contains(p, "/spec/")
}

// isTestFile determines if a file path represents a test file.
func isTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test.go")
}

// isImplementationFile determines if a file path represents an implementation file.
func isImplementationFile(path string) bool {
	if filepath.Ext(path) != ".go" {
		return false
	}
	if isTestFile(path) {
		return false
	}
	p := filepath.ToSlash(path)
	return !strings.Contains(p, "/testdata/")
}
