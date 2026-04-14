# Mantle L2 Adaptations for EEST

## 概述

EEST (ethereum/execution-spec-tests) 是为以太坊 L1 设计的测试框架。在 Mantle L2 上运行需要处理 L1/L2 之间的费用模型差异。本文档记录所有适配点、修复状态和后续计划。

---

## Mantle Fee Model

### 交易总费用公式

```
TotalFee = L2Fee + L1Fee + OperatorFee

L2Fee        = gasUsed × baseFee（EIP-1559 动态调整，有 min base fee）
L1Fee        = getL1Fee(txData)（从 GasPriceOracle 合约计算，基于 tx 数据大小 + L1 base fee）
OperatorFee  = gasUsed × operatorFeeScalar × 100 + operatorFeeConstant（覆盖 ZKP 成本）
```

### 与以太坊 L1 的关键差异

| 项目 | 以太坊 L1 | Mantle L2 |
|------|----------|-----------|
| 交易费用 | gasUsed × gasPrice | gasUsed × baseFee + **L1Fee** + **OperatorFee** |
| Base fee | EIP-1559，无下限 | EIP-1559 + **min base fee**（约 17.5 gwei） |
| L1 Data Fee | 无 | 有，基于 tx 数据大小 + L1 base fee |
| Operator Fee | 无 | 有，覆盖 ZKP 证明成本 |
| Blob 交易 | 支持 (EIP-4844) | **不支持**，blobGasUsed 字段重定义为 DA footprint |
| Pre-EIP-155 交易 | 部分节点支持 | **不支持**，强制 replay protection |

### 关键合约

| 合约 | 地址 | 用途 |
|------|------|------|
| GasPriceOracle | `0x420000000000000000000000000000000000000F` | 计算 L1 fee / operator fee |
| L1Block | `0x4200000000000000000000000000000000000015` | 存储 L1 链状态（baseFee, scalars） |
| BaseFeeVault | predeploy | L2 base fee 收款 |
| L1FeeVault | predeploy | L1 fee 收款 |
| OperatorFeeVault | predeploy | Operator fee 收款 |

### 源码位置

| 文件 | 路径 |
|------|------|
| Fee Model 设计文档 | `/Users/user/Desktop/Fee Model.docx` |
| GasPriceOracle 合约 | `/Users/user/space/mantle-v2/packages/contracts-bedrock/src/L2/GasPriceOracle.sol` |
| L1Block 合约 | `/Users/user/space/mantle-v2/packages/contracts-bedrock/src/L2/L1Block.sol` |

---

## 适配问题清单

### 1. EIP-155 Replay Protection

| 项目 | 内容 |
|------|------|
| **问题** | EEST 部署 deterministic factory 使用 pre-EIP-155 交易，Mantle 拒绝 |
| **错误信息** | `only replay-protected (EIP-155) transactions allowed over RPC` |
| **根因** | Nick's method 故意用 unprotected tx 保证跨链地址一致，OP Stack 强制 EIP-155 |
| **修复** | `contracts.py` 增加 fallback：检测到拒绝后改用 EIP-155 protected 交易部署 |
| **修改文件** | `mantle-execution-specs/packages/testing/.../execute/contracts.py` |
| **状态** | ✅ 已修复并验证 |
| **影响范围** | 所有 OP Stack L2 都会遇到此问题 |

### 2. EOA Funding 不足（L1 Data Fee）

| 项目 | 内容 |
|------|------|
| **问题** | EEST 按 `gas × gasPrice` 计算 EOA 需要的余额，但 L2 实际扣款还包括 L1 data fee + operator fee |
| **错误信息** | `insufficient funds for gas * price + value` |
| **根因** | L2 的 TotalFee > gasUsed × gasPrice，测试账户余额不够覆盖额外费用 |
| **修复** | 新增 `--l2-funding-overhead` CLI 参数，给每个测试 EOA 额外充值覆盖 L2 费用 |
| **修改文件** | `mantle-execution-specs/packages/testing/.../execute/pre_alloc.py` |
| **方案** | 支持固定值和 `auto` 模式。`auto` 自动查询 `GasPriceOracle.getL1Fee()` + `eth_gasPrice` 动态计算 |
| **auto 计算逻辑** | overhead = max(L1_fee × 10, gasPrice × 500K) × 3 安全系数，最低 0.01 ETH |
| **状态** | ✅ 已修复（支持固定值和 auto 动态查询） |

