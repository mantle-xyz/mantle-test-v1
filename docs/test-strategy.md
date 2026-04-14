# Mantle 链完整测试策略

## 测试层级

![Test Layers](images/test-layers.svg)

---

## 不在本框架范围内

| 测试 | 归属 | 说明 |
|------|------|------|
| 各模块单元测试 | 各项目自有 CI | op-geth `make test`、op-node `go test`、contracts `forge test` 等 |
| L1 合约测试 | mantle-v2/packages/contracts-bedrock | `forge test`，验证合约逻辑 |
| 压测 | 独立项目 | TPS、延迟、稳定性测试 |

---

## Layer 1: EVM 一致性测试 (EEST)

验证 Mantle L2 的 EVM 行为与以太坊标准一致。

| 项目 | 说明 |
|------|------|
| **框架** | ethereum/execution-specs (fork: mantle-execution-specs) |
| **测试数量** | 5000+ fixture，覆盖 Frontier → Cancun 所有 EIP |
| **验证内容** | opcode gas、precompile I/O、交易执行、状态转换 |
| **执行方式** | `uv run execute remote` 对 L2 RPC 发真实交易 |
| **位置** | `/Users/user/space/mantle-execution-specs/` |
| **状态** | ✅ 已跑通，frontier 全量 686 用例 0 FAIL |

---

## Layer 2: RPC 合规测试 (execution-apis)

验证 Mantle 节点的 JSON-RPC 接口符合以太坊规范。

| 项目 | 说明 |
|------|------|
| **框架** | ethereum/execution-apis (fork: mantle-execution-apis) |
| **工具** | `rpctestgen`（语义验证）+ `speccheck`（格式验证） |
| **测试数量** | 30+ RPC method × 多场景 = 221 个测试函数 |
| **验证内容** | method 参数/返回值格式、语义正确性、错误码 |
| **位置** | `/Users/user/space/mantle-execution-apis/` |
| **状态** | ✅ 基础 RPC 合规验证通过（38/38 passed） |

---

## Layer 3: OP Stack 端到端 + 验收测试

### 3a: op-e2e (端到端)

| 项目 | 说明 |
|------|------|
| **位置** | `/Users/user/space/mantle-v2/op-e2e/` |
| **验证内容** | |
| - Sequencer 出块 | sequencer 正常产出 L2 block |
| - Batcher 提交 | batch 正确提交到 L1 |
| - Deposit 全流程 | L1→L2 deposit 从发起到 L2 确认 |
| - Withdraw 全流程 | L2→L1 withdraw 从发起到 L1 finalize |
| - Fault Proof | fault proof 挑战/响应全流程 |
| - Interop | 跨链互操作 |
| **执行方式** | `go test ./...` (启动 L1+L2 全节点) |

### 3b: op-acceptance-tests (验收)

| 项目 | 说明 |
|------|------|
| **位置** | `/Users/user/space/optimism/op-acceptance-tests/` |
| **验证内容** | 链升级后的验收检查，按 gate 分组 |
| - base gate | 基础 smoke (deposit, chain, withdrawal) |
| - ecotone gate | Ecotone 升级验收 |
| - fjord gate | Fjord 升级验收 |
| - isthmus gate | Isthmus 升级验收 |
| **Mantle 应新增** | |
| - **operator fee gate** | operator fee 计算、扣款、vault 余额 |
| - **gas oracle gate** | L1 data fee 计算、GasPriceOracle 合约行为 |
| - **DA footprint gate** | blobGasUsed 重定义、DA footprint 计算 |
| - **min base fee gate** | EIP-1559 + min base fee 行为 |
| **执行方式** | `op-acceptor` 读取 `acceptance-tests.yaml` 按 gate 执行 |

---

## Layer 4: 多客户端集成测试 (Hive)

验证多客户端在网络层面的一致性。**多客户端迁移（op-geth → reth）的核心验证工具。**

