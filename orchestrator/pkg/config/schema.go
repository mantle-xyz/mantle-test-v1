package config

import (
	"fmt"

	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/module"
)

// Config represents the top-level orchestrator configuration loaded from YAML.
type Config struct {
	Environment EnvironmentConfig `yaml:"environment"`
	Modules     ModulesConfig     `yaml:"modules"`
	Execution   ExecutionConfig   `yaml:"execution"`
}

// EnvironmentConfig specifies which environment to target and its connection details.
type EnvironmentConfig struct {
	Type        module.EnvironmentType `yaml:"type"`
	L2RPCURL    string                 `yaml:"l2_rpc_url"`
	L2WSURL     string                 `yaml:"l2_ws_url,omitempty"`
	L1RPCURL    string                 `yaml:"l1_rpc_url,omitempty"`
	ChainID     uint64                 `yaml:"chain_id"`
	DeployerKey string                 `yaml:"deployer_key,omitempty"`
	SeedKey     string                 `yaml:"seed_key,omitempty"`

	// Localchain-specific: devnet management
	DevnetCompose string `yaml:"devnet_compose,omitempty"` // path to docker-compose.yml
}

// ModulesConfig controls which modules and suites to run.
type ModulesConfig struct {
	Dir    string   `yaml:"dir"`              // directory containing module manifests (default: "modules/")
	Filter []string `yaml:"filter,omitempty"` // module names to include (empty = all)
	Tags   []string `yaml:"tags,omitempty"`   // suite tags to filter by
}

// ExecutionConfig controls how tests are executed.
type ExecutionConfig struct {
	Phases    []module.Phase `yaml:"phases,omitempty"`     // phases to run (empty = all)
	Parallel  int            `yaml:"parallel,omitempty"`   // max concurrent suites per phase (default: 4)
	FailFast  bool           `yaml:"fail_fast,omitempty"`
	Output    []string       `yaml:"output,omitempty"`     // report formats: "console", "json", "junit"
	ReportsDir string        `yaml:"reports_dir,omitempty"` // directory for collected reports (default: "reports/")
}

// EnvVars returns the environment variables map that the orchestrator injects into module commands.
func (c *Config) EnvVars() map[string]string {
	vars := map[string]string{}
	if c.Environment.L2RPCURL != "" {
		vars["L2_RPC_URL"] = c.Environment.L2RPCURL
	}
	if c.Environment.L2WSURL != "" {
		vars["L2_WS_URL"] = c.Environment.L2WSURL
	}
	if c.Environment.L1RPCURL != "" {
		vars["L1_RPC_URL"] = c.Environment.L1RPCURL
	}
	if c.Environment.ChainID != 0 {
		vars["L2_CHAIN_ID"] = fmt.Sprintf("%d", c.Environment.ChainID)
	}
	if c.Environment.DeployerKey != "" {
		vars["DEPLOYER_KEY"] = c.Environment.DeployerKey
	}
	if c.Environment.SeedKey != "" {
		vars["SEED_KEY"] = c.Environment.SeedKey
	}
	return vars
}
