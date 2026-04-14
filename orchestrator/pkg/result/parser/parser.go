package parser

import (
	"fmt"

	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/module"
	"github.com/mantlenetworkio/mantle-test/orchestrator/pkg/result"
)

// ForFormat returns the appropriate parser for a result format.
func ForFormat(format module.ResultFormat) (result.Parser, error) {
	switch format {
	case module.ResultGoTestJSON:
		return &GoTestParser{}, nil
	case module.ResultJUnitXML:
		return &JUnitParser{}, nil
	case module.ResultEESTJSON:
		return &EESTParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported result format: %s", format)
	}
}
