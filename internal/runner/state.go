package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// StateStore handles reading and writing runner state.
type StateStore struct {
	baseDir string
}

// NewStateStore creates a store at the given base directory (e.g. .cortex/run).
func NewStateStore(baseDir string) *StateStore {
	return &StateStore{baseDir: baseDir}
}

func (s *StateStore) lastRunPath() string {
	return filepath.Join(s.baseDir, "last-run.json")
}

// ReadLastRun loads the last execution summary.
func (s *StateStore) ReadLastRun() (*LastRun, error) {
	path := s.lastRunPath()
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil // Not found is clean state
	}
	if err != nil {
		return nil, fmt.Errorf("opening last run file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var last LastRun
	if err := json.NewDecoder(f).Decode(&last); err != nil {
		return nil, fmt.Errorf("decoding last run: %w", err)
	}
	return &last, nil
}

func (s *StateStore) ReadSkill(skillID string) (*SkillResult, error) {
	path := filepath.Join(s.baseDir, "skills", skillID+".json")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var res SkillResult
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

// WriteLastRun saves the execution summary.
func (s *StateStore) WriteLastRun(last LastRun) (err error) {
	path := s.lastRunPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		cerr := f.Close()
		if err == nil {
			err = cerr
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(last)
}

// WriteSkillResult saves a skill's result.
func (s *StateStore) WriteSkillResult(res SkillResult) (err error) {
	path := filepath.Join(s.baseDir, "skills", res.Skill+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		cerr := f.Close()
		if err == nil {
			err = cerr
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// Reset clears the state directory.
func (s *StateStore) Reset() error {
	return os.RemoveAll(s.baseDir)
}

// LoadFailedSkills returns a list of skills that failed in the last run.
func (s *StateStore) LoadFailedSkills() ([]string, error) {
	last, err := s.ReadLastRun()
	if err != nil {
		return nil, err
	}
	if last == nil {
		return nil, nil
	}
	return last.Failed, nil
}
