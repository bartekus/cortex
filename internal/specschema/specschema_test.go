// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package specschema

import (
	"os"
	"path/filepath"
	"testing"
)

// Feature: GOV_CORE
// Spec: spec/governance/GOV_CORE.md

func TestExtractFrontmatter_ValidFrontmatter(t *testing.T) {
	content := `---
feature: TEST_FEATURE
version: v1
status: done
domain: test
inputs:
  flags: []
outputs:
  exit_codes:
    success: 0
---
# Test Feature
Content here.
`

	fm, err := ExtractFrontmatter(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if fm.Feature != "TEST_FEATURE" {
		t.Errorf("expected feature 'TEST_FEATURE', got %q", fm.Feature)
	}
	if fm.Version != "v1" {
		t.Errorf("expected version 'v1', got %q", fm.Version)
	}
	if fm.Status != "done" {
		t.Errorf("expected status 'done', got %q", fm.Status)
	}
	if fm.Domain != "test" {
		t.Errorf("expected domain 'test', got %q", fm.Domain)
	}
}

func TestExtractFrontmatter_MissingDelimiter(t *testing.T) {
	content := `# Test Feature
No frontmatter here.
`

	_, err := ExtractFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for missing frontmatter delimiter")
	}
}

func TestExtractFrontmatter_UnclosedDelimiter(t *testing.T) {
	content := `---
feature: TEST_FEATURE
version: v1
# Missing closing ---
`

	_, err := ExtractFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for unclosed frontmatter")
	}
}

func TestExtractFrontmatter_InvalidYAML(t *testing.T) {
	content := `---
feature: TEST_FEATURE
version: v1
invalid: [unclosed
---
`

	_, err := ExtractFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestExtractFrontmatter_WithFlags(t *testing.T) {
	content := `---
feature: TEST_FEATURE
version: v1
status: wip
domain: test
inputs:
  flags:
    - name: --env
      type: string
      default: ""
      description: "Target environment"
outputs:
  exit_codes:
    success: 0
    error: 1
---
`

	fm, err := ExtractFrontmatter(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(fm.Inputs.Flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(fm.Inputs.Flags))
	}

	flag := fm.Inputs.Flags[0]
	if flag.Name != "--env" {
		t.Errorf("expected flag name '--env', got %q", flag.Name)
	}
	if flag.Type != "string" {
		t.Errorf("expected flag type 'string', got %q", flag.Type)
	}

	if len(fm.Outputs.ExitCodes) != 2 {
		t.Fatalf("expected 2 exit codes, got %d", len(fm.Outputs.ExitCodes))
	}
	if fm.Outputs.ExitCodes["success"] != 0 {
		t.Errorf("expected success=0, got %d", fm.Outputs.ExitCodes["success"])
	}
}

func TestLoadSpec_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "TEST_FEATURE.md")

	content := `---
feature: TEST_FEATURE
version: v1
status: todo
domain: test
---
# Test Feature
`

	if err := os.WriteFile(specPath, []byte(content), 0o600); err != nil { //nolint:gosec // G306: test file
		t.Fatalf("failed to write spec file: %v", err)
	}

	spec, err := LoadSpec(specPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if spec.Path != specPath {
		t.Errorf("expected path %q, got %q", specPath, spec.Path)
	}
	if spec.Frontmatter.Feature != "TEST_FEATURE" {
		t.Errorf("expected feature 'TEST_FEATURE', got %q", spec.Frontmatter.Feature)
	}
}

