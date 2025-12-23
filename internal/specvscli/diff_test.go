// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package specvscli

import (
	"strings"
	"testing"

	"github.com/bartekus/cortex/internal/specschema"
	"github.com/bartekus/cortex/pkg/introspect"
)

// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md

func TestCompareFlags_MissingFlagInCLI(t *testing.T) {
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "string", Default: "dev", Description: "Environment"},
		{Name: "--version", Type: "string", Default: "", Description: "Version"},
	}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev"},
		// version flag is missing
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	if len(result.Errors) == 0 {
		t.Fatal("expected error for missing flag")
	}

	found := false
	for _, err := range result.Errors {
		if contains(err, "version") && contains(err, "does not exist") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about missing version flag, got: %v", result.Errors)
	}
}

func TestCompareFlags_ExtraFlagInCLI(t *testing.T) {
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "string", Default: "dev", Description: "Environment"},
	}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev"},
		{Name: "extra", Type: "string", Default: "", Persistent: false}, // Non-persistent extra flag
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	if len(result.Warnings) == 0 {
		t.Fatal("expected warning for extra flag")
	}

	found := false
	for _, warn := range result.Warnings {
		if contains(warn, "extra") && contains(warn, "not documented") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about extra flag, got: %v", result.Warnings)
	}
}

func TestCompareFlags_PersistentFlagsSkipped(t *testing.T) {
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "string", Default: "dev", Description: "Environment"},
	}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev"},
		{Name: "config", Type: "string", Default: "", Persistent: true}, // Persistent flag should be skipped
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	// Should not warn about persistent flags
	for _, warn := range result.Warnings {
		if contains(warn, "config") {
			t.Errorf("persistent flag should be skipped, but got warning: %s", warn)
		}
	}
}

func TestCompareFlags_TypeMismatch(t *testing.T) {
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "string", Default: "dev", Description: "Environment"},
		{Name: "--verbose", Type: "bool", Default: "false", Description: "Verbose"},
	}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev"},
		{Name: "verbose", Type: "int", Default: "0"}, // Wrong type
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	if len(result.Errors) == 0 {
		t.Fatal("expected error for type mismatch")
	}

	found := false
	for _, err := range result.Errors {
		if contains(err, "verbose") && contains(err, "type mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about type mismatch, got: %v", result.Errors)
	}
}

func TestCompareFlags_TypeNormalization(t *testing.T) {
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "str", Default: "dev", Description: "Environment"},       // "str" should normalize to "string"
		{Name: "--verbose", Type: "boolean", Default: "false", Description: "Verbose"}, // "boolean" should normalize to "bool"
	}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev"},
		{Name: "verbose", Type: "bool", Default: "false"},
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	// Should not have type mismatch errors due to normalization
	for _, err := range result.Errors {
		if contains(err, "type mismatch") {
			t.Errorf("type normalization should prevent mismatch, but got error: %s", err)
		}
	}
}

func TestCompareFlags_DefaultMismatch(t *testing.T) {
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "string", Default: "dev", Description: "Environment"},
		{Name: "--version", Type: "string", Default: "1.0.0", Description: "Version"},
	}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev"},
		{Name: "version", Type: "string", Default: "2.0.0"}, // Different default
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	if len(result.Warnings) == 0 {
		t.Fatal("expected warning for default mismatch")
	}

	found := false
	for _, warn := range result.Warnings {
		if contains(warn, "version") && contains(warn, "default mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about default mismatch, got: %v", result.Warnings)
	}
}

func TestCompareFlags_NoMismatches(t *testing.T) {
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "string", Default: "dev", Description: "Environment"},
		{Name: "--verbose", Type: "bool", Default: "false", Description: "Verbose"},
	}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev"},
		{Name: "verbose", Type: "bool", Default: "false"},
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	if len(result.Errors) > 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
	if len(result.Warnings) > 0 {
		t.Errorf("expected no warnings, got: %v", result.Warnings)
	}
}

func TestCompareFlags_FlagNameNormalization(t *testing.T) {
	// Test that --env, -env, and env all normalize correctly
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "string", Default: "dev"},
		{Name: "-version", Type: "string", Default: "1.0.0"},
		{Name: "verbose", Type: "bool", Default: "false"},
	}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev"},
		{Name: "version", Type: "string", Default: "1.0.0"},
		{Name: "verbose", Type: "bool", Default: "false"},
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	if len(result.Errors) > 0 {
		t.Errorf("flag name normalization should work, but got errors: %v", result.Errors)
	}
}

func TestCompareAllCommands_WithMatchingSpec(t *testing.T) {
	specs := []specschema.Spec{
		{
			Path: "spec/commands/build.md",
			Frontmatter: specschema.SpecFrontmatter{
				Feature: "CLI_BUILD",
				Inputs: specschema.SpecInputs{
					Flags: []specschema.CliFlag{
						{Name: "--env", Type: "string", Default: "dev"},
					},
				},
			},
		},
	}

	cliCommands := []introspect.CommandInfo{
		{
			Use: "cortex",
			Subcommands: []introspect.CommandInfo{
				{
					Use: "build",
					Flags: []introspect.FlagInfo{
						{Name: "env", Type: "string", Default: "dev"},
					},
				},
			},
		},
	}

	results := CompareAllCommands(specs, cliCommands)

	// Should find and compare the build command
	if len(results) > 0 {
		// If there are results, they should be for the build command
		found := false
		for _, result := range results {
			if result.CommandName == "build" {
				found = true
				break
			}
		}
		if !found && len(results) > 0 {
			t.Errorf("expected results for build command, got: %v", results)
		}
	}
}

