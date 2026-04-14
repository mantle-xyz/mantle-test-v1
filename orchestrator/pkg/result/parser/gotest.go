package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/result"
)

// GoTestParser parses `go test -json` output.
type GoTestParser struct{}

// goTestEvent represents a single JSON line from `go test -json`.
type goTestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

func (p *GoTestParser) Parse(data []byte) (*result.SuiteResult, error) {
	sr := &result.SuiteResult{}
	outputs := make(map[string][]string) // test name → output lines

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var event goTestEvent
		if err := json.Unmarshal(line, &event); err != nil {
			continue // skip non-JSON lines
		}

		if event.Test == "" {
			continue // package-level event
		}

		switch event.Action {
		case "pass":
			sr.Passed++
			sr.Total++
		case "fail":
			sr.Failed++
			sr.Total++
			sr.Failures = append(sr.Failures, result.TestFailure{
				TestID:  event.Test,
				Message: "FAIL",
				Output:  strings.Join(outputs[event.Test], ""),
			})
		case "skip":
			sr.Skipped++
			sr.Total++
		case "output":
			outputs[event.Test] = append(outputs[event.Test], event.Output)
		}
	}

	return sr, nil
}
