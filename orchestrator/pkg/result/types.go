package result

import "time"

// SuiteResult holds the outcome of running a single test suite.
type SuiteResult struct {
	Module   string        `json:"module"`
	Suite    string        `json:"suite"`
	Phase    string        `json:"phase"`
	Env      string        `json:"environment"`
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
	Duration  time.Duration    `json:"duration"`
	ExitCode  int              `json:"exit_code"`
	Failures  []TestFailure    `json:"failures,omitempty"`
	Artifacts map[string]string `json:"artifacts,omitempty"` // name → file path of native report artifacts
}

// TestFailure describes a single failed test case.
type TestFailure struct {
	TestID  string `json:"test_id"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

// RunReport aggregates all suite results for a single orchestrator run.
type RunReport struct {
	RunID       string        `json:"run_id"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Environment string        `json:"environment"`
	Suites      []SuiteResult `json:"suites"`
	Summary Summary `json:"summary"`
}

// Summary provides top-level counts across all suites.
type Summary struct {
	TotalSuites  int `json:"total_suites"`
	PassedSuites int `json:"passed_suites"`
	FailedSuites int `json:"failed_suites"`
	TotalTests   int `json:"total_tests"`
	PassedTests  int `json:"passed_tests"`
	FailedTests  int `json:"failed_tests"`
	SkippedTests int `json:"skipped_tests"`
}