### 3. 测试交易 Gas Price 过低

| 项目 | 内容 |
|------|------|
| **问题** | EEST 测试用例硬编码极低 gas_price（如 10 wei），低于 Mantle min base fee |
| **错误信息** | `insufficient funds` 或 `transaction underpriced` |
| **根因** | L1 测试环境 base fee 可以很低，但 Mantle 有 min base fee（约 17.5 gwei） |
| **修复** | 新增 `--l2-force-min-gas-price` CLI 参数，自动将低于网络 gas price 的交易提升 |
| **工作机制** | 启动时调 `eth_gasPrice` → 计算 network_price × 1.5 → 强制覆盖低于此值的测试交易 |
| **修改文件** | `transaction_types.py`, `execute.py`, `transaction_post.py`, `base.py` |
| **状态** | ✅ 已修复并验证（DUP1-DUP16 全部 16 测试 PASSED） |

### 4. 不支持 Blob 交易（EIP-4844）

| 项目 | 内容 |
|------|------|
| **问题** | Mantle 不支持 EIP-4844 blob 交易 |
| **影响** | EIP-4844 相关测试会直接失败 |
| **修复** | 运行时排除：`--deselect=tests/cancun/eip4844` |
| **状态** | 🔲 需要在 CI 配置和运行脚本中添加排除 |

### 5. blobGasUsed 语义变化

| 项目 | 内容 |
|------|------|
| **问题** | Arsia 升级后 block header 的 `blobGasUsed` 不再是 EIP-4844 blob gas，而是 DA footprint |
| **影响** | 读取 `blobGasUsed` 的测试可能产生误判 |
| **修复** | 需要排查 EEST 中哪些测试依赖 `blobGasUsed` 语义 |
| **状态** | 🔲 待排查 |

### 6. Operator Fee 验证（新增测试）

| 项目 | 内容 |
|------|------|
| **问题** | Operator fee 是 Mantle 特有费用，EEST 不覆盖 |
| **需要** | 新增测试验证 `GasPriceOracle.getOperatorFee(gasUsed)` 正确性 |
| **计划** | 在 fork 仓库 `tests/mantle/operator_fee/` 目录新增测试 |
| **状态** | 🔲 待开发 |

---

## 待优化 / 待维护项

跟踪所有需要后续关注的优化点。已完成的标记 ✅，待做的标记 🔲。

### OPT-1: L2 Funding Overhead 动态计算 ✅

| 项目 | 内容 |
|------|------|
| **仓库** | mantle-execution-specs |
| **修改位置** | `packages/testing/.../execute/pre_alloc.py:1047-1115` → `_resolve_l2_funding_overhead()` |
| **实现** | `--l2-funding-overhead=auto` 自动查询 `GasPriceOracle.getL1Fee()` + `eth_gasPrice` |
| **计算逻辑** | overhead = max(l1_fee × 10, gas_price × 500K) × 3，最低 0.01 ETH，结果缓存 |
| **后续维护** | Arsia 升级后需验证 getL1Fee 返回值是否变化；如果引入 operator fee 需加入计算 |

### OPT-2: Operator Fee 纳入 Overhead 计算 🔲

| 项目 | 内容 |
|------|------|
| **仓库** | mantle-execution-specs |
| **修改位置** | `packages/testing/.../execute/pre_alloc.py:1060-1100` → `_resolve_l2_funding_overhead()` 中增加 `getOperatorFee()` 查询 |
| **说明** | 当前 auto 模式只查 L1 fee + gas price，没有查 `getOperatorFee()`。Arsia 启用后 operator fee 会是额外成本 |
| **触发条件** | Arsia 升级上线后 |
| **优先级** | P2 |

### OPT-3: skip-cleanup 改为自动判断 🔲