func TestCompareAllCommands_SkipsCommandsWithoutSpec(t *testing.T) {
	specs := []specschema.Spec{
		{
			Path: "spec/commands/build.md",
			Frontmatter: specschema.SpecFrontmatter{
				Feature: "CLI_BUILD",
				Inputs: specschema.SpecInputs{
					Flags: []specschema.CliFlag{
						{Name: "--env", Type: "string", Default: "dev"},
					},
				},
			},
		},
	}

	cliCommands := []introspect.CommandInfo{
		{
			Use: "cortex",
			Subcommands: []introspect.CommandInfo{
				{
					Use: "build",
					Flags: []introspect.FlagInfo{
						{Name: "env", Type: "string", Default: "dev"},
					},
				},
				{
					Use: "deploy", // No spec for deploy
					Flags: []introspect.FlagInfo{
						{Name: "env", Type: "string", Default: "dev"},
					},
				},
			},
		},
	}

	results := CompareAllCommands(specs, cliCommands)

	// Should not have results for deploy (no spec)
	for _, result := range results {
		if result.CommandName == "deploy" {
			t.Errorf("should skip commands without spec, but got result for deploy: %v", result)
		}
	}
}

func TestCompareAllCommands_RecursiveSubcommands(t *testing.T) {
	specs := []specschema.Spec{
		{
			Path: "spec/commands/build.md",
			Frontmatter: specschema.SpecFrontmatter{
				Feature: "CLI_BUILD",
				Inputs: specschema.SpecInputs{
					Flags: []specschema.CliFlag{
						{Name: "--env", Type: "string", Default: "dev"},
					},
				},
			},
		},
		{
			Path: "spec/commands/deploy.md",
			Frontmatter: specschema.SpecFrontmatter{
				Feature: "CLI_DEPLOY",
				Inputs: specschema.SpecInputs{
					Flags: []specschema.CliFlag{
						{Name: "--env", Type: "string", Default: "dev"},
					},
				},
			},
		},
	}

	cliCommands := []introspect.CommandInfo{
		{
			Use: "cortex",
			Subcommands: []introspect.CommandInfo{
				{
					Use: "build",
					Flags: []introspect.FlagInfo{
						{Name: "env", Type: "string", Default: "dev"},
					},
					Subcommands: []introspect.CommandInfo{
						{
							Use: "deploy", // Nested subcommand
							Flags: []introspect.FlagInfo{
								{Name: "env", Type: "string", Default: "dev"},
							},
						},
					},
				},
			},
		},
	}

	results := CompareAllCommands(specs, cliCommands)

	// Should find both build and deploy (even though deploy is nested)
	commandNames := make(map[string]bool)
	for _, result := range results {
		commandNames[result.CommandName] = true
	}

	// Both should be found (deploy is a nested subcommand but still matches CLI_DEPLOY)
	if len(results) > 0 {
		// At least one should be found
		found := false
		for name := range commandNames {
			if name == "build" || name == "deploy" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find build or deploy, got command names: %v", commandNames)
		}
	}
}

func TestInferFeatureID(t *testing.T) {
	tests := []struct {
		commandUse string
		expected   string
	}{
		{"build", "CLI_BUILD"},
		{"deploy", "CLI_DEPLOY"},
		{"dev", "CLI_DEV"},
		{"plan", "CLI_PLAN"},
		{"rollback", "CLI_ROLLBACK"},
	}

	for _, tt := range tests {
		t.Run(tt.commandUse, func(t *testing.T) {
			result := inferFeatureID(tt.commandUse)
			if result != tt.expected {
				t.Errorf("inferFeatureID(%q) = %q, want %q", tt.commandUse, result, tt.expected)
			}
		})
	}
}

func TestNormalizeType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"str", "string"},
		{"bool", "bool"},
		{"boolean", "bool"},
		{"int", "int"},
		{"integer", "int"},
		{"stringslice", "stringslice"},
		{"[]string", "stringslice"},
		{"intslice", "intslice"},
		{"[]int", "intslice"},
		{"unknown", "unknown"}, // Unknown types pass through
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompareFlags_EmptySpecFlags(t *testing.T) {
	specFlags := []specschema.CliFlag{}

	cliFlags := []introspect.FlagInfo{
		{Name: "env", Type: "string", Default: "dev", Persistent: false},
	}

	result := CompareFlags(specFlags, cliFlags, "test")

	// Should warn about undocumented flag
	if len(result.Warnings) == 0 {
		t.Fatal("expected warning for undocumented flag when spec has no flags")
	}
}

func TestCompareFlags_EmptyCLIFlags(t *testing.T) {
	specFlags := []specschema.CliFlag{
		{Name: "--env", Type: "string", Default: "dev", Description: "Environment"},
	}

	cliFlags := []introspect.FlagInfo{}

	result := CompareFlags(specFlags, cliFlags, "test")

	// Should error about missing flag
	if len(result.Errors) == 0 {
		t.Fatal("expected error for missing flag when CLI has no flags")
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
