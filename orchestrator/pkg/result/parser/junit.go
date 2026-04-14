package parser

import (
	"encoding/xml"

	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/result"
)

// JUnitParser parses JUnit XML test reports.
type JUnitParser struct{}

type junitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	XMLName  xml.Name        `xml:"testsuite"`
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Skipped  int             `xml:"skipped,attr"`
	Cases    []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Failure   *junitFailure `xml:"failure"`
	Error     *junitFailure `xml:"error"`
	Skipped   *struct{}     `xml:"skipped"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

func (p *JUnitParser) Parse(data []byte) (*result.SuiteResult, error) {
	// Try parsing as <testsuites> first
	var suites junitTestSuites
	if err := xml.Unmarshal(data, &suites); err != nil {
		// Try as a single <testsuite>
		var single junitTestSuite
		if err2 := xml.Unmarshal(data, &single); err2 != nil {
			return nil, err2
		}
		suites.Suites = []junitTestSuite{single}
	}

	sr := &result.SuiteResult{}

	for _, suite := range suites.Suites {
		for _, tc := range suite.Cases {
			sr.Total++
			switch {
			case tc.Failure != nil:
				sr.Failed++
				sr.Failures = append(sr.Failures, result.TestFailure{
					TestID:  tc.ClassName + "/" + tc.Name,
					Message: tc.Failure.Message,
					Output:  tc.Failure.Body,
				})
			case tc.Error != nil:
				sr.Failed++
				sr.Failures = append(sr.Failures, result.TestFailure{
					TestID:  tc.ClassName + "/" + tc.Name,
					Message: tc.Error.Message,
					Output:  tc.Error.Body,
				})
			case tc.Skipped != nil:
				sr.Skipped++
			default:
				sr.Passed++
			}
		}
	}

	return sr, nil
}
