package scanner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterFiles(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		opts     FilterOptions
		expected []string
	}{
		{
			name:  "exclude node_modules",
			paths: []string{"a.go", "node_modules/bad.js", "pkg/good.go"},
			opts: FilterOptions{
				ExcludeDirs: []string{"node_modules"},
			},
			expected: []string{"a.go", "pkg/good.go"},
		},
		{
			name:  "exclude nested vendor",
			paths: []string{"vendor/a", "pkg/vendor/b", "internal/c"},
			opts: FilterOptions{
				ExcludeDirs: []string{"vendor"},
			},
			expected: []string{"internal/c"},
		},
		{
			name:  "segment matching only",
			paths: []string{"vendor_stuff/a", "myvendor/b"},
			opts: FilterOptions{
				ExcludeDirs: []string{"vendor"},
			},
			expected: []string{"myvendor/b", "vendor_stuff/a"},
		},
		{
			name:  "extension filter",
			paths: []string{"a.go", "b.md", "c.go"},
			opts: FilterOptions{
				IncludeExtensions: []string{".go"},
			},
			expected: []string{"a.go", "c.go"},
		},
		{
			name:  "excludes and extensions",
			paths: []string{"vendor/a.go", "b.go", "c.js"},
			opts: FilterOptions{
				ExcludeDirs:       []string{"vendor"},
				IncludeExtensions: []string{".go"},
			},
			expected: []string{"b.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterFiles(tt.paths, tt.opts)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestScanner(t *testing.T) {
	// Create a temp directory for the git repo
	dir := t.TempDir()
	ctx := context.Background()

	// Initialize git repo
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	// Create some files
	createFile(t, dir, "main.go")
	createFile(t, dir, "vendor/foo.go")
	createFile(t, dir, "node_modules/bar.js")
	createFile(t, dir, ".gitignore", "ignored.txt")
	createFile(t, dir, "ignored.txt")
	createFile(t, dir, "pkg/util.go")

	// Commit them
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Initial commit")

	s := New(dir)

	// Test TrackedFiles (should return all tracked files, not ignored ones)
	tracked, err := s.TrackedFiles(ctx)
	require.NoError(t, err)
	assert.Contains(t, tracked, "main.go")
	assert.Contains(t, tracked, "vendor/foo.go")
	assert.NotContains(t, tracked, "ignored.txt") // respected .gitignore

	// Test TrackedFilesFiltered
	filtered, err := s.TrackedFilesFiltered(ctx, FilterOptions{
		ExcludeDirs: []string{"vendor", "node_modules"},
	})
	require.NoError(t, err)
	assert.Contains(t, filtered, "main.go")
	assert.Contains(t, filtered, "pkg/util.go")
	assert.NotContains(t, filtered, "vendor/foo.go")
	assert.NotContains(t, filtered, "node_modules/bar.js")

	// Test TrackedGoFiles
	goFiles, err := s.TrackedGoFiles(ctx)
	require.NoError(t, err)
	// Default excludes include vendor and node_modules
	assert.Contains(t, goFiles, "main.go")
	assert.Contains(t, goFiles, "pkg/util.go")
	assert.NotContains(t, goFiles, "vendor/foo.go")
	assert.NotContains(t, goFiles, ".gitignore")
}

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, out)
	}
}

func createFile(t *testing.T, dir, path string, content ...string) {
	fullPath := filepath.Join(dir, path)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	require.NoError(t, err)

	data := ""
	if len(content) > 0 {
		data = content[0]
	}
	err = os.WriteFile(fullPath, []byte(data), 0644)
	require.NoError(t, err)
}
