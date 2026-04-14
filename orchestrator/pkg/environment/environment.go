package environment

import (
	"context"
	"fmt"

	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/config"
	"github.com/mantle-xyz/mantle-test/orchestrator/pkg/module"
)

// Environment provides connection info and lifecycle management for a test environment.
type Environment interface {
	Type() module.EnvironmentType
	EnvVars() map[string]string
	Setup(ctx context.Context) error
	Teardown(ctx context.Context) error
}

// New creates an Environment based on the config's environment type.
func New(cfg *config.Config) (Environment, error) {
	switch cfg.Environment.Type {
	case module.EnvUnit:
		return &unitEnv{}, nil
	case module.EnvLocalChain:
		return &localChainEnv{cfg: cfg}, nil
	case module.EnvQA:
		return &remoteEnv{cfg: cfg, envType: module.EnvQA}, nil
	case module.EnvMainnet:
		return &remoteEnv{cfg: cfg, envType: module.EnvMainnet}, nil
	default:
		return nil, fmt.Errorf("unsupported environment type: %s", cfg.Environment.Type)
	}
}

// unitEnv is a no-op environment for unit tests that need no external resources.
type unitEnv struct{}

func (e *unitEnv) Type() module.EnvironmentType    { return module.EnvUnit }
func (e *unitEnv) EnvVars() map[string]string       { return nil }
func (e *unitEnv) Setup(_ context.Context) error    { return nil }
func (e *unitEnv) Teardown(_ context.Context) error { return nil }

// localChainEnv manages a local devnet (docker-compose/kurtosis).
type localChainEnv struct {
	cfg *config.Config
}

func (e *localChainEnv) Type() module.EnvironmentType { return module.EnvLocalChain }

func (e *localChainEnv) EnvVars() map[string]string {
	return e.cfg.EnvVars()
}

func (e *localChainEnv) Setup(ctx context.Context) error {
	// TODO: implement devnet startup (docker-compose up / kurtosis run)
	// For now, assume the devnet is already running and config provides URLs.
	return nil
}

func (e *localChainEnv) Teardown(ctx context.Context) error {
	// TODO: implement devnet shutdown
	return nil
}

// remoteEnv represents QA or mainnet environments where the chain is already running.
type remoteEnv struct {
	cfg     *config.Config
	envType module.EnvironmentType
}

func (e *remoteEnv) Type() module.EnvironmentType { return e.envType }

func (e *remoteEnv) EnvVars() map[string]string {
	return e.cfg.EnvVars()
}

func (e *remoteEnv) Setup(_ context.Context) error    { return nil }
func (e *remoteEnv) Teardown(_ context.Context) error { return nil }
