package phase

import (
	"fmt"
	"sort"

	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/module"
)

// Plan represents an ordered set of phases, each containing scheduled suite runs.
type Plan struct {
	Phases []PhasePlan
}

// PhasePlan groups all scheduled runs for a single phase.
type PhasePlan struct {
	Phase module.Phase
	Runs  []ScheduledRun
}

// ScheduledRun represents a single suite execution to be performed.
type ScheduledRun struct {
	ID        string
	Module    string
	Suite     *module.Suite
	DependsOn []string // IDs of runs that must complete first
}

// BuildPlan creates an execution plan from registered modules, filtered by phases and environment.
func BuildPlan(
	modules []*module.Descriptor,
	phases []module.Phase,
	env module.EnvironmentType,
) (*Plan, error) {
	// If no phases specified, use all in order
	if len(phases) == 0 {
		phases = module.PhaseOrder
	}

	// Build lookup of all qualified suite IDs for dependency resolution
	allRuns := make(map[string]*ScheduledRun)
	var plan Plan

	for _, p := range phases {
		pp := PhasePlan{Phase: p}

		for _, mod := range modules {
			for i := range mod.Suites {
				suite := &mod.Suites[i]
				if suite.Phase != p {
					continue
				}
				if !suite.SupportsEnv(env) {
					continue
				}

				id := suite.QualifiedName(mod.Name)
				run := ScheduledRun{
					ID:     id,
					Module: mod.Name,
					Suite:  suite,
				}

				// Resolve dependencies
				for _, dep := range suite.DependsOn {
					run.DependsOn = append(run.DependsOn, dep)
				}

				allRuns[id] = &run
				pp.Runs = append(pp.Runs, run)
			}
		}

		if len(pp.Runs) > 0 {
			plan.Phases = append(plan.Phases, pp)
		}
	}

	// Validate dependencies exist
	for _, pp := range plan.Phases {
		for _, run := range pp.Runs {
			for _, dep := range run.DependsOn {
				if _, ok := allRuns[dep]; !ok {
					return nil, fmt.Errorf("suite %s depends on %s, which is not in the plan", run.ID, dep)
				}
			}
		}
	}

	// Sort runs within each phase by dependency order
	for i := range plan.Phases {
		sortByDeps(plan.Phases[i].Runs)
	}

	return &plan, nil
}

// sortByDeps performs a topological sort of runs based on DependsOn.
func sortByDeps(runs []ScheduledRun) {
	idxMap := make(map[string]int)
	for i, r := range runs {
		idxMap[r.ID] = i
	}

	sort.SliceStable(runs, func(i, j int) bool {
		// If j depends on i, i should come first
		for _, dep := range runs[j].DependsOn {
			if dep == runs[i].ID {
				return true
			}
		}
		return false
	})
}

// TotalRuns returns the total number of suite runs across all phases.
func (p *Plan) TotalRuns() int {
	total := 0
	for _, pp := range p.Phases {
		total += len(pp.Runs)
	}
	return total
}
