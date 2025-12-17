package runner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSkill implements Skill for testing.
type MockSkill struct {
	id     string
	result SkillResult
	called bool
}

func (m *MockSkill) ID() string {
	return m.id
}

func (m *MockSkill) Run(ctx context.Context, deps *Deps) SkillResult {
	m.called = true
	return m.result
}

func TestRunner_RunAll(t *testing.T) {
	dir := t.TempDir()
	store := NewStateStore(dir)

	s1 := &MockSkill{id: "s1", result: SkillResult{Skill: "s1", Status: StatusPass}}
	s2 := &MockSkill{id: "s2", result: SkillResult{Skill: "s2", Status: StatusPass}}

	r := NewRunner([]Skill{s1, s2}, store, &Deps{})

	err := r.RunAll(context.Background())
	require.NoError(t, err)

	assert.True(t, s1.called)
	assert.True(t, s2.called)

	// Verify state
	last, err := store.ReadLastRun()
	require.NoError(t, err)
	assert.Equal(t, "pass", last.Status)
	assert.Equal(t, []string{"s1", "s2"}, last.Skills)
	assert.Empty(t, last.Failed)
}

func TestRunner_RunAll_Failure(t *testing.T) {
	dir := t.TempDir()
	store := NewStateStore(dir)

	s1 := &MockSkill{id: "s1", result: SkillResult{Skill: "s1", Status: StatusFail, ExitCode: 1}}
	s2 := &MockSkill{id: "s2", result: SkillResult{Skill: "s2", Status: StatusPass}}

	r := NewRunner([]Skill{s1, s2}, store, &Deps{})

	err := r.RunAll(context.Background())
	require.Error(t, err) // Should return error on failure

	assert.True(t, s1.called)
	assert.True(t, s2.called) // Should continue despite failure (soft fail for "all")

	// Verify state
	last, err := store.ReadLastRun()
	require.NoError(t, err)
	assert.Equal(t, "fail", last.Status)
	assert.Equal(t, []string{"s1"}, last.Failed)
}

func TestRunner_Resume(t *testing.T) {
	dir := t.TempDir()
	store := NewStateStore(dir)

	// Seed failure
	initialState := LastRun{
		Status: "fail",
		Skills: []string{"s1", "s2"},
		Failed: []string{"s2"},
	}
	err := store.WriteLastRun(initialState)
	require.NoError(t, err)

	s1 := &MockSkill{id: "s1", result: SkillResult{Skill: "s1", Status: StatusPass}}
	s2 := &MockSkill{id: "s2", result: SkillResult{Skill: "s2", Status: StatusPass}} // it passes this time

	r := NewRunner([]Skill{s1, s2}, store, &Deps{})

	err = r.Resume(context.Background())
	require.NoError(t, err)

	assert.False(t, s1.called) // Should NOT run s1
	assert.True(t, s2.called)  // Should run s2

	// State should differ now?
	// Resume updates state for the *run*, effectively overwriting?
	// scripts/run.sh writes a NEW last-run.json for the resume run.
	// So only the resumed skills are in the "skills" list of the new run state.

	last, err := store.ReadLastRun()
	require.NoError(t, err)
	assert.Equal(t, "pass", last.Status)
	assert.Equal(t, []string{"s2"}, last.Skills)
}
