package runner

// SkillStatus represents the outcome of a skill execution.
type SkillStatus string

const (
	StatusPass SkillStatus = "pass"
	StatusFail SkillStatus = "fail"
	StatusSkip SkillStatus = "skip"
)

// SkillResult represents the result of a single skill execution.
// Matches .cortex/run/skills/<skill>.json schema.
type SkillResult struct {
	Skill    string      `json:"skill"`
	Status   SkillStatus `json:"status"`
	ExitCode int         `json:"exit_code"`
	Note     string      `json:"note,omitempty"`
}

// LastRun represents the summary of the last execution.
// Matches .cortex/run/last-run.json schema.
type LastRun struct {
	Status string   `json:"status"` // "pass" or "fail"
	Skills []string `json:"skills"` // Ordered list of skills run
	Failed []string `json:"failed"` // List of failed skills
}
