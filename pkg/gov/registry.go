package gov

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type FeatureStatus string

const (
	StatusDraft       FeatureStatus = "draft"
	StatusReview      FeatureStatus = "review"
	StatusApproved    FeatureStatus = "approved"
	StatusImplemented FeatureStatus = "implemented"
	StatusDeprecated  FeatureStatus = "deprecated"
)

type Feature struct {
	ID        string        `yaml:"id"`
	Title     string        `yaml:"title"`
	Status    FeatureStatus `yaml:"status"`
	Spec      string        `yaml:"spec"`
	Owner     string        `yaml:"owner"`
	Group     string        `yaml:"group"`
	Tests     []string      `yaml:"tests"`
	DependsOn []string      `yaml:"depends_on"`
}

type Registry struct {
	Features []Feature `yaml:"features"`
}

func LoadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file: %w", err)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry YAML: %w", err)
	}

	return &reg, nil
}

func (r *Registry) Validate() error {
	seenIDs := make(map[string]bool)

	for i, f := range r.Features {
		// Required fields
		if f.ID == "" {
			return fmt.Errorf("feature at index %d missing ID", i)
		}
		if f.Title == "" {
			return fmt.Errorf("feature %s missing title", f.ID)
		}
		if f.Status == "" {
			return fmt.Errorf("feature %s missing status", f.ID)
		}
		if f.Spec == "" {
			return fmt.Errorf("feature %s missing spec", f.ID)
		}
		if f.Owner == "" {
			return fmt.Errorf("feature %s missing owner", f.ID)
		}
		if f.Group == "" {
			return fmt.Errorf("feature %s missing group", f.ID)
		}

		// Status enum
		switch f.Status {
		case StatusDraft, StatusReview, StatusApproved, StatusImplemented, StatusDeprecated:
			// valid
		default:
			return fmt.Errorf("feature %s has invalid status: %s", f.ID, f.Status)
		}

		// Duplicate IDs
		if seenIDs[f.ID] {
			return fmt.Errorf("duplicate feature ID: %s", f.ID)
		}
		seenIDs[f.ID] = true

		// Spec paths
		if filepath.IsAbs(f.Spec) {
			return fmt.Errorf("feature %s spec path must be relative: %s", f.ID, f.Spec)
		}
		if strings.HasPrefix(f.Spec, "../") || strings.Contains(f.Spec, "/../") {
			return fmt.Errorf("feature %s spec path must not escape repo root: %s", f.ID, f.Spec)
		}
	}

	return nil
}
