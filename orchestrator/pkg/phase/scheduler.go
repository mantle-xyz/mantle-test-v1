package phase

import (
	"context"
	"fmt"
	"sync"

	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/result"
)

// SuiteRunner is a function that executes a single scheduled run and returns its result.
type SuiteRunner func(ctx context.Context, run ScheduledRun) (*result.SuiteResult, error)

// Scheduler executes a Plan phase-by-phase with parallelism within each phase.
type Scheduler struct {
	Parallel int  // max concurrent runs per phase
	FailFast bool // stop on first failure within a phase
}

// Execute runs the entire plan and returns all suite results.
func (s *Scheduler) Execute(ctx context.Context, plan *Plan, runner SuiteRunner) ([]result.SuiteResult, error) {
	var allResults []result.SuiteResult

	for _, pp := range plan.Phases {
		phaseResults, err := s.executePhase(ctx, pp, runner)
		allResults = append(allResults, phaseResults...)
		if err != nil {
			return allResults, fmt.Errorf("phase %s: %w", pp.Phase, err)
		}

		// Check if any suite in this phase failed
		if s.FailFast {
			for _, sr := range phaseResults {
				if sr.Failed > 0 || sr.ExitCode != 0 {
					return allResults, fmt.Errorf("phase %s: suite %s:%s failed, stopping (fail-fast)", pp.Phase, sr.Module, sr.Suite)
				}
			}
		}
	}

	return allResults, nil
}

func (s *Scheduler) executePhase(ctx context.Context, pp PhasePlan, runner SuiteRunner) ([]result.SuiteResult, error) {
	parallel := s.Parallel
	if parallel <= 0 {
		parallel = 4
	}

	// Group runs: those without deps run in parallel, those with deps run after their deps complete
	completed := make(map[string]bool)
	var mu sync.Mutex
	var results []result.SuiteResult

	// Simple approach: run non-dependent suites first in parallel, then dependent ones
	var independent, dependent []ScheduledRun
	for _, run := range pp.Runs {
		if len(run.DependsOn) == 0 {
			independent = append(independent, run)
		} else {
			dependent = append(dependent, run)
		}
	}

	// Run independent suites in parallel
	indResults, err := s.runParallel(ctx, independent, runner, parallel)
	results = append(results, indResults...)
	if err != nil {
		return results, err
	}

	mu.Lock()
	for _, r := range independent {
		completed[r.ID] = true
	}
	mu.Unlock()

	// Run dependent suites (sequentially for simplicity; could optimize with DAG scheduler)
	for _, run := range dependent {
		// Check deps are met
		for _, dep := range run.DependsOn {
			if !completed[dep] {
				return results, fmt.Errorf("dependency %s not completed for %s", dep, run.ID)
			}
		}

		sr, err := runner(ctx, run)
		if err != nil {
			return results, fmt.Errorf("running %s: %w", run.ID, err)
		}
		results = append(results, *sr)
		completed[run.ID] = true
	}

	return results, nil
}

func (s *Scheduler) runParallel(ctx context.Context, runs []ScheduledRun, runner SuiteRunner, parallel int) ([]result.SuiteResult, error) {
	if len(runs) == 0 {
		return nil, nil
	}

	sem := make(chan struct{}, parallel)
	var mu sync.Mutex
	var results []result.SuiteResult
	var firstErr error

	var wg sync.WaitGroup
	for _, run := range runs {
		wg.Add(1)
		go func(r ScheduledRun) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			sr, err := runner(ctx, r)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			results = append(results, *sr)
		}(run)
	}

	wg.Wait()
	return results, firstErr
}
