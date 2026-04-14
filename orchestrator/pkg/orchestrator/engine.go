package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/adapter"
	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/config"
	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/environment"
	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/module"
	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/phase"
	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/result"
	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/result/parser"
)

// Engine is the top-level orchestrator that wires config, modules, environments, and execution.
type Engine struct {
	cfg      *config.Config
	registry *module.Registry
	env      environment.Environment
}

// New creates a new Engine from a loaded config.
func New(cfg *config.Config) (*Engine, error) {
	// Load module manifests
	reg := module.NewRegistry()
	if err := reg.LoadDir(cfg.Modules.Dir); err != nil {
		return nil, fmt.Errorf("loading modules: %w", err)
	}

	// Create environment
	env, err := environment.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating environment: %w", err)
	}

	return &Engine{
		cfg:      cfg,
		registry: reg,
		env:      env,
	}, nil
}

// RunOptions holds CLI-provided options that override config.
type RunOptions struct {
	ConfigPath string
	Overrides  map[string]string
}

// Run executes the full test pipeline: setup → plan → execute → compare → report → teardown.
func (e *Engine) Run(ctx context.Context) (*result.RunReport, error) {
	report := &result.RunReport{
		RunID:       fmt.Sprintf("run-%d", time.Now().Unix()),
		StartedAt:   time.Now(),
		Environment: string(e.cfg.Environment.Type),
	}

	// 1. Setup environment
	fmt.Printf("Setting up %s environment...\n", e.cfg.Environment.Type)
	if err := e.env.Setup(ctx); err != nil {
		return nil, fmt.Errorf("environment setup: %w", err)
	}
	defer func() {
		fmt.Println("Tearing down environment...")
		e.env.Teardown(ctx)
	}()

	// 2. Build execution plan
	modules := e.registry.Filter(e.cfg.Modules.Filter)
	if len(modules) == 0 {
		return nil, fmt.Errorf("no modules found matching filter %v", e.cfg.Modules.Filter)
	}

	plan, err := phase.BuildPlan(modules, e.cfg.Execution.Phases, e.cfg.Environment.Type)
	if err != nil {
		return nil, fmt.Errorf("building plan: %w", err)
	}

	fmt.Printf("Plan: %d suites across %d phases\n", plan.TotalRuns(), len(plan.Phases))
	for _, pp := range plan.Phases {
		fmt.Printf("  Phase %s: %d suites\n", pp.Phase, len(pp.Runs))
		for _, run := range pp.Runs {
			fmt.Printf("    - %s\n", run.ID)
		}
	}

	// 3. Execute
	runner := e.buildRunner(modules)
	scheduler := &phase.Scheduler{
		Parallel: e.cfg.Execution.Parallel,
		FailFast: e.cfg.Execution.FailFast,
	}

	results, err := scheduler.Execute(ctx, plan, runner)
	report.Suites = results
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
	}

	// 4. Build summary
	report.CompletedAt = time.Now()
	report.Summary = buildSummary(results)

	// 5. Print console report
	printConsoleReport(report)

	// 6. Save report to reports directory
	if err := e.saveReport(report); err != nil {
		fmt.Printf("Warning: failed to save report: %v\n", err)
	}

	return report, nil
}

