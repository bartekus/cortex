// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package cliintrospect provides CLI command introspection functionality.
package introspect

import (
	"testing"

	"github.com/spf13/cobra"
)

// Feature: CLI_COMMAND_GOV
// Spec: spec/cli/gov.md

func TestIntrospect_RootCommand(t *testing.T) {
	root := &cobra.Command{
		Use:   "cortex",
		Short: "Test CLI",
	}

	root.PersistentFlags().String("config", "", "config file")
	root.PersistentFlags().Bool("verbose", false, "verbose output")

	commands := Introspect(root)
	if len(commands) == 0 {
		t.Fatal("expected at least one command")
	}

	rootCmd := commands[0]
	if rootCmd.Use != "cortex" {
		t.Errorf("expected Use 'cortex', got %q", rootCmd.Use)
	}

	// Check persistent flags are included
	if len(rootCmd.Flags) < 2 {
		t.Errorf("expected at least 2 flags, got %d", len(rootCmd.Flags))
	}
}

func TestIntrospect_WithSubcommands(t *testing.T) {
	root := &cobra.Command{
		Use:   "cortex",
		Short: "Test CLI",
	}

	sub1 := &cobra.Command{
		Use:   "build",
		Short: "Build command",
	}
	sub1.Flags().String("version", "", "version")

	sub2 := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy command",
	}
	sub2.Flags().String("env", "", "environment")

	root.AddCommand(sub1)
	root.AddCommand(sub2)

	commands := Introspect(root)
	if len(commands) == 0 {
		t.Fatal("expected at least one command")
	}

	rootCmd := commands[0]
	if len(rootCmd.Subcommands) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(rootCmd.Subcommands))
	}
}

func TestFlagToInfo_StringFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().StringP("env", "e", "dev", "Environment")

	flags := collectFlags(cmd)
	if len(flags) == 0 {
		t.Fatal("expected at least one flag")
	}

	info := flags[0]
	if info.Name != "env" {
		t.Errorf("expected name 'env', got %q", info.Name)
	}
	if info.Shorthand != "e" {
		t.Errorf("expected shorthand 'e', got %q", info.Shorthand)
	}
	if info.Usage != "Environment" {
		t.Errorf("expected usage 'Environment', got %q", info.Usage)
	}
	if info.Default != "dev" {
		t.Errorf("expected default 'dev', got %q", info.Default)
	}
}

func TestFlagToInfo_BoolFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.PersistentFlags().Bool("verbose", false, "Verbose output")

	flags := collectFlags(cmd)
	if len(flags) == 0 {
		t.Fatal("expected at least one flag")
	}

	info := flags[0]
	if info.Type != "bool" {
		t.Errorf("expected type 'bool', got %q", info.Type)
	}
	if !info.Persistent {
		t.Error("expected persistent to be true")
	}
}

func TestFindCommand(t *testing.T) {
	commands := []CommandInfo{
		{
			Use:   "cortex",
			Short: "Root",
			Subcommands: []CommandInfo{
				{Use: "build", Short: "Build"},
				{Use: "deploy", Short: "Deploy"},
			},
		},
	}

	// Find root
	cmd := FindCommand(commands, "cortex")
	if cmd == nil {
		t.Fatal("expected to find root command")
	}

	// Find subcommand
	cmd = FindCommand(commands, "build")
	if cmd == nil {
		t.Fatal("expected to find build command")
		return
	}
	if cmd.Use != "build" {
		t.Errorf("expected Use 'build', got %q", cmd.Use)
	}
}

func TestGetCommandFlags(t *testing.T) {
	commands := []CommandInfo{
		{
			Use:   "cortex",
			Flags: []FlagInfo{{Name: "config", Type: "string"}},
			Subcommands: []CommandInfo{
				{
					Use:   "build",
					Flags: []FlagInfo{{Name: "version", Type: "string"}},
				},
			},
		},
	}

	flags := GetCommandFlags(commands, "cortex build")
	if len(flags) == 0 {
		t.Fatal("expected to find flags")
	}

	found := false
	for _, flag := range flags {
		if flag.Name == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'version' flag")
	}
}