func TestLoadSpec_FileNotFound(t *testing.T) {
	_, err := LoadSpec("/nonexistent/path.md")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadAllSpecs_WalksDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create spec dir: %v", err)
	}

	// Create valid spec
	spec1 := filepath.Join(specDir, "FEATURE1.md")
	content1 := `---
feature: FEATURE1
version: v1
status: done
domain: test
---
`
	if err := os.WriteFile(spec1, []byte(content1), 0o644); err != nil { //nolint:gosec // test file
		t.Fatalf("failed to write spec1: %v", err)
	}

	// Create another valid spec in subdirectory
	subDir := filepath.Join(specDir, "sub")
	if err := os.MkdirAll(subDir, 0o750); err != nil { //nolint:gosec // G301: test directory
		t.Fatalf("failed to create subdir: %v", err)
	}
	spec2 := filepath.Join(subDir, "FEATURE2.md")
	content2 := `---
feature: FEATURE2
version: v1
status: wip
domain: test
---
`
	if err := os.WriteFile(spec2, []byte(content2), 0o644); err != nil { //nolint:gosec // test file
		t.Fatalf("failed to write spec2: %v", err)
	}

	// Create non-markdown file (should be ignored)
	nonMd := filepath.Join(specDir, "readme.txt")
	if err := os.WriteFile(nonMd, []byte("not a spec"), 0o644); err != nil { //nolint:gosec // test file
		t.Fatalf("failed to write non-md file: %v", err)
	}

	specs, err := LoadAllSpecs(specDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(specs) != 2 {
		t.Fatalf("expected 2 specs, got %d", len(specs))
	}
}

func TestExpectedFeatureIDFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"spec/governance/GOV_CORE.md", "GOV_CORE"},
		{"spec/commands/build.md", "build"},
		{"FEATURE.md", "FEATURE"},
		{"path/to/SPEC.md", "SPEC"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ExpectedFeatureIDFromPath(tt.path)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestValidateSpec_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{
			name: "valid spec",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Feature: "FEATURE",
					Version: "v1",
					Status:  "done",
					Domain:  "test",
				},
			},
			wantErr: false,
		},
		{
			name: "missing feature",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Version: "v1",
					Status:  "done",
					Domain:  "test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Feature: "FEATURE",
					Status:  "done",
					Domain:  "test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing status",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Feature: "FEATURE",
					Version: "v1",
					Domain:  "test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing domain",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Feature: "FEATURE",
					Version: "v1",
					Status:  "done",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Feature: "FEATURE",
					Version: "v1",
					Status:  "invalid",
					Domain:  "test",
				},
			},
			wantErr: true,
		},
		{
			name: "feature ID mismatch",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Feature: "WRONG_FEATURE",
					Version: "v1",
					Status:  "done",
					Domain:  "test",
				},
			},
			wantErr: false, // Feature ID mismatch is now validated by integrity check, not ValidateSpec
		},
		{
			name: "empty flag name",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Feature: "FEATURE",
					Version: "v1",
					Status:  "done",
					Domain:  "test",
					Inputs: SpecInputs{
						Flags: []CliFlag{
							{Name: ""},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative exit code",
			spec: Spec{
				Path: "spec/test/FEATURE.md",
				Frontmatter: SpecFrontmatter{
					Feature: "FEATURE",
					Version: "v1",
					Status:  "done",
					Domain:  "test",
					Outputs: SpecOutputs{
						ExitCodes: map[string]int{
							"error": -1,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpec(&tt.spec)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateAll_MultipleSpecs(t *testing.T) {
	specs := []Spec{
		{
			Path: "spec/test/FEATURE1.md",
			Frontmatter: SpecFrontmatter{
				Feature: "FEATURE1",
				Version: "v1",
				Status:  "done",
				Domain:  "test",
			},
		},
		{
			Path: "spec/test/FEATURE2.md",
			Frontmatter: SpecFrontmatter{
				Feature: "FEATURE2",
				Version: "v1",
				Status:  "wip",
				Domain:  "test",
			},
		},
	}

	err := ValidateAll(specs)
	if err != nil {
		t.Errorf("expected no error for valid specs, got: %v", err)
	}
}

func TestValidateAll_WithErrors(t *testing.T) {
	specs := []Spec{
		{
			Path: "spec/test/FEATURE1.md",
			Frontmatter: SpecFrontmatter{
				Feature: "FEATURE1",
				Version: "v1",
				Status:  "done",
				Domain:  "test",
			},
		},
		{
			Path: "spec/test/FEATURE2.md",
			Frontmatter: SpecFrontmatter{
				Feature: "FEATURE2",
				Version: "v1",
				Status:  "invalid",
				Domain:  "test",
			},
		},
	}

	err := ValidateAll(specs)
	if err == nil {
		t.Fatal("expected error for invalid spec")
	}
}