// saveReport collects each module's native report artifacts into a timestamped reports directory.
// It does NOT regenerate reports — it copies the original files produced by each module.
func (e *Engine) saveReport(report *result.RunReport) error {
	reportsDir := e.cfg.Execution.ReportsDir
	if reportsDir == "" {
		reportsDir = "reports/"
	}

	timestamp := report.StartedAt.Format("20060102-150405")
	runDir := filepath.Join(reportsDir, timestamp)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return fmt.Errorf("creating reports dir: %w", err)
	}

	// Collect each suite's native report artifact
	for _, sr := range report.Suites {
		if sr.Artifacts == nil {
			continue
		}
		moduleDir := filepath.Join(runDir, sr.Module)
		if err := os.MkdirAll(moduleDir, 0755); err != nil {
			continue
		}
		for name, srcPath := range sr.Artifacts {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				fmt.Printf("  Warning: could not read artifact %s: %v\n", srcPath, err)
				continue
			}
			dstPath := filepath.Join(moduleDir, name)
			os.WriteFile(dstPath, data, 0644)
		}
	}

	// Also collect EEST logs if mantle-execution-specs is available
	specsDir := os.Getenv("SPECS_DIR")
	if specsDir == "" {
		specsDir = filepath.Join("..", "mantle-execution-specs")
	}
	eestLogsDir := filepath.Join(specsDir, "logs")
	if info, err := os.Stat(eestLogsDir); err == nil && info.IsDir() {
		eestReportDir := filepath.Join(runDir, "evm-conformance", "eest-logs")
		os.MkdirAll(eestReportDir, 0755)
		entries, _ := os.ReadDir(eestLogsDir)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			src := filepath.Join(eestLogsDir, entry.Name())
			dst := filepath.Join(eestReportDir, entry.Name())
			data, err := os.ReadFile(src)
			if err == nil {
				os.WriteFile(dst, data, 0644)
			}
		}
	}

	// Collect EEST HTML report if exists
	eestHTML := filepath.Join(specsDir, "execution_results", "report_execute.html")
	if data, err := os.ReadFile(eestHTML); err == nil {
		eestDir := filepath.Join(runDir, "evm-conformance")
		os.MkdirAll(eestDir, 0755)
		os.WriteFile(filepath.Join(eestDir, "report_execute.html"), data, 0644)
	}

	fmt.Printf("\nReports collected to %s/\n", runDir)

	// List what was collected
	filepath.Walk(runDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(runDir, path)
		fmt.Printf("  %s (%d bytes)\n", rel, info.Size())
		return nil
	})

	return nil
}

func (e *Engine) buildRunner(modules []*module.Descriptor) phase.SuiteRunner {
	// Build module lookup
	modMap := make(map[string]*module.Descriptor)
	for _, m := range modules {
		modMap[m.Name] = m
	}

	execRunner := &adapter.ExecRunner{
		WorkDir: ".", // TODO: configurable
		BaseEnv: e.env.EnvVars(),
	}

	return func(ctx context.Context, run phase.ScheduledRun) (*result.SuiteResult, error) {
		mod, ok := modMap[run.Module]
		if !ok {
			return nil, fmt.Errorf("module %s not found", run.Module)
		}

		p, err := parser.ForFormat(run.Suite.ResultFormat)
		if err != nil {
			return nil, err
		}

		fmt.Printf("  Running %s...\n", run.ID)
		sr, err := execRunner.RunSuite(ctx, mod, run.Suite, p)
		if err != nil {
			return nil, err
		}

		status := "PASS"
		if sr.Failed > 0 || sr.ExitCode != 0 {
			status = "FAIL"
		}
		fmt.Printf("  %s %s (%d passed, %d failed, %d skipped) [%s]\n",
			status, run.ID, sr.Passed, sr.Failed, sr.Skipped, sr.Duration.Round(time.Second))

		return sr, nil
	}
}

func buildSummary(results []result.SuiteResult) result.Summary {
	var s result.Summary
	s.TotalSuites = len(results)
	for _, r := range results {
		if r.Failed == 0 && r.ExitCode == 0 {
			s.PassedSuites++
		} else {
			s.FailedSuites++
		}
		s.TotalTests += r.Total
		s.PassedTests += r.Passed
		s.FailedTests += r.Failed
		s.SkippedTests += r.Skipped
	}
	return s
}

func printConsoleReport(report *result.RunReport) {
	fmt.Println("\n" + "=== Run Report ===")
	fmt.Printf("Run ID:      %s\n", report.RunID)
	fmt.Printf("Environment: %s\n", report.Environment)
	fmt.Printf("Duration:    %s\n", report.CompletedAt.Sub(report.StartedAt).Round(time.Second))
	fmt.Println()
	fmt.Printf("Suites:  %d total, %d passed, %d failed\n",
		report.Summary.TotalSuites, report.Summary.PassedSuites, report.Summary.FailedSuites)
	fmt.Printf("Tests:   %d total, %d passed, %d failed, %d skipped\n",
		report.Summary.TotalTests, report.Summary.PassedTests, report.Summary.FailedTests, report.Summary.SkippedTests)

	if report.Summary.FailedSuites > 0 {
		fmt.Println("\nFailed suites:")
		for _, sr := range report.Suites {
			if sr.Failed > 0 || sr.ExitCode != 0 {
				fmt.Printf("  - %s:%s (%d failures)\n", sr.Module, sr.Suite, sr.Failed)
				for _, f := range sr.Failures {
					fmt.Printf("      %s: %s\n", f.TestID, f.Message)
				}
			}
		}
	}
}
