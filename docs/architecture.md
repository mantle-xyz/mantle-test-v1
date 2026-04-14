# Architecture

## Overview

![Test Platform Architecture](images/test-platform.svg)

![Data Flow](images/data-flow.svg)

mantle-test-v1 is a **generic test orchestration platform** for Mantle chain-level behavioral verification. It schedules test modules, manages environments, and collects native reports.

**Core design principle:** The orchestrator does not implement any test logic. Each module is an independent repo with its own tests, CI, and reports. The orchestrator only invokes and collects.

**Extensibility:** Any new module can plug in by adding a single YAML manifest — no code changes to the orchestrator.

## Test Layers

| Layer | Module | Verifies |
|-------|--------|----------|
| EVM Conformance | EEST (mantle-execution-specs) | Opcode gas, precompile, transaction execution, state transition |
| RPC Conformance | execution-apis (mantle-execution-apis) | JSON-RPC method params, return values, error codes |
| OP Stack E2E + Acceptance | op-e2e + op-acceptance-tests | Deposit/withdraw, sequencer, fault proof, operator fee, gas oracle |
| Multi-client | Hive (future) | P2P, sync, Engine API, consensus across clients |

## Orchestrator

```
orchestrator/
├── cmd/mantle-test/       # CLI: run, plan
├── pkg/
│   ├── module/            # YAML manifest parsing + module registry
│   ├── config/            # Environment config + ${ENV_VAR} resolution
│   ├── adapter/           # exec (shell command) + ci-trigger (GitHub API)
│   ├── environment/       # unit / localchain / qa / mainnet
│   ├── phase/             # Phase ordering + DAG dependency + parallel scheduling
│   ├── orchestrator/      # Top-level Run() + report collection
│   └── result/            # Parsers (go test JSON, JUnit XML, EEST)
├── modules/               # Module manifests (YAML)
└── configs/               # Environment configs
```

### Module Plugin Interface

Every module is defined by a YAML manifest:

```yaml
name: module-name
suites:
  - name: suite-name
    phase: unit | integration | e2e | acceptance
    environments: [localchain, qa, mainnet]
    command: "shell command to execute"
    env_vars: [L2_RPC_URL, L2_CHAIN_ID, ...]
    result_format: gotest-json | junit-xml | eest-json
    timeout: 30m
```

The orchestrator:
1. Discovers modules from `modules/*.yaml`
2. Filters by environment and phase
3. Executes `command` with injected env vars
4. Collects native report artifacts to `reports/<timestamp>/<module>/`
5. Multi-client consistency: both clients pass the same EEST = behavior identical (no diff tool needed)

### Data Flow

See [Data Flow Diagram](images/data-flow.svg) above.

### Environment Matrix

| Environment | RPC | Sends Tx | Use Case |
|-------------|-----|----------|----------|
| localchain | localhost | Yes | Local devnet, full testing |
| qa | QA RPC | Yes | QA environment validation |
| mainnet | mainnet RPC | **No** | Read-only RPC conformance |

## Fork Repositories

| Fork | Upstream | Branch | Purpose |
|------|----------|--------|---------|
| mantle-execution-specs | ethereum/execution-specs | mantle/main | EVM test cases + framework |
| mantle-execution-apis | ethereum/execution-apis | mantle/main | RPC spec + test tools |
| mantle-op-acceptance-tests | optimism/op-acceptance-tests | mantle/main | Acceptance tests + Mantle gates |

Each fork:
- `origin` → mantle-xyz (push)
- `upstream` → official repo (sync via `git merge`)
- Has its own CI for running tests independently
- Can be invoked by orchestrator or run standalone
