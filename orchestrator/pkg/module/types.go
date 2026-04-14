package module

// Phase represents a testing phase in the execution pipeline.
// Phases execute in order: unit → integration → e2e → acceptance.
type Phase string

const (
	PhaseUnit        Phase = "unit"
	PhaseIntegration Phase = "integration"
	PhaseE2E         Phase = "e2e"
	PhaseAcceptance  Phase = "acceptance"
)

// PhaseOrder defines the execution order of phases.
var PhaseOrder = []Phase{PhaseUnit, PhaseIntegration, PhaseE2E, PhaseAcceptance}

// PhaseIndex returns the position of a phase in the execution order, or -1 if unknown.
func PhaseIndex(p Phase) int {
	for i, phase := range PhaseOrder {
		if phase == p {
			return i
		}
	}
	return -1
}

// EnvironmentType represents a test execution environment.
type EnvironmentType string

const (
	EnvUnit       EnvironmentType = "unit"
	EnvLocalChain EnvironmentType = "localchain"
	EnvQA         EnvironmentType = "qa"
	EnvMainnet    EnvironmentType = "mainnet"
)

// ResultFormat describes the output format a module's test suite produces.
type ResultFormat string

const (
	ResultGoTestJSON ResultFormat = "gotest-json"
	ResultJUnitXML   ResultFormat = "junit-xml"
	ResultEESTJSON   ResultFormat = "eest-json"
)
