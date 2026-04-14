package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/config"
	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/module"
	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/orchestrator"
	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/phase"
)

func main() {
	root := &cobra.Command{
		Use:   "mantle-test",
		Short: "Mantle test orchestrator — cross-repo, multi-environment test runner",
	}

	root.AddCommand(runCmd())
	root.AddCommand(planCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runCmd() *cobra.Command {
	var (
		configPath string
		envType    string
		phases     []string
		modules    []string
		parallel   int
		failFast   bool
		outputJSON string
		setFlags   []string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute test suites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			// Apply CLI overrides
			overrides := parseSetFlags(setFlags)
			if err := config.MergeOverrides(cfg, overrides); err != nil {
				return err
			}

			if envType != "" {
				cfg.Environment.Type = module.EnvironmentType(envType)
			}
			if len(phases) > 0 {
				cfg.Execution.Phases = toPhases(phases)
			}
			if len(modules) > 0 {
				cfg.Modules.Filter = modules
			}
			if parallel > 0 {
				cfg.Execution.Parallel = parallel
			}
			if failFast {
				cfg.Execution.FailFast = true
			}

			engine, err := orchestrator.New(cfg)
			if err != nil {
				return err
			}

			report, err := engine.Run(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

			// Write JSON report if requested
			if outputJSON != "" && report != nil {
				data, _ := json.MarshalIndent(report, "", "  ")
				if writeErr := os.WriteFile(outputJSON, data, 0644); writeErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to write JSON report: %v\n", writeErr)
				} else {
					fmt.Printf("JSON report written to %s\n", outputJSON)
				}
			}

			if report != nil && report.Summary.FailedSuites > 0 {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "configs/local.yaml", "config file path")
	cmd.Flags().StringVar(&envType, "env", "", "environment type (unit, localchain, qa, mainnet)")
	cmd.Flags().StringSliceVar(&phases, "phase", nil, "phases to run (unit, integration, e2e, acceptance)")
	cmd.Flags().StringSliceVar(&modules, "modules", nil, "modules to include")
	cmd.Flags().IntVar(&parallel, "parallel", 0, "max concurrent suites per phase")
	cmd.Flags().BoolVar(&failFast, "fail-fast", false, "stop on first failure")
	cmd.Flags().StringVar(&outputJSON, "output-json", "", "write JSON report to file")
	cmd.Flags().StringSliceVar(&setFlags, "set", nil, "config overrides (key=value)")

	return cmd
}

func planCmd() *cobra.Command {
	var (
		configPath string
		envType    string
		phases     []string
		modules    []string
	)

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show execution plan without running tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			if envType != "" {
				cfg.Environment.Type = module.EnvironmentType(envType)
			}
			if len(phases) > 0 {
				cfg.Execution.Phases = toPhases(phases)
			}

			reg := module.NewRegistry()
			if err := reg.LoadDir(cfg.Modules.Dir); err != nil {
				return err
			}

			mods := reg.Filter(modules)
			p, err := phase.BuildPlan(mods, cfg.Execution.Phases, cfg.Environment.Type)
			if err != nil {
				return err
			}

			fmt.Printf("Execution Plan (%s environment)\n", cfg.Environment.Type)
			fmt.Printf("Total: %d suites\n\n", p.TotalRuns())
			for _, pp := range p.Phases {
				fmt.Printf("Phase: %s (%d suites)\n", pp.Phase, len(pp.Runs))
				for _, run := range pp.Runs {
					deps := ""
					if len(run.DependsOn) > 0 {
						deps = fmt.Sprintf(" (depends on: %s)", strings.Join(run.DependsOn, ", "))
					}
					fmt.Printf("  - %s [%s]%s\n", run.ID, run.Suite.Command, deps)
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "configs/local.yaml", "config file path")
	cmd.Flags().StringVar(&envType, "env", "", "environment type")
	cmd.Flags().StringSliceVar(&phases, "phase", nil, "phases to run")
	cmd.Flags().StringSliceVar(&modules, "modules", nil, "modules to include")

	return cmd
}

func parseSetFlags(flags []string) map[string]string {
	result := make(map[string]string)
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func toPhases(names []string) []module.Phase {
	phases := make([]module.Phase, len(names))
	for i, name := range names {
		phases[i] = module.Phase(name)
	}
	return phases
}
