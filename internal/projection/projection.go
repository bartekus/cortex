// SPDX-License-Identifier: AGPL-3.0-or-later
package projection

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AtomicWrite writes content to path atomically by writing to a temp file and renaming it.
func AtomicWrite(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, "projection-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing content: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("moving temp file to %s: %w", path, err)
	}

	return nil
}

// SortedKeys returns the keys of a map[string]int sorted lexicographically.
func SortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// RenderTable renders a Markdown table.
// It assumes rows are already sorted if determinism is required.
func RenderTable(headers []string, rows [][]string) string {
	var b strings.Builder

	// Header
	b.WriteString("| " + strings.Join(headers, " | ") + " |\n")

	// Separator
	b.WriteString("|")
	for range headers {
		b.WriteString(" --- |")
	}
	b.WriteString("\n")

	// Rows
	for _, row := range rows {
		b.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}

	return b.String()
}

// RenderList renders a simple unordered Markdown list.
func RenderList(items []string) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString(fmt.Sprintf("- %s\n", item))
	}
	return b.String()
}

// RenderHeader renders a Markdown header.
func RenderHeader(level int, text string) string {
	return fmt.Sprintf("%s %s\n\n", strings.Repeat("#", level), text)
}
