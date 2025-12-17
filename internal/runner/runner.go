package runner

import (
	"context"
	"fmt"
	"time"
)

// Runner manages the execution of skills.
type Runner struct {
	skills []Skill
	store  *StateStore
	deps   *Deps
}

// NewRunner creates a new runner with the given skills and dependencies.
func NewRunner(skills []Skill, store *StateStore, deps *Deps) *Runner {
	return &Runner{
		skills: skills,
		store:  store,
		deps:   deps,
	}
}

// RunAll executes all skills in order.
// It continues execution even if a skill fails, accumulating failures.
// Returns an error if ANY skill failed.
func (r *Runner) RunAll(ctx context.Context) error {
	return r.executeSequence(ctx, r.skills)
}

// Resume continues execution from the first failed skill in the last run.
func (r *Runner) Resume(ctx context.Context) error {
	failed, err := r.store.LoadFailedSkills()
	if err != nil {
		return fmt.Errorf("loading failed skills: %w", err)
	}

	if len(failed) == 0 {
		// Nothing failed, do nothing or run all?
		// Scripts/run.sh says "No failed skills to resume" and exits 0.
		// We'll mimic that by returning nil.
		return nil
	}

	// Logic from scripts/run.sh:
	// It basically re-runs the failed skills (and arguably subsequent ones, but the script logic
	// actually just re-runs the failed ones if I recall correctly, OR it re-runs from the failure point).
	// Let's re-read the script logic in my head:
	// "while IFS= read -r skill; do ... run_skill ... done << failed_list"
	// It only re-runs the SPECIFIC failed skills.

	// Wait, the Review said "Resume logic: find first failed skill from last run; re-run from there."
	// But the script says: "failed_list=...; while read skill; do run_skill...".
	// The script logic is "Retry the ones that failed".
	// However, if a skill blocked subsequent skills from running, they wouldn't be in "failed", they just wouldn't be in the list?
	// Actually scripts/run.sh stops on first fail. So "failed" contains the ONE skill that failed (and maybe others if it kept going, but it has `set -e` behavior sort of).
	// Actually `run_skill || { overall_rc=1; failed+=("$skill"); }`
	// And the loop continues?
	// "run_skill ... || { ... }" implies it continues?
	// Let's check run.sh content again...
	// Line 218: "run_skill "$skill" || { overall_rc=1; failed+=("$skill"); }"
	// It continues! `run.sh all` runs EVERYTHING and collects failures.
	// BUT `run_skill` line 138 `return 2`? No.
	// `run_skill` returns the rc.
	// The loop in `all` continues.

	// So `resume` runs ONLY the skills that failed previously.
	// The script does NOT run subsequent skipped skills because `all` runs everything anyway (it doesn't abort early).
	// WAIT. `run.sh` line 17 check: `set -euo pipefail`.
	// Line 218 loop. `||` catches the error so `set -e` doesn't trigger.

	// So `cortex run all` should run ALL skills and collect failures.
	// And `cortex run resume` should run ONLY the failed skills.

	// I will implement "Run specific list of skills".

	toRun := []Skill{}
	for _, id := range failed {
		skill := r.findSkill(id)
		if skill != nil {
			toRun = append(toRun, skill)
		}
	}

	return r.executeSequence(ctx, toRun)
}

// RunList executes a specific list of skill IDs.
func (r *Runner) RunList(ctx context.Context, skillIDs []string) error {
	var toRun []Skill
	for _, id := range skillIDs {
		s := r.findSkill(id)
		if s == nil {
			return fmt.Errorf("skill not found: %s", id)
		}
		toRun = append(toRun, s)
	}
	return r.executeSequence(ctx, toRun)
}

func (r *Runner) findSkill(id string) Skill {
	for _, s := range r.skills {
		if s.ID() == id {
			return s
		}
	}
	return nil
}

// executeSequence runs a sequence of skills, updating state.
// It returns error if ANY skill failed.
func (r *Runner) executeSequence(ctx context.Context, skills []Skill) error {
	var failed []string
	var skillNames []string

	overallSuccess := true

	for _, skill := range skills {
		id := skill.ID()
		skillNames = append(skillNames, id)

		fmt.Println("")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("SKILL: %s\n", id)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("")

		// Measure duration?
		start := time.Now()
		_ = start

		res := skill.Run(ctx, r.deps)

		// Save individual result
		if err := r.store.WriteSkillResult(res); err != nil {
			return fmt.Errorf("writing result for %s: %w", id, err)
		}

		if res.Status == StatusSkip {
			fmt.Printf("SKIP: %s\n", id)
			if res.Note != "" {
				fmt.Println(res.Note)
			}
			continue
		}

		if res.Status != StatusPass {
			failed = append(failed, id)
			overallSuccess = false
			fmt.Printf("FAIL: %s (exit %d)\n", id, res.ExitCode)
			if res.Note != "" {
				fmt.Println(res.Note)
			}
		} else {
			// passed = append(passed, id)
			fmt.Printf("PASS: %s\n", id)
			if res.Note != "" {
				fmt.Println(res.Note)
			}
		}
	}

	// Update last run
	lastRun := LastRun{
		Status: "pass",
		Skills: skillNames,
		Failed: failed,
	}
	if !overallSuccess {
		lastRun.Status = "fail"
	}

	if err := r.store.WriteLastRun(lastRun); err != nil {
		return fmt.Errorf("writing last run: %w", err)
	}

	if !overallSuccess {
		return fmt.Errorf("run failed: %v", failed)
	}
	return nil
}
