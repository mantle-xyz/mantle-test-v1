package reporter

import (
	"fmt"
	"io"
	"time"

	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/result"
)

// ConsoleReporter writes a human-readable summary to an io.Writer.
type ConsoleReporter struct {
	Writer io.Writer
}

func (r *ConsoleReporter) Report(report *result.RunReport) error {
	w := r.Writer

	fmt.Fprintf(w, "\n=== Mantle Test Report ===\n")
	fmt.Fprintf(w, "Run:         %s\n", report.RunID)
	fmt.Fprintf(w, "Environment: %s\n", report.Environment)
	fmt.Fprintf(w, "Duration:    %s\n", report.CompletedAt.Sub(report.StartedAt).Round(time.Second))
	fmt.Fprintln(w)

	// Suite results
	for _, sr := range report.Suites {
		status := "PASS"
		if sr.Failed > 0 || sr.ExitCode != 0 {
			status = "FAIL"
		}
		fmt.Fprintf(w, "  [%s] %s:%s — %d passed, %d failed, %d skipped (%s)\n",
			status, sr.Module, sr.Suite, sr.Passed, sr.Failed, sr.Skipped,
			sr.Duration.Round(time.Millisecond))
	}

	fmt.Fprintln(w)

	// Summary
	s := report.Summary
	fmt.Fprintf(w, "Summary:\n")
	fmt.Fprintf(w, "  Suites: %d total, %d passed, %d failed\n",
		s.TotalSuites, s.PassedSuites, s.FailedSuites)
	fmt.Fprintf(w, "  Tests:  %d total, %d passed, %d failed, %d skipped\n",
		s.TotalTests, s.PassedTests, s.FailedTests, s.SkippedTests)

	return nil
}