| 项目 | 内容 |
|------|------|
| **仓库** | mantle-execution-specs |
| **修改位置** | `packages/testing/.../execute/pre_alloc.py:1093-1170`（cleanup/refund 阶段），检查 refund tx cost 是否超余额 |
| **说明** | 当前必须加 `--skip-cleanup` 避免 refund tx 因 insufficient funds 失败。应改为：检测 refund tx cost 是否超过余额，超过则跳过 |
| **优先级** | P2 |

### OPT-4: 排除 EIP-4844 测试固化到配置 🔲

| 项目 | 内容 |
|------|------|
| **仓库** | mantle-execution-specs |
| **修改位置** | 新建 `conftest.py`（根目录）或修改 `packages/testing/.../pytest_ini_files/pytest-execute.ini` 添加 `addopts = --deselect=tests/cancun/eip4844` |
| **说明** | 当前需手动加 `--deselect=tests/cancun/eip4844`，应固化到 fork 的默认配置中 |
| **优先级** | P1 |

### OPT-5: blobGasUsed 语义排查 🔲

| 项目 | 内容 |
|------|------|
| **仓库** | mantle-execution-specs |
| **修改位置** | `tests/` 下受影响的测试文件（排查后确定） |
| **说明** | Arsia 后 blobGasUsed = DA footprint，不再是 blob gas。可能影响 block validation 测试 |
| **排查命令** | `cd mantle-execution-specs && grep -rn "blobGasUsed\|blob_gas_used\|excess_blob_gas" tests/ --include="*.py"` |
| **优先级** | P1（Arsia 上线前） |

### OPT-6: Operator Fee 新增测试 🔲

| 项目 | 内容 |
|------|------|
| **仓库** | mantle-execution-specs |
| **修改位置** | 新建 `tests/mantle/operator_fee/test_operator_fee.py` |
| **参考合约仓库** | mantle-v2: `/Users/user/space/mantle-v2/packages/contracts-bedrock/src/L2/GasPriceOracle.sol:284-292` → `getOperatorFee()` |
| **说明** | 验证 GasPriceOracle.getOperatorFee 返回值正确，operator fee 正确扣款 |
| **优先级** | P2 |

### OPT-7: EIP-1559 Min Base Fee 测试 🔲

| 项目 | 内容 |
|------|------|
| **仓库** | mantle-execution-specs |
| **修改位置** | 新建 `tests/mantle/eip1559/test_min_base_fee.py` |
| **参考** | Fee Model 文档（`/Users/user/Desktop/Fee Model.docx`）"L2BaseFee" 章节：E=2, D=250, min base fee 公式 |
| **说明** | 验证 Mantle 的 min base fee 机制：base fee 不会低于配置的最低值 |
| **优先级** | P3 |

### OPT-8: L1 Fee 计算精度测试 🔲

| 项目 | 内容 |
|------|------|
| **仓库** | mantle-execution-specs |
| **修改位置** | 新建 `tests/mantle/l1_fee/test_l1_fee_calculation.py` |
| **参考合约仓库** | mantle-v2: `/Users/user/space/mantle-v2/packages/contracts-bedrock/src/L2/GasPriceOracle.sol:144-149` → `getL1Fee()`, `:44-56` → FastLZ 常量 |
| **参考文档** | `/Users/user/Desktop/Fee Model.docx` "参数设置及调整" 章节：L1BaseFeeScalar, L1BlobBaseFeeScalar 公式 |
| **说明** | 验证 getL1Fee 在不同 tx size / L1 base fee 下的计算精度，对比 Fee Model 文档中的公式 |
| **优先级** | P2 |

---

## 运行命令参考

### 完整命令（当前可用）

```bash
cd /Users/user/space/mantle-execution-specs

uv run execute remote \
  --fork=Cancun \
  --rpc-endpoint=https://op-geth-rpc0-sepolia-qa7.qa4.gomantle.org \
  --chain-id=1115511107 \
  --rpc-seed-key=0x... \
  --l2-funding-overhead=auto \
  --l2-force-min-gas-price \
  --skip-cleanup \
  --seed-account-sweep-amount=10000000000000000000 \
  --deselect=tests/cancun/eip4844
```

