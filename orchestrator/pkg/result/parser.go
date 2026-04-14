package result

// Parser converts raw test output bytes into a structured SuiteResult.
type Parser interface {
	Parse(data []byte) (*SuiteResult, error)
}
