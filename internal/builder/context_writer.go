// SPDX-License-Identifier: AGPL-3.0-or-later
package builder

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/bartekus/cortex/internal/xray"
)

// Meta represents .cortex/meta.json
type Meta struct {
	ProjectName string `json:"project_name"`
	Generator   string `json:"generator"`
}

// ManifestEntry represents an item in .cortex/files/manifest.json
type ManifestEntry struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// BuildContext generates the deterministic .cortex/ structure.
func BuildContext(repoRoot string, index *xray.Index) error {
	ctxDir := filepath.Join(repoRoot, ".cortex")
	if err := os.MkdirAll(filepath.Join(ctxDir, "files"), 0o755); err != nil {
		return fmt.Errorf("creating context structure: %w", err)
	}

	// 1. Generate meta.json
	meta := Meta{
		ProjectName: filepath.Base(repoRoot),
		Generator:   "cortex-v0.1.0",
	}
	metaBytes, err := writeJSON(filepath.Join(ctxDir, "meta.json"), meta)
	if err != nil {
		return err
	}

	// 2. Generate files/manifest.json
	// Sort files by path (XRAY index should be sorted, but enforce it)
	// We preserve the ordering from XRAY if it's already sorted, but explicit sort ensures safety.
	manifest := make([]ManifestEntry, 0, len(index.Files))
	for _, f := range index.Files {
		manifest = append(manifest, ManifestEntry{
			Path: f.Path,
			Hash: f.Hash,
		})
	}
	sort.Slice(manifest, func(i, j int) bool {
		return manifest[i].Path < manifest[j].Path
	})

	manifestBytes, err := writeJSON(filepath.Join(ctxDir, "files", "manifest.json"), manifest)
	if err != nil {
		return err
	}

	// 3. Generate files/chunks.ndjson
	// Loop manifest, read file, chunk.
	// Contract: Max 200 lines. UTF-8 only.
	// Ordering: Manifest order (Sorted), then StartLine.

	var chunksBuffer []byte
	for _, item := range manifest {
		fullPath := filepath.Join(repoRoot, item.Path)

		// Read file
		content, err := os.ReadFile(fullPath)
		if err != nil {
			// If file missing (race condition vs XRAY?), skip or error?
			// XRAY index says it exists. Error is safer.
			return fmt.Errorf("reading source file %s: %w", item.Path, err)
		}

		// Skip binary / invalid UTF-8 (simplistic check: standard library utf8.Valid)
		// Contract says: "Text-only: define 'binary' as 'invalid UTF-8'"
		// Actually utf8.Valid(content) is better.
		if !isText(content) {
			continue
		}

		// Skip huge files? User said: "pick a cap... reuse 2MB from LOC".
		// Let's explicitly skip > 2MB for now to be safe.
		const MaxFileSize = 2 * 1024 * 1024
		if len(content) > MaxFileSize {
			continue
		}

		fileChunks := chunkContent(item.Path, string(content))
		for _, c := range fileChunks {
			// Marshal individually for NDJSON
			line, err := json.Marshal(c)
			if err != nil {
				return fmt.Errorf("marshaling chunk: %w", err)
			}
			chunksBuffer = append(chunksBuffer, line...)
			chunksBuffer = append(chunksBuffer, '\n')
		}
	}

	chunksPath := filepath.Join(ctxDir, "files", "chunks.ndjson")
	if err := writeFileAtomic(chunksPath, chunksBuffer, 0o644); err != nil {
		return fmt.Errorf("writing chunks.ndjson: %w", err)
	}

	// 4. Generate digest.txt
	// Digest is SHA-256 over the exact bytes written for: manifest.json then meta.json then chunks.ndjson
	hasher := sha256.New()
	_, _ = hasher.Write(manifestBytes)
	_, _ = hasher.Write(metaBytes)
	_, _ = hasher.Write(chunksBuffer)
	digest := hex.EncodeToString(hasher.Sum(nil))

	if err := writeFileAtomic(filepath.Join(ctxDir, "digest.txt"), []byte(digest+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing digest.txt: %w", err)
	}

	return nil
}

// Chunk represents a segment of code in chunks.ndjson
type Chunk struct {
	FilePath  string `json:"file_path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Content   string `json:"content"`
}

func isText(b []byte) bool {
	// Basic null byte check + UTF8 validity
	// XRAY does this too.
	if bytes.IndexByte(b, 0) >= 0 {
		return false
	}
	// Also use utf8.ValidString if we want to be strict
	return utf8.Valid(b)
}

func chunkContent(path, content string) []Chunk {
	// Normalize newlines to \n
	content = strings.ReplaceAll(content, "\r\n", "\n")

	lines := strings.Split(content, "\n")
	// If last line is empty (due to trailing newline), keep it?
	// Usually editors treat trailing newline as end of line.
	// XRAY might count LOC differently.
	// For context, we want textual representation.
	// If split gives empty string at end, it means content file ended with \n.

	// Chunk by 200
	const ChunkSize = 200
	var chunks []Chunk

	totalLines := len(lines)
	for i := 0; i < totalLines; i += ChunkSize {
		end := i + ChunkSize
		if end > totalLines {
			end = totalLines
		}

		// Join lines back
		// We need to preserve newlines between lines.
		// Slice is lines[i:end]
		chunkLines := lines[i:end]
		chunkStr := strings.Join(chunkLines, "\n")

		chunks = append(chunks, Chunk{
			FilePath:  path,
			StartLine: i + 1,
			EndLine:   i + len(chunkLines), // Actual end line
			Content:   chunkStr,
		})
	}
	return chunks
}

// writeJSON marshals and writes a file with consistent indentation.
// Returns the exact bytes written (including trailing newline) so callers can hash persisted output.
func writeJSON(path string, v interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling json: %w", err)
	}
	// Single trailing newline for stable POSIX output.
	if !strings.HasSuffix(string(data), "\n") {
		data = append(data, '\n')
	}

	if err := writeFileAtomic(path, data, 0o644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", path, err)
	}
	return data, nil
}

// writeFileAtomic writes content to a temp file in the same directory, fsyncs it, then renames into place.
// This prevents partial writes and keeps outputs deterministic under interruption.
func writeFileAtomic(path string, content []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	// Ensure cleanup on failure.
	success := false
	defer func() {
		if !success {
			// Just close if not closed?
			// Standard pattern: Close() is idempotent if we ignore error on second close? NO.
			// User requested: "no deferred close, do explicit close once".
			// So defer only Remove.
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	if err := os.Chmod(path, mode); err != nil {
		return err
	}

	success = true
	return nil
}