| 项目 | 说明 |
|------|------|
| **框架** | ethereum/hive + EEST `execute hive` 命令 |
| **验证内容** | |
| - 同一套 EEST 测试 | 分别对 op-geth 和 reth 跑，自动比对结果 |
| - Engine API | execution client 出块行为 |
| - 同步 | 客户端间同步一致性 |
| - 多客户端共识 | 同一交易在不同客户端产生相同 state root |
| **执行方式** | `./hive --dev --client mantle-opgeth` + `uv run execute hive --fork=Cancun` |
| **优先级** | P0 — 多客户端迁移必需 |

---

## 独立项目：压测

不属于功能测试体系，单独仓库/项目管理。

| 项目 | 说明 |
|------|------|
| **现有** | `/Users/user/space/mantle-test/testkit/chainregression/benchmark/` |
| **验证内容** | TPS、gas 消耗、延迟、稳定性 |
| **执行方式** | 独立部署，对 QA/staging 链持续压测 |

---

## 覆盖矩阵

| 验证内容 | EEST (L1) | RPC (L2) | op-e2e (L3) | 验收 (L3) | Hive (L4) |
|---------|:---------:|:--------:|:-----------:|:---------:|:---------:|
| opcode gas | ● | | | | |
| precompile | ● | | | | |
| 交易类型 | ● | | ● | ● | |
| 状态转换 | ● | | | | |
| RPC 格式+语义 | | ● | | | |
| deposit/withdraw | | | ● | ● | |
| sequencer/batcher | | | ● | | |
| fault proof | | | ● | | |
| P2P/sync | | | | | ● |
| Engine API | | | | | ● |
| 多客户端一致性 | ● | | | | ● |
| operator fee | | | | ● | |
| gas oracle / L1 fee | | | | ● | |
| DA footprint | | | | ● | |
| min base fee | | | | ● | |
| 链升级验收 | | | | ● | |

● = 主要覆盖

---

## CI Pipeline

![CI Pipeline](images/ci-pipeline.svg)

Each module has its own CI in its own repo. The orchestrator CI triggers them and collects reports.

各模块在各自仓库有独立 CI。编排器 CI 负责触发和收集报告。

---

## chainregression 迁移计划

chainregression 弃用，其测试迁移到对应层级：

| chainregression 目录 | 迁移到 | 层级 |
|---------------------|--------|------|
| `opcode/` | EEST 已覆盖 | Layer 1 |
| `precompiles/` | EEST 已覆盖 | Layer 1 |
| `EIP/` | EEST 已覆盖 | Layer 1 |
| `rpc/method/` | execution-apis | Layer 2 |
| `rpc/estimategas/` | op-acceptance (gas oracle gate) | Layer 3 |
| `evm/operatorfee/` | op-acceptance (operator fee gate) | Layer 3 |
| `transaction/deposit/` | op-e2e / op-acceptance | Layer 3 |
| `transaction/withdraw/` | op-e2e / op-acceptance | Layer 3 |
| `benchmark/` | 独立压测项目 | 不在本框架 |
| `token/` | op-acceptance | Layer 3 |

---

## 各仓库位置

| 仓库 | 本地路径 | 层级 | 用途 |
|------|---------|------|------|
| mantle-test-v1 | `/Users/user/space/mantle-test-v1/` | Orchestrator | 编排平台 |
| mantle-execution-specs | `/Users/user/space/mantle-execution-specs/` | Layer 1 | EVM 一致性测试 (fork) |
| mantle-execution-apis | `/Users/user/space/mantle-execution-apis/` | Layer 2 | RPC 合规测试 (fork) |
| mantle-v2 | `/Users/user/space/mantle-v2/` | Layer 3 | op-e2e + op-acceptance（已适配） |
| mantle-test (弃用) | `/Users/user/space/mantle-test/` | — | chainregression → 迁移到上述各层 |
