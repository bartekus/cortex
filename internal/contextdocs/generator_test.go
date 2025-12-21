// SPDX-License-Identifier: AGPL-3.0-or-later
package contextdocs_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bartekus/cortex/internal/contextdocs"
	"github.com/bartekus/cortex/internal/xray"
)

func TestGenerator_Generate_Golden(t *testing.T) {
	// 1. Load Fixture
	fixturePath := filepath.Join("testdata", "fixture_index.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	var index xray.Index
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatalf("unmarshaling fixture: %v", err)
	}

	// 2. Setup Output Dir
	outDir := t.TempDir()

	// 3. Generate
	gen := &contextdocs.Generator{
		Index:  &index,
		OutDir: outDir,
	}
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// 4. Compare with Golden
	goldenDir := filepath.Join("testdata", "golden", "context")
	files, err := os.ReadDir(goldenDir)
	if err != nil {
		t.Fatalf("reading golden dir: %v", err)
	}

	for _, f := range files {
		want, err := os.ReadFile(filepath.Join(goldenDir, f.Name()))
		if err != nil {
			t.Fatalf("reading golden file %s: %v", f.Name(), err)
		}

		got, err := os.ReadFile(filepath.Join(outDir, f.Name()))
		if err != nil {
			t.Errorf("missing generated file %s: %v", f.Name(), err)
			continue
		}

		if string(got) != string(want) {
			t.Errorf("content mismatch for %s:\nGOT:\n%s\nWANT:\n%s", f.Name(), got, want)
		}
	}
}
