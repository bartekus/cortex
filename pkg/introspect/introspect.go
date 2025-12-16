// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package introspect

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandInfo represents information about a Cobra command and its flags.
type CommandInfo struct {
	Use         string        `json:"use"`
	Short       string        `json:"short"`
	Long        string        `json:"long"`
	Flags       []FlagInfo    `json:"flags"`
	Subcommands []CommandInfo `json:"subcommands,omitempty"`
}

// FlagInfo represents information about a CLI flag.
type FlagInfo struct {
	Name       string `json:"name"`
	Shorthand  string `json:"shorthand"`
	Type       string `json:"type"`
	Default    string `json:"default"`
	Usage      string `json:"usage"`
	Persistent bool   `json:"persistent"`
	Required   bool   `json:"required"`
}

// Introspect introspects a Cobra command tree and returns information about all commands and flags.
func Introspect(root *cobra.Command) []CommandInfo {
	var commands []CommandInfo
	collectCommands(root, &commands, true)
	return commands
}

// collectCommands recursively collects command information.
func collectCommands(cmd *cobra.Command, commands *[]CommandInfo, includeRoot bool) {
	if !includeRoot && !cmd.HasSubCommands() {
		return
	}

	info := CommandInfo{
		Use:   cmd.Use,
		Short: cmd.Short,
		Long:  cmd.Long,
		Flags: collectFlags(cmd),
	}

	// Collect subcommands (direct children only, they will recursively collect their own subcommands)
	var subcommands []CommandInfo
	allSubcmds := cmd.Commands()
	for _, subcmd := range allSubcmds {
		// Skip help topic commands (but only if they're actually help topics, not regular commands)
		// A command is a help topic if it's marked as such AND has no parent (auto-generated)
		// Commands explicitly added via AddCommand should have a parent set
		if subcmd.IsAdditionalHelpTopicCommand() && subcmd.Parent() == nil {
			continue
		}
		// Skip if command is hidden
		if subcmd.Hidden {
			continue
		}
		// Collect this subcommand and its descendants recursively
		var subcmdInfos []CommandInfo
		collectCommands(subcmd, &subcmdInfos, true)
		// The first item in subcmdInfos is the subcommand itself (because includeRoot=true)
		// Append it to subcommands
		if len(subcmdInfos) > 0 {
			subcommands = append(subcommands, subcmdInfos[0])
		}
	}
	info.Subcommands = subcommands

	*commands = append(*commands, info)
}

// collectFlags extracts flag information from a Cobra command.
// Flags are sorted by name for deterministic output.
func collectFlags(cmd *cobra.Command) []FlagInfo {
	var flags []FlagInfo
	flagMap := make(map[string]FlagInfo)

	// Collect persistent flags defined on this command first
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flagMap[flag.Name] = flagToInfo(flag, true)
	})

	// Collect local (non-persistent) flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		// Only add if not already collected as persistent
		if _, exists := flagMap[flag.Name]; !exists {
			flagMap[flag.Name] = flagToInfo(flag, false)
		}
	})

	// Collect inherited flags from parent commands (but only if they're not already collected)
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if _, exists := flagMap[flag.Name]; !exists {
			flagMap[flag.Name] = flagToInfo(flag, true)
		}
	})

	// Convert map to slice and sort by name
	flagNames := make([]string, 0, len(flagMap))
	for name := range flagMap {
		flagNames = append(flagNames, name)
	}
	sort.Strings(flagNames)

	for _, name := range flagNames {
		flags = append(flags, flagMap[name])
	}

	return flags
}

// flagToInfo converts a pflag.Flag to FlagInfo.
func flagToInfo(flag *pflag.Flag, persistent bool) FlagInfo {
	info := FlagInfo{
		Name:       flag.Name,
		Shorthand:  flag.Shorthand,
		Usage:      flag.Usage,
		Persistent: persistent,
		Required:   false, // pflag doesn't expose this directly, would need custom logic
	}

	// Determine type
	info.Type = inferFlagType(flag)

	// Get default value
	if flag.DefValue != "" {
		info.Default = flag.DefValue
	} else {
		info.Default = inferDefaultValue(flag)
	}

	return info
}

// inferFlagType attempts to infer the flag type from its value representation.
func inferFlagType(flag *pflag.Flag) string {
	valueType := flag.Value.Type()
	switch valueType {
	case "bool":
		return "bool"
	case "string":
		return "string"
	case "stringSlice":
		return "stringSlice"
	case "int":
		return "int"
	case "intSlice":
		return "intSlice"
	case "duration":
		return "duration"
	default:
		// Try to infer from default value
		if strings.Contains(valueType, "bool") {
			return "bool"
		}
		if strings.Contains(valueType, "string") {
			return "string"
		}
		return valueType
	}
}

// inferDefaultValue attempts to get a string representation of the default value.
func inferDefaultValue(flag *pflag.Flag) string {
	if flag.DefValue != "" {
		return flag.DefValue
	}
	// For boolean flags, default is typically "false"
	if inferFlagType(flag) == "bool" {
		return "false"
	}
	return ""
}

// FindCommand finds a command by its Use string in the command tree.
func FindCommand(commands []CommandInfo, use string) *CommandInfo {
	for i := range commands {
		if commands[i].Use == use {
			return &commands[i]
		}
		// Search subcommands
		if found := FindCommand(commands[i].Subcommands, use); found != nil {
			return found
		}
	}
	return nil
}

// GetAllCommandPaths returns all command paths (e.g., "stagecraft build", "stagecraft deploy").
func GetAllCommandPaths(root *cobra.Command) []string {
	var paths []string
	collectPaths(root, "", &paths)
	return paths
}

// collectPaths recursively collects all command paths.
func collectPaths(cmd *cobra.Command, prefix string, paths *[]string) {
	currentPath := cmd.Use
	if prefix != "" {
		currentPath = prefix + " " + currentPath
	}

	if cmd.HasSubCommands() {
		*paths = append(*paths, currentPath)
	}

	for _, subcmd := range cmd.Commands() {
		if !subcmd.IsAvailableCommand() || subcmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		collectPaths(subcmd, currentPath, paths)
	}
}

// GetCommandFlags returns all flags for a specific command path.
func GetCommandFlags(commands []CommandInfo, commandPath string) []FlagInfo {
	parts := strings.Fields(commandPath)
	if len(parts) == 0 {
		return nil
	}

	// Find root command
	var current *CommandInfo
	for i := range commands {
		if commands[i].Use == parts[0] {
			current = &commands[i]
			break
		}
	}

	if current == nil {
		return nil
	}

	// Navigate to the target command
	for i := 1; i < len(parts); i++ {
		found := false
		for j := range current.Subcommands {
			if current.Subcommands[j].Use == parts[i] {
				current = &current.Subcommands[j]
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	return current.Flags
}