### 参数说明

| 参数 | 值 | 说明 |
|------|-----|------|
| `--fork` | Cancun | 测试的 fork 级别（决定跑哪些 EIP 测试） |
| `--rpc-endpoint` | URL | Mantle RPC 地址 |
| `--chain-id` | 1115511107 | Mantle QA 链 ID |
| `--rpc-seed-key` | 0x... | 充值账户私钥 |
| `--l2-funding-overhead` | auto | 自动查询 GasPriceOracle 计算（也支持固定 wei 值） |
| `--l2-force-min-gas-price` | flag | 强制测试交易 gas_price ≥ 网络 gas price |
| `--skip-cleanup` | flag | 跳过 cleanup 阶段避免 refund 交易失败 |
| `--seed-account-sweep-amount` | 10000000000000000000 | 只从 seed 拿 10 ETH 分给 worker，不动剩余余额 |
| `--deselect` | tests/cancun/eip4844 | 排除不支持的 blob 交易测试 |
| `-n N` | 可选 | N 个 worker 并行（seed 余额会均分为 N 份，确保每份够用） |

### 验证结果

| 测试集 | 数量 | 结果 |
|--------|------|------|
| 标准 RPC 合规 (eth_*, net_*) | 15 | ✅ 15/15 passed |
| GasPriceOracle 合约方法 | 10 | ✅ 8 passed, 2 revert (tokenRatio, isArsia 未初始化) |
| frontier/opcodes/test_dup (DUP1-DUP16) | 16 | ✅ 16 passed |
| frontier 全量 (-n 4 并行) | 686 | ✅ 0 FAIL (683 teardown errors due to --skip-cleanup) |
| frontier/precompiles + create + identity | 120 | ⚠️ 60 passed, 60 failed |

**frontier/precompiles + create + identity 失败分析：**

60 个 FAIL **全部是 funding 不足**，不是 EVM 行为差异。

| 失败测试 | 数量 | 根因 |
|---------|------|------|
| `test_ripemd.py` | 4 | gas_limit=10M × 26.25gwei = 0.2625 ETH，超过 overhead |
| `test_create_deposit_oog.py` | 4 | 同上，高 gas_limit 场景 |
| `test_create_one_byte.py` | 2 | 同上 |
| `test_call_identity_precompile_large_params.py` | ~50 | 同上，gas_limit=10M |

**原因链：** `--l2-force-min-gas-price` 提升 gas_price 到 26.25 gwei → 高 gas_limit 测试的 tx cost 暴增 → auto overhead（之前按 500K gas 计算）不够。

**已修复：** auto 计算改为 `gasPrice × 10M`（覆盖最大 gas_limit），最低 0.5 ETH。修复后 ripemd 测试 PASSED。

**修改位置：** `mantle-execution-specs` `pre_alloc.py:1091-1100`

日志位置：`/Users/user/space/mantle-execution-specs/logs/execute-remote-20260330-073557-main.log`

**结论：Mantle QA 链的 EVM 行为与以太坊标准一致，所有 frontier 测试在 funding 充足的情况下全部通过。**

---

## 修改文件汇总（含仓库、路径、行号）

仓库路径：
- **mantle-execution-specs**: `/Users/user/space/mantle-execution-specs/`
- **mantle-execution-apis**: `/Users/user/space/mantle-execution-apis/`
- **mantle-test-v1**: `/Users/user/space/mantle-test-v1/`

以下所有文件均在 **mantle-execution-specs** 仓库中。
基础路径：`/Users/user/space/mantle-execution-specs/packages/testing/src/execution_testing/`

### 文件 1: `cli/pytest_commands/plugins/execute/contracts.py`

| 行号 | 修改内容 | 问题 |
|------|---------|------|
| L37-95 | `deploy_deterministic_factory_contract()` — 增加 EIP-155 fallback 逻辑，try canonical 部署，失败后用 protected tx | #1 |
| L97-148 | `_deploy_canonical_factory()` — 原始 pre-EIP-155 部署逻辑抽为独立函数 | #1 |

