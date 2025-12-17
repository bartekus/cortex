package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bartekus/cortex/internal/projectroot"
	"github.com/bartekus/cortex/internal/runner"
	"github.com/bartekus/cortex/internal/scanner"
	"github.com/bartekus/cortex/internal/skills"
)

var (
	runJSON          bool
	runStateDir      string
	runFailOnWarning bool
)

var runCmd = &cobra.Command{
	Use:   "run <command|skill> [flags]",
	Short: "Orchestrate Cortex skills and governance checks",
	Long: `Deterministiacally run skills, tests, and governance checks.
Maintains state in .cortex/run to allow resuming failures.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		// If argument is not a subcommand, treat it as a skill name
		return runSkill(cmd.Context(), args)
	},
}

func init() {
	runCmd.PersistentFlags().BoolVar(&runJSON, "json", false, "Output results in JSON")
	runCmd.PersistentFlags().StringVar(&runStateDir, "state-dir", ".cortex/run", "Directory to store run state")
	runCmd.PersistentFlags().BoolVar(&runFailOnWarning, "fail-on-warning", false, "Fail if warnings occur")

	runCmd.AddCommand(runListCmd)
	runCmd.AddCommand(runAllCmd)
	runCmd.AddCommand(runResumeCmd)
	runCmd.AddCommand(runReportCmd)
	runCmd.AddCommand(runResetCmd)

	// Register with root (assuming rootCmd exists in package, but usually it's passed or init-ed)
	// We'll export RunCmd or similar?
	// The existing pattern in Cortex seems to be manual registration in main or root.go.
	// I'll assume I need to export RunCmd or add it to RootCmd if available.
}

// GetRunCmd exposes the command to the main package.
func GetRunCmd() *cobra.Command {
	return runCmd
}

func resolveStateStore(wd string) (*runner.StateStore, error) {
	repoRoot, err := projectroot.Find(wd)
	if err != nil {
		return nil, err
	}
	stateDir := runStateDir
	if !filepath.IsAbs(stateDir) {
		stateDir = filepath.Join(repoRoot, stateDir)
	}
	return runner.NewStateStore(stateDir), nil
}

func setupRunner(ctx context.Context) (*runner.Runner, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	repoRoot, err := projectroot.Find(wd)
	if err != nil {
		return nil, err
	}

	store, err := resolveStateStore(wd)
	if err != nil {
		return nil, err
	}

	// We need the resolved state dir string for Deps
	// resolveStateStore returns *StateStore, we can ask it or resolve the path again.
	// But `runStateDir` is available and resolveStateStore uses it + logic.
	// Actually `resolvestateStore` uses logic to anchor.
	// Let's modify resolveStateStore to return path too or just duplicate logic slightly or ask store?
	// Store has `dir` field but it's private.
	// Let's expose `Dir()` on store or just use the logic here.

	// Re-using logic:
	stateDir := runStateDir
	if !filepath.IsAbs(stateDir) {
		stateDir = filepath.Join(repoRoot, stateDir)
	}

	scn := scanner.New(repoRoot)
	deps := &runner.Deps{
		RepoRoot:      repoRoot,
		StateDir:      stateDir,
		Scanner:       scn,
		FailOnWarning: runFailOnWarning,
	}

	return runner.NewRunner(skills.Registry, store, deps), nil
}

type SkillListItem struct {
	ID string `json:"id"`
}

var runListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		list := make([]SkillListItem, 0, len(skills.Registry))
		for _, s := range skills.Registry {
			list = append(list, SkillListItem{ID: s.ID()})
		}

		if runJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			// Create a string slice for JSON
			ids := make([]string, 0, len(list))
			for _, item := range list {
				ids = append(ids, item.ID)
			}
			return encoder.Encode(map[string]interface{}{"skills": ids})
		}

		for _, s := range list {
			fmt.Println(s.ID)
		}
		return nil
	},
}

var runAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := setupRunner(cmd.Context())
		if err != nil {
			return err
		}
		return r.RunAll(cmd.Context())
	},
}

var runResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume from last failure",
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := setupRunner(cmd.Context())
		if err != nil {
			return err
		}
		return r.Resume(cmd.Context())
	},
}

var runResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear run state",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		store, err := resolveStateStore(wd)
		if err != nil {
			return err
		}
		return store.Reset()
	},
}

var runReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show last run status",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		store, err := resolveStateStore(wd)
		if err != nil {
			return err
		}
		last, err := store.ReadLastRun()
		if err != nil {
			return err
		}

		if runJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(last)
		}

		if last == nil {
			fmt.Println("No run state found.")
			return nil
		}

		fmt.Printf("Status: %s\n", last.Status)
		if len(last.Failed) > 0 {
			fmt.Println("Failed:")
			for _, f := range last.Failed {
				fmt.Printf("  - %s\n", f)
			}
		} else {
			fmt.Println("All passed.")
		}
		return nil
	},
}

func runSkill(ctx context.Context, skillIDs []string) error {
	r, err := setupRunner(ctx)
	if err != nil {
		return err
	}

	// Verify skills exist first
	// Runner.RunList handles it
	return r.RunList(ctx, skillIDs)
}
