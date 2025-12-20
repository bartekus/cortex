// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Cortex - Cortex is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bartekus/cortex/internal/builder"
	"github.com/bartekus/cortex/internal/projectroot"
	"github.com/bartekus/cortex/internal/xray"

	"github.com/spf13/cobra"
)

// NewContextCommand returns the `cortex context` command group.
func NewContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "AI context pipeline commands",
		Long:  "Commands for building and managing AI-readable context representations of the repository.",
	}

	cmd.AddCommand(NewContextBuildCommand())
	cmd.AddCommand(NewContextDocsCommand())
	cmd.AddCommand(NewContextXrayCommand())

	// Shared flag for all context commands (needed by build and xray)
	cmd.PersistentFlags().String("xray-bin", "", "Path to xray binary")

	return cmd
}

// NewContextBuildCommand returns the `cortex context build` command.
func NewContextBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build AI context representation",
		Long:  "Builds a deterministic AI-readable context representation in .cortex/",
		RunE:  runContextBuild,
	}
}

// NewContextXrayCommand returns the `cortex context xray` command.
func NewContextXrayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xray [subcommand] [flags]",
		Short: "Run XRAY scan",
		Long:  "Runs XRAY scan (always against repository root) to analyze repository structure and dependencies.",
		RunE:  runContextXray,
	}

	// Flag moved to parent
	// cmd.PersistentFlags().String("xray-bin", "", "Path to xray binary")

	scanCmd := &cobra.Command{
		Use:   "scan [target]",
		Short: "Run XRAY scan",
		Long:  "Runs XRAY scan against the specified target (default: .) and writes index.json to the output directory.",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(c *cobra.Command, args []string) error {
			// Parse target
			target := "."
			if len(args) == 1 {
				target = args[0]
			}

			out, err := c.Flags().GetString("output")
			if err != nil {
				return fmt.Errorf("reading --output: %w", err)
			}
			if out == "" {
				// Enforce explicit output per contract
				repoRoot, err := projectroot.Find(".")
				if err != nil {
					return fmt.Errorf("finding repo root: %w", err)
				}
				// slug := filepath.Base(repoRoot)
				// out = filepath.Join(repoRoot, ".cortex", slug, "data")
				out = filepath.Join(repoRoot, ".cortex", "data")
			}

			// Forward args in the Rust CLI order: scan <target> --output <dir>
			xrayArgs := []string{target, "--output", out}
			return runXraySubcommand(c, "scan", xrayArgs)
		},
	}
	scanCmd.Flags().String("output", "", "Output directory for index.json (default: .cortex/<slug>/data)")
	cmd.AddCommand(scanCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "docs",
		Short: "Run XRAY docs",
		RunE: func(c *cobra.Command, args []string) error {
			return runXraySubcommand(c, "docs", args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "Run XRAY all",
		RunE: func(c *cobra.Command, args []string) error {
			return runXraySubcommand(c, "all", args)
		},
	})

	return cmd
}

// NewContextDocsCommand returns the `cortex context docs` command.
func NewContextDocsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "docs",
		Short: "Generate AI-Agent documentation",
		Long:  "Generates human-readable documentation from AI-Agent outputs (chunks.ndjson, manifest.json, XRAY index.json)",
		RunE:  runContextDocs,
		Args:  cobra.NoArgs,
	}
}

// runContextBuild executes the context:build npm script.
func runContextBuild(cmd *cobra.Command, _ []string) error {
	repoRoot, err := projectroot.Find(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[cortex] building AI context...\n")

	// 1. Run XRAY scan (Rust)
	// slug := filepath.Base(repoRoot)
	// outputDir := filepath.Join(repoRoot, ".cortex", slug, "data")
	outputDir := filepath.Join(repoRoot, ".cortex", "data")

	// Rust CLI order: scan <target> --output <dir>
	xrayArgs := []string{".", "--output", outputDir}

	if err := runXraySubcommand(cmd, "scan", xrayArgs); err != nil {
		return fmt.Errorf("xray scan pre-build failed: %w", err)
	}

	// 2. Read XRAY Index
	indexPath := filepath.Join(outputDir, "index.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read xray index at %s: %w", indexPath, err)
	}

	var index xray.Index
	if err := json.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("unmarshaling xray index: %w", err)
	}

	// 3. Build .cortex structure
	if err := builder.BuildContext(repoRoot, &index); err != nil {
		return fmt.Errorf("building .cortex: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[cortex] AI context ready → .cortex/\n")

	return nil
}

// runContextXray acts as a fallback if no subcommand given?
// Or we force subcommands.
func runContextXray(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func resolveXrayBin(cmd *cobra.Command) (string, error) {
	// 1. Flag
	bin, _ := cmd.Flags().GetString("xray-bin")
	if bin != "" {
		return bin, nil
	}

	// 2. Env
	if bin = os.Getenv("XRAY_BIN"); bin != "" {
		return bin, nil
	}

	// 3. Default (Repo Relative)
	// We need repo root.
	repoRoot, err := projectroot.Find(".")
	if err != nil {
		return "", fmt.Errorf("finding repo root: %w", err)
	}

	// Try release first, then debug
	releasePath := filepath.Join(repoRoot, "rust/target/release/xray")
	if _, err := os.Stat(releasePath); err == nil {
		return releasePath, nil
	}

	debugPath := filepath.Join(repoRoot, "rust/target/debug/xray")
	if _, err := os.Stat(debugPath); err == nil {
		return debugPath, nil
	}

	return "", fmt.Errorf("xray binary not found. Build it with `cargo build` in rust/xray/ or specify --xray-bin")
}

func runXraySubcommand(cmd *cobra.Command, sub string, args []string) error {
	repoRoot, err := projectroot.Find(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	bin, err := resolveXrayBin(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// User Req: "Cortex wrappers must pass --output explicitly"
	// For 'context build', caller must provide these args.
	// For 'xray scan' (CLI), user provides them or we rely on defaults?
	// If we want constraints, caller handles them.

	// Construct final args: [subcommand, ...args]
	// Filter out subcommand if it's already in args? No, args from cobra exclude command name.
	xrayArgs := []string{sub}
	xrayArgs = append(xrayArgs, args...)

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[cortex] Invoking XRAY: %s %v\n", bin, xrayArgs)

	execCmd := exec.CommandContext(ctx, bin, xrayArgs...)
	execCmd.Dir = repoRoot
	execCmd.Stdout = cmd.OutOrStdout()
	execCmd.Stderr = cmd.ErrOrStderr()

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("xray %s failed: %w", sub, err)
	}

	return nil
}

// runContextDocs executes the context:docs npm script.
func runContextDocs(cmd *cobra.Command, _ []string) error {
	_, err := projectroot.Find(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	// Deprecation Notice (Option A)
	return fmt.Errorf("context docs generation via Node is deprecated. Docs projection will be reimplemented in Go in Phase 4")

	// _, _ = fmt.Fprintf(cmd.OutOrStdout(), "[cortex] generating AI-Agent docs...\n")
	// ... removed ...
}
