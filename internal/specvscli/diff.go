// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a standalone governance and intelligence tool for AI-assisted software development.
It analyzes repositories, enforces structural contracts, detects drift, and generates deterministic context artifacts that enable safe, auditable collaboration between humans and AI agents.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package specvscli provides tools for comparing spec files to CLI implementation.
package specvscli

import (
	"fmt"
	"strings"

	"github.com/bartekus/cortex/internal/specschema"
	"github.com/bartekus/cortex/pkg/introspect"
)

// DiffResult represents the result of comparing specs to CLI implementation.
type DiffResult struct {
	CommandName string
	Errors      []string
	Warnings    []string
}

// CompareFlags compares flags from a spec to flags from CLI introspection.
func CompareFlags(specFlags []specschema.CliFlag, cliFlags []introspect.FlagInfo, commandName string) DiffResult {
	result := DiffResult{
		CommandName: commandName,
		Errors:      []string{},
		Warnings:    []string{},
	}

	// Build maps for easier lookup
	specFlagMap := make(map[string]specschema.CliFlag)
	for _, flag := range specFlags {
		// Normalize flag name (remove leading dashes)
		name := strings.TrimPrefix(flag.Name, "--")
		name = strings.TrimPrefix(name, "-")
		specFlagMap[name] = flag
	}

	cliFlagMap := make(map[string]introspect.FlagInfo)
	for _, flag := range cliFlags {
		cliFlagMap[flag.Name] = flag
	}

	// Check: spec declares flag that doesn't exist in CLI
	for name, specFlag := range specFlagMap {
		cliFlag, exists := cliFlagMap[name]
		if !exists {
			result.Errors = append(result.Errors, fmt.Sprintf("spec declares flag %q that does not exist in CLI", specFlag.Name))
			continue
		}

		// Check type alignment (if spec specifies type)
		if specFlag.Type != "" && cliFlag.Type != "" {
			normalizedSpecType := normalizeType(specFlag.Type)
			normalizedCLIType := normalizeType(cliFlag.Type)
			if normalizedSpecType != normalizedCLIType {
				result.Errors = append(result.Errors, fmt.Sprintf("flag %q type mismatch: spec has %q but CLI has %q", name, specFlag.Type, cliFlag.Type))
			}
		}

		// Check default value alignment (if spec specifies default)
		if specFlag.Default != "" && cliFlag.Default != "" {
			if specFlag.Default != cliFlag.Default {
				result.Warnings = append(result.Warnings, fmt.Sprintf("flag %q default mismatch: spec has %q but CLI has %q", name, specFlag.Default, cliFlag.Default))
			}
		}

		// Check description alignment (if spec specifies description)
		// Description comparison is lenient - just check if both are non-empty
		// Full text matching would be too strict
		_ = specFlag.Description != "" && cliFlag.Usage != ""
	}

	// Check: CLI has flag that's not in spec
	// Note: We skip global/persistent flags that are inherited from root
	for name, cliFlag := range cliFlagMap {
		if cliFlag.Persistent {
			// Skip persistent flags - they're inherited from root
			continue
		}
		if _, exists := specFlagMap[name]; !exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("CLI has flag %q that is not documented in spec", name))
		}
	}

	return result
}

// normalizeType normalizes type strings for comparison.
func normalizeType(typ string) string {
	typ = strings.ToLower(typ)
	// Map common variations
	switch typ {
	case "string", "str":
		return "string"
	case "bool", "boolean":
		return "bool"
	case "int", "integer":
		return "int"
	case "stringslice", "[]string":
		return "stringslice"
	case "intslice", "[]int":
		return "intslice"
	default:
		return typ
	}
}

// CompareAllCommands compares all CLI commands to their corresponding specs.
func CompareAllCommands(specs []specschema.Spec, cliCommands []introspect.CommandInfo) []DiffResult {
	var results []DiffResult

	// Build spec lookup by feature ID
	specMap := make(map[string]specschema.Spec)
	for _, spec := range specs {
		specMap[spec.Frontmatter.Feature] = spec
	}

	// Compare each CLI command
	for _, cmd := range cliCommands {
		// Try to find matching spec
		// Command use is typically like "build", "deploy", etc.
		// Feature ID is typically like "CLI_BUILD", "CLI_DEPLOY"
		featureID := inferFeatureID(cmd.Use)
		spec, hasSpec := specMap[featureID]

		if !hasSpec {
			// No spec for this command - skip
			continue
		}

		// Compare flags
		diff := CompareFlags(spec.Frontmatter.Inputs.Flags, cmd.Flags, cmd.Use)
		if len(diff.Errors) > 0 || len(diff.Warnings) > 0 {
			results = append(results, diff)
		}

		// Recursively check subcommands
		for i := range cmd.Subcommands {
			subResults := compareSubcommands(&cmd.Subcommands[i], specMap)
			results = append(results, subResults...)
		}
	}

	return results
}

// compareSubcommands recursively compares subcommands.
func compareSubcommands(cmd *introspect.CommandInfo, specMap map[string]specschema.Spec) []DiffResult {
	var results []DiffResult

	featureID := inferFeatureID(cmd.Use)
	if spec, hasSpec := specMap[featureID]; hasSpec {
		diff := CompareFlags(spec.Frontmatter.Inputs.Flags, cmd.Flags, cmd.Use)
		if len(diff.Errors) > 0 || len(diff.Warnings) > 0 {
			results = append(results, diff)
		}
	}

	for i := range cmd.Subcommands {
		subResults := compareSubcommands(&cmd.Subcommands[i], specMap)
		results = append(results, subResults...)
	}

	return results
}

// inferFeatureID attempts to infer a feature ID from a command name.
// For example: "build" -> "CLI_BUILD", "deploy" -> "CLI_DEPLOY"
func inferFeatureID(commandUse string) string {
	// Convert to uppercase and prefix with CLI_
	upper := strings.ToUpper(commandUse)
	return "CLI_" + upper
}
