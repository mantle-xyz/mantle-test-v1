package parser

import (
	"encoding/json"

	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/result"
)

// EESTParser parses EEST (ethereum/execution-spec-tests) JSON output.
type EESTParser struct{}

// eestResult represents the top-level EEST output structure.
type eestResult struct {
	Tests []eestTest `json:"tests"`
}

type eestTest struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "passed", "failed", "skipped"
	Error  string `json:"error,omitempty"`
}

func (p *EESTParser) Parse(data []byte) (*result.SuiteResult, error) {
	var eest eestResult
	if err := json.Unmarshal(data, &eest); err != nil {
		// EEST may output line-by-line JSON; try line-by-line parsing
		return p.parseLines(data)
	}
	return p.fromTests(eest.Tests), nil
}

func (p *EESTParser) parseLines(data []byte) (*result.SuiteResult, error) {
	// Fallback: try to parse as a flat array of test objects
	var tests []eestTest
	if err := json.Unmarshal(data, &tests); err != nil {
		return nil, err
	}
	return p.fromTests(tests), nil
}

func (p *EESTParser) fromTests(tests []eestTest) *result.SuiteResult {
	sr := &result.SuiteResult{Total: len(tests)}
	for _, t := range tests {
		switch t.Status {
		case "passed":
			sr.Passed++
		case "failed":
			sr.Failed++
			sr.Failures = append(sr.Failures, result.TestFailure{
				TestID:  t.Name,
				Message: t.Error,
			})
		case "skipped":
			sr.Skipped++
		default:
			sr.Skipped++
		}
	}
	return sr
}
