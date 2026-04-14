# OP Stack 公链测试实践对比

## 各链测试体系一览

| | **Optimism** | **Base** | **Mantle（新方案）** | **其他 OP Stack 链**（Zora, World Chain 等） |
|--|-------------|---------|---------------------|-------------------------------------------|
| **EVM 一致性** | 依赖 op-geth 继承 geth 的 EEST 合规 | 同 Optimism | fork EEST 直接对 Mantle RPC 跑 | 通常不独立做，依赖 op-geth |
| **RPC 合规** | 无专门框架 | 无专门框架 | fork execution-apis (rpctestgen) | 通常不做 |
| **OP Stack 测试** | [ethereum-optimism/tests](https://github.com/ethereum-optimism/tests)（opt8n + opdn 生成 fixture） | 同 Optimism | 复用 op-e2e + op-acceptance | 依赖上游 Optimism |
| **E2E 测试** | op-e2e（173 文件） | 同 Optimism + Base 特有 | op-e2e（已有 Mantle 适配） | 通常直接用上游 op-e2e |
| **验收测试** | op-acceptance-tests（按 gate 分组） | 同 Optimism | op-acceptance + Mantle gate | 通常不做专门验收 |
| **多客户端** | op-geth + op-reth（[正在推进](https://blog.base.dev/scaling-base-with-reth)） | **op-geth → reth 迁移中**（Reth 将成为 canonical client） | op-geth + reth（EEST 两端跑） | 大多只用 op-geth |
| **Hive** | 未独立使用 | 未独立使用 | 未来计划 | 未使用 |
| **链级行为统一入口** | 无（各测试分散） | 无 | **mantle-test-v1 编排器** | 无 |
| **CI** | 各模块独立 CI | 各模块独立 CI | 各模块独立 CI + 编排器触发 | 各模块独立 CI |

---

## Optimism 测试结构

```
ethereum-optimism/
├── optimism/                     # 主仓库
│   ├── op-e2e/                   # E2E 测试（173 文件）
│   │   ├── actions/              # 确定性 action 测试
│   │   ├── system/               # 全系统集成测试（deprecated，新测试用 acceptance）
│   │   └── faultproofs/          # fault proof 测试
│   ├── op-acceptance-tests/      # 验收测试
│   │   ├── tests/base/           # 基础 gate（deposit, chain, withdrawal）
│   │   ├── tests/ecotone/        # Ecotone 升级 gate
│   │   ├── tests/fjord/          # Fjord 升级 gate
│   │   ├── tests/isthmus/        # Isthmus 升级 gate（含 operator_fee）
│   │   └── acceptance-tests.yaml # gate 注册
│   └── devnet-sdk/               # 测试基础设施（chain 连接、环境管理）
│
├── tests/                        # 标准化测试 fixture
│   ├── fixtures/                 # JSON 静态 fixture
│   ├── opt8n/                    # execution test fixture 生成器
│   └── opdn/                     # derivation test fixture 生成器
│
└── op-geth/                      # 执行客户端
    └── (继承 geth 的 EEST 合规)
```

**特点：**
- EVM 一致性依赖 op-geth 继承 geth 的 EEST 合规，不独立跑
- [ethereum-optimism/tests](https://github.com/ethereum-optimism/tests) 仓库有自己的 fixture 格式（不是 EEST）
- op-acceptance-tests 按 fork 升级分 gate（base → ecotone → fjord → isthmus）
- 无统一编排器，各测试分散在不同仓库

---

## Base 测试结构

```
base/
├── node-reth/                    # Reth 客户端（Base 版本）
│   └── (E2E testing crate)       # 内置 E2E 测试
│
└── 复用 Optimism 上游
    ├── op-e2e
    ├── op-acceptance-tests
    └── op-geth
```

**特点：**
- [Base 正在从 op-geth 迁移到 Reth](https://blog.base.dev/scaling-base-with-reth)
- Base 在 Reth 仓库里有内置的 E2E testing crate
- 多客户端运行已帮助 Base 避免了多次宕机
- 测试主要复用 Optimism 上游

---

## Mantle 新方案 vs 行业对比

### Mantle 做了而其他链没做的

| Mantle 独有 | 说明 |
|------------|------|
| **fork EEST 直接对链跑** | 其他 OP Stack 链依赖 op-geth 继承 EEST 合规，不直接验证自己的链 |
| **fork execution-apis** | 其他链没有做 RPC 合规测试 |
| **统一编排器** | 其他链的测试分散在各仓库，没有统一入口 |
| **L2 适配层** | EEST 是为 L1 设计的，Mantle 适配了 EIP-155/funding/gas price 问题 |

### 各链共有的

| 共有 | 说明 |
|------|------|
| op-e2e | 所有 OP Stack 链都用 |
| op-acceptance-tests | Optimism 和 Base 在用，Mantle 也有 |
| 多客户端 | Base 在推 Reth，Mantle 也在做 |

### Mantle 还没做但行业有的

| 行业有 | 说明 | 建议 |
|--------|------|------|
| ethereum-optimism/tests（opt8n + opdn） | OP Stack 特有的 execution + derivation fixture | 评估是否需要接入 |
| Base 的 Reth E2E testing crate | Reth 客户端级别的 E2E | 等 Mantle 迁移 Reth 时参考 |

---

## 参考链接

- [ethereum-optimism/tests](https://github.com/ethereum-optimism/tests) — OP Stack 标准化测试 fixture
- [Scaling Base With Reth](https://blog.base.dev/scaling-base-with-reth) — Base 多客户端迁移实践
- [OP Test Vectors](https://static.optimism.io/tests/) — OP Stack 测试向量说明
- [OP Stack Specification](https://specs.optimism.io/) — OP Stack 协议规范
- [ethereum/execution-spec-tests](https://github.com/ethereum/execution-spec-tests) — EEST 官方仓库
- [ethereum/execution-apis](https://github.com/ethereum/execution-apis) — execution-apis 官方仓库
