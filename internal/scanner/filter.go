package scanner

import (
	"sort"
	"strings"
)

// FilterOptions defines criteria for including or excluding files.
type FilterOptions struct {
	// ExcludeDirs is a list of directory names to exclude.
	// Matching is segment-aware: "vendor" excludes "vendor/foo" and "pkg/vendor/bar",
	// but not "vendor_stuff/foo".
	ExcludeDirs []string

	// IncludeExtensions is a list of extensions to include (e.g., ".go").
	// If empty, all extensions are included.
	IncludeExtensions []string
}

// DefaultExcludeDirs returns the standard list of directories to exclude in Cortex.
func DefaultExcludeDirs() []string {
	return []string{
		"node_modules",
		".git",
		"dist",
		"build",
		"out",
		"vendor",
		"target",
		".idea",
		".bin",
		"tools",
		".cortex",
	}
}

// FilterFiles applies the filter options to a list of file paths.
// It returns a new slice of strings, sorted deterministically.
func FilterFiles(paths []string, opts FilterOptions) []string {
	if len(paths) == 0 {
		return nil
	}

	// Pre-process exclusions for faster lookup?
	// For small lists, iteration is fine.
	// We want strict segment matching.

	var filtered []string
	for _, path := range paths {
		if shouldExclude(path, opts.ExcludeDirs) {
			continue
		}
		if !shouldIncludeExtension(path, opts.IncludeExtensions) {
			continue
		}
		filtered = append(filtered, path)
	}

	sort.Strings(filtered)
	return filtered
}

// shouldExclude returns true if the path contains any of the excluded segments.
func shouldExclude(path string, excludes []string) bool {
	if len(excludes) == 0 {
		return false
	}
	parts := strings.Split(path, "/")
	for _, part := range parts {
		for _, exclude := range excludes {
			if part == exclude {
				return true
			}
		}
	}
	return false
}

// shouldIncludeExtension returns true if length is 0 OR path matches one extension.
func shouldIncludeExtension(path string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
