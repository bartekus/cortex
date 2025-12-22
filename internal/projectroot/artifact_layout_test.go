package projectroot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Feature: REL_ARTIFACT_LAYOUT
// Spec: spec/release/contract.md

func TestRelArtifactLayout(t *testing.T) {
	root, err := Find(".")
	if err != nil {
		t.Fatalf("failed to get project root: %v", err)
	}

	// Assert .goreleaser.yaml exists
	configPath := filepath.Join(root, ".goreleaser.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("missing release config: .goreleaser.yaml")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	sContent := string(content)

	// Assert Anchor Keys for Artifact Layout
	// We expect specific configuration keys that define the release contract
	requiredAnchors := []string{
		"project_name:",
		"builds:",
		"archives:",
		// We expect the 'cortex' binary to be defined
		"- cortex",
	}

	for _, anchor := range requiredAnchors {
		if !strings.Contains(sContent, anchor) {
			t.Errorf("missing required anchor %q in .goreleaser.yaml", anchor)
		}
	}
}