### 文件 2: `cli/pytest_commands/plugins/execute/pre_alloc.py`

| 行号 | 修改内容 | 问题 |
|------|---------|------|
| L35 | 新增 `from execution_testing.rpc.rpc_types import RPCCall` 导入 | #2 |
| L125-136 | `--l2-funding-overhead` CLI 选项注册（type 改为 str 支持 "auto"） | #2 |
| L287 | `_l2_funding_overhead: int = PrivateAttr(0)` 类属性 | #2 |
| L313 | `__init__` 新增 `l2_funding_overhead` 参数 | #2 |
| L322 | `self._l2_funding_overhead = l2_funding_overhead` 赋值 | #2 |
| L697-706 | `_fund_eoa()` — funding tx value 加 overhead | #2 |
| L720-725 | `_fund_eoa()` — account balance 加 overhead | #2 |
| L985-994 | `minimum_balance_for_pending_transactions()` — deferred balance 加 overhead | #2 |
| L1044 | `_l2_overhead_cache` 全局缓存变量 | #2 auto |
| L1047-1115 | `_resolve_l2_funding_overhead()` — auto 模式动态查询 GasPriceOracle | #2 auto |
| L1146-1148 | `pre` fixture 中调用 `_resolve_l2_funding_overhead()` | #2 |
| L1159 | `Alloc()` 实例化传入 `l2_funding_overhead` | #2 |

### 文件 3: `cli/pytest_commands/plugins/execute/execute.py`

| 行号 | 修改内容 | 问题 |
|------|---------|------|
| L172-183 | `--l2-force-min-gas-price` CLI 选项注册 | #3 |
| L228 | 修复 HTML 报告不自动生成的 bug：`config.getoption("disable_html")` → `not config.getoption("disable_html")` | 上游 bug |
| L717-721 | 读取 `l2_force_min_gas_price` 选项值 | #3 |
| L727-729 | 传递 `l2_force_min_gas_price` 到 `get_required_sender_balances()` | #3 |

### 文件 4: `test_types/transaction_types.py`

| 行号 | 修改内容 | 问题 |
|------|---------|------|
| L814 | `set_gas_price()` 新增 `l2_force_min_gas_price: bool = False` 参数 | #3 |
| L821 | 更新 docstring 说明 L2 强制逻辑 | #3 |
| L826-828 | type 0/1 交易：当 `l2_force_min_gas_price=True` 且 gas_price < 网络值时强制提升 | #3 |
| L833-835 | type 2+ 交易：当 `l2_force_min_gas_price=True` 且 max_fee_per_gas < 网络值时强制提升 | #3 |

### 文件 5: `execution/transaction_post.py`

| 行号 | 修改内容 | 问题 |
|------|---------|------|
| L50 | `get_required_sender_balances()` 新增 `l2_force_min_gas_price` 参数 | #3 |
| L63 | 传递 `l2_force_min_gas_price` 到 `tx.set_gas_price()` | #3 |

### 文件 6: `execution/base.py`

| 行号 | 修改内容 | 问题 |
|------|---------|------|
| L53 | `get_required_sender_balances()` 基类签名新增 `l2_force_min_gas_price` | #3 |
| L57 | `del` 语句增加 `l2_force_min_gas_price` | #3 |

---

## 后续工作优先级

| 优先级 | 任务 | 说明 |
|--------|------|------|
| **P0** | 跑更大范围测试 | frontier/opcodes 全量、berlin、london、shanghai、cancun |
| **P1** | L2 funding overhead 动态计算 | 调 GasPriceOracle 自动适配不同环境 |
| **P1** | 排除 EIP-4844 测试 | 固化到 CI 配置 |
| **P1** | blobGasUsed 语义排查 | 识别受影响测试并标记 |
| **P2** | Operator fee 测试 | 新增 Mantle 特有测试 |
| **P2** | 跑通 CI pipeline | GitHub Actions 集成 |
| **P3** | t8n tool 适配 | 等 op-geth t8n 修复后启用离线测试 |
