package adapter

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/module"
	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/result"
)

// ExecRunner runs module test suites as shell commands.
type ExecRunner struct {
	// WorkDir is the base directory where module repos are checked out.
	WorkDir string
	// BaseEnv holds environment variables injected by the orchestrator (RPC URL, chain ID, etc).
	BaseEnv map[string]string
}

// RunResult wraps the raw output and exit code from a command execution.
type RunResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Duration time.Duration
}

// Run executes a module suite's command and returns the raw output.
func (r *ExecRunner) Run(ctx context.Context, mod *module.Descriptor, suite *module.Suite) (*RunResult, error) {
	// Build the command
	cmd := exec.CommandContext(ctx, "sh", "-c", suite.Command)
	cmd.Dir = r.moduleDir(mod.Name)

	// Build env: inherit current env + orchestrator base env + suite-specific vars
	cmd.Env = r.buildEnv(suite.EnvVars)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Apply timeout from suite config
	if suite.Timeout.Duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, suite.Timeout.Duration)
		defer cancel()
		cmd.Cancel = func() error {
			return cmd.Process.Kill()
		}
	}

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("executing %s:%s: %w", mod.Name, suite.Name, err)
		}
	}

	return &RunResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// RunSuite executes a suite and parses the result into a SuiteResult.
func (r *ExecRunner) RunSuite(ctx context.Context, mod *module.Descriptor, suite *module.Suite, parser result.Parser) (*result.SuiteResult, error) {
	raw, err := r.Run(ctx, mod, suite)
	if err != nil {
		return nil, err
	}

	// Try to parse structured results
	var sr *result.SuiteResult
	if suite.ResultPath != "" {
		// Read result from file
		resultFile := suite.ResultPath
		if !isAbsPath(resultFile) {
			resultFile = r.moduleDir(mod.Name) + "/" + resultFile
		}
		data, readErr := os.ReadFile(resultFile)
		if readErr == nil {
			sr, err = parser.Parse(data)
		}
	}

	// Fall back to parsing stdout
	if sr == nil {
		sr, err = parser.Parse(raw.Stdout)
		if err != nil {
			// If parsing fails, create a minimal result from exit code
			sr = &result.SuiteResult{
				Total:  1,
				Passed: boolToInt(raw.ExitCode == 0),
				Failed: boolToInt(raw.ExitCode != 0),
			}
			if raw.ExitCode != 0 {
				sr.Failures = []result.TestFailure{{
					TestID:  suite.Name,
					Message: fmt.Sprintf("exit code %d", raw.ExitCode),
					Output:  string(raw.Stderr),
				}}
			}
		}
	}

	sr.Module = mod.Name
	sr.Suite = suite.Name
	sr.Phase = string(suite.Phase)
	sr.Duration = raw.Duration
	sr.ExitCode = raw.ExitCode

	return sr, nil
}

func (r *ExecRunner) moduleDir(moduleName string) string {
	return r.WorkDir + "/" + moduleName
}

func (r *ExecRunner) buildEnv(suiteVars []string) []string {
	// Start with current process environment
	env := os.Environ()
	// Add orchestrator base env vars
	for k, v := range r.BaseEnv {
		env = append(env, k+"="+v)
	}
	return env
}

func isAbsPath(path string) bool {
	return len(path) > 0 && path[0] == '/'
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
