package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads a YAML config file, resolves ${ENV_VAR} references, and returns the parsed Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	// Resolve ${ENV_VAR} and ${ENV_VAR:-default} references
	resolved := resolveEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(resolved), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	// Apply defaults
	if cfg.Modules.Dir == "" {
		cfg.Modules.Dir = "modules/"
	}
	if cfg.Execution.Parallel == 0 {
		cfg.Execution.Parallel = 4
	}
	if len(cfg.Execution.Output) == 0 {
		cfg.Execution.Output = []string{"console"}
	}

	return &cfg, nil
}

var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-(.*?))?\}`)

// resolveEnvVars replaces ${VAR} and ${VAR:-default} with environment variable values.
func resolveEnvVars(input string) string {
	return envVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := envVarPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		name := parts[1]
		defaultVal := ""
		if len(parts) >= 3 {
			defaultVal = parts[2]
		}
		if val, ok := os.LookupEnv(name); ok {
			return val
		}
		return defaultVal
	})
}

// MergeOverrides applies CLI --set key=value overrides to a Config.
// Keys use dot notation: "environment.l2_rpc_url", "execution.parallel".
func MergeOverrides(cfg *Config, overrides map[string]string) error {
	for key, val := range overrides {
		switch strings.ToLower(key) {
		case "environment.l2_rpc_url":
			cfg.Environment.L2RPCURL = val
		case "environment.l2_ws_url":
			cfg.Environment.L2WSURL = val
		case "environment.l1_rpc_url":
			cfg.Environment.L1RPCURL = val
		case "environment.chain_id":
			var chainID uint64
			if _, err := fmt.Sscanf(val, "%d", &chainID); err != nil {
				return fmt.Errorf("invalid chain_id %q: %w", val, err)
			}
			cfg.Environment.ChainID = chainID
		case "environment.deployer_key":
			cfg.Environment.DeployerKey = val
		case "environment.seed_key":
			cfg.Environment.SeedKey = val
		case "execution.fail_fast":
			cfg.Execution.FailFast = val == "true"
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
	}
	return nil
}
