# chainregression 迁移计划

## 概述

chainregression 共 67 个测试文件、约 250 个测试函数，迁移到 4 个目标模块。

| 目标模块 | 测试文件数 | 测试函数数 | 说明 |
|---------|-----------|-----------|------|
| EEST (mantle-execution-specs) | 23 | ~100 | EVM + 标准 EIP |
| execution-apis (mantle-execution-apis) | 5 | ~10 | RPC method 合规 |
| op-e2e / op-acceptance (mantle-v2) | 30 | ~110 | L2 业务逻辑 |
| 不需要迁移 | 9 | ~30 | helper/lib/bindings/benchmark |

---

## 1. → EEST（EVM 一致性）

EEST 已有 5000+ 用例覆盖以下内容。**不需要手动迁移代码，只需确认 EEST 覆盖后弃用 chainregression 对应文件。**

### 1.1 opcode 测试

| 原文件 | 函数数 | 测试内容 | EEST 覆盖 | 迁移方式 |
|--------|-------|---------|----------|---------|
| `opcode/computational/computational_test.go` | 28 | ADD/SUB/MUL/DIV/MOD/EXP 等算术 opcode | `tests/frontier/opcodes/test_all_opcodes.py` 含所有 opcode | 弃用，EEST 覆盖更全 |
| `opcode/evmcontrol/evmcontrol_test.go` | 30 | CALL/DELEGATECALL/STATICCALL/CREATE/CREATE2 | `tests/frontier/opcodes/test_call.py` 等 | 弃用，EEST 覆盖更全 |

### 1.2 precompile 测试

| 原文件 | 函数数 | 测试内容 | EEST 覆盖 | 迁移方式 |
|--------|-------|---------|----------|---------|
| `precompiles/precompiles_v1_test.go` | 8 | ecrecover/sha256/ripemd/identity/modexp/bn128 | `tests/frontier/precompiles/test_precompiles.py` | 弃用 |
| `precompiles/precompiles_v2_test.go` | 11 | 同上 + 边界测试 | 同上 | 弃用 |
| `precompiles/v3/precompiles_v3_test.go` | 3 | 新增 precompile | `tests/cancun/` 相关 | 弃用 |
| `precompiles/delegate_call/delegatecall_test.go` | 2 | DELEGATECALL 调用 precompile | EEST 已覆盖 | 弃用 |
| `precompiles/cast_precompiles_test.sh` | — | shell 脚本测试 | EEST 已覆盖 | 弃用 |

### 1.3 EVM storage 测试

| 原文件 | 函数数 | 测试内容 | EEST 覆盖 | 迁移方式 |
|--------|-------|---------|----------|---------|
| `evm/storage/consistency_test.go` | 1 | SLOAD/SSTORE 一致性 | `tests/berlin/eip2929_gas_cost_increases/` | 弃用 |

### 1.4 标准 EIP 测试

| 原文件 | 函数数 | EIP | EEST 对应 | 迁移方式 |
|--------|-------|-----|----------|---------|
| `EIP/eip1153/eip1153_test.go` | 3 | TSTORE/TLOAD | `tests/cancun/eip1153_tstore/` (6 文件) | 弃用 |
| `EIP/eip1559/eip1559_opcode_test.go` | 1 | BASEFEE opcode | `tests/london/eip1559_fee_market_change/` | 弃用 |
| `EIP/eip2935/eip2935_test.go` | 1 | BLOCKHASH from state | `tests/prague/eip2935_*/` | 弃用 |
| `EIP/eip3855/eip3855_test.go` | 1 | PUSH0 | `tests/shanghai/eip3855_push0/` | 弃用 |
| `EIP/eip4788/eip4788_test.go` | 2 | Beacon root | `tests/cancun/eip4788_beacon_root/` | 弃用 |
| `EIP/eip4844/eip4844_test.go` | 1 | Blob tx | `tests/cancun/eip4844_blobs/` | ⚠️ Mantle 不支持，排除 |
| `EIP/eip4895/eip4895_test.go` | 5 | Withdrawals | `tests/shanghai/eip4895_withdrawals/` | 弃用 |
| `EIP/eip5656/eip5656_test.go` | 2 | MCOPY | `tests/cancun/eip5656_mcopy/` | 弃用 |
| `EIP/eip6780/eip6780_v2_test.go` | 1 | SELFDESTRUCT 限制 | `tests/cancun/eip6780_selfdestruct/` | 弃用 |
| `EIP/eip7516/eip7516_test.go` | 2 | BLOBBASEFEE opcode | `tests/cancun/eip7516_blobgasfee/` | ⚠️ Mantle blobGasUsed 语义不同 |
| `EIP/eip7685/eip7685_test.go` | 1 | General purpose EL requests | `tests/prague/eip7685_*/` | 弃用 |
| `EIP/eip7702/eip7702_test.go` | 4 | Set code tx | `tests/prague/eip7702_set_code_tx/` | 弃用 |
| `EIP/eip7939/eip7939_clz_test.go` | 2 | CLZ opcode | `tests/osaka/eip7939_*/` | 弃用 |
| `transaction/transactiontype/transaction_type_test.go` | 1 | 交易类型 | `tests/berlin/eip2930_*` + `tests/london/eip1559_*` | 弃用 |

### 1.5 需确认 EEST 覆盖情况

| 原文件 | 函数数 | EIP | 状态 | 行动 |
|--------|-------|-----|------|------|
| `EIP/eip7212/eip7212_secp256r1_test.go` | 1 | secp256r1 precompile | 🔲 需确认 EEST 是否已有 | 如无需在 fork 中新增 |
| `EIP/eip7823eip7883/eip7823_eip7883_test.go` | 2 | modexp gas increase | `tests/osaka/eip7883_modexp_gas_increase/` | 确认后弃用 |
| `EIP/eip7934/eip7934_test.go` | 3 | — | 🔲 需确认 | — |

---

## 2. → execution-apis（RPC 合规）

RPC method 格式和语义验证迁移到 execution-apis 的 rpctestgen。

| 原文件 | 函数数 | 测试内容 | 迁移到 | 迁移方式 |
|--------|-------|---------|--------|---------|
| `rpc/method/http_method_test.go` | 5 | eth_getBlockByNumber 等 HTTP 调用 | rpctestgen 已有 221 个语义测试 | 弃用 |
| `rpc/method/compare_client_rpc_http_test.go` | 1 | 多客户端 RPC 对比 | rpctestgen 对各客户端分别跑 | 弃用 |
| `rpc/method/rpc_block_receipts_test.go` | 2 | eth_getBlockReceipts | rpctestgen 已有 | 弃用 |
| `rpc/method/rpc_sync_test.go` | 1 | eth_syncing | rpctestgen 已有 | 弃用 |
| `rpc/gasrefund/gasrefund_test.go` | 0 | gas refund 机制 | EEST EVM 层已覆盖 | 弃用 |

---

## 3. → op-e2e（OP Stack 端到端）

L1↔L2 跨链和 OP Stack 特有功能迁移到 mantle-v2/op-e2e，大部分已有对应测试。

| 原文件 | 函数数 | 测试内容 | op-e2e 对应 | 迁移方式 |
|--------|-------|---------|------------|---------|
| `transaction/deposit/deposit_test.go` | 14 | L1→L2 deposit 全流程 | `actions/mantletests/derivation/` | 已有，弃用 |
| `transaction/withdraw/withdraw_test.go` | 1 | L2→L1 withdraw | `system/` withdraw 相关 | 已有，弃用 |
| `transaction/withdraw/prove_test.go` | 1 | withdraw prove | `faultproofs/` | 已有，弃用 |
| `transaction/withdraw/withdraw_buz_main_test.go` | 1 | withdraw 业务主流程 | 同上 | 已有，弃用 |
| `transaction/transfer_to_optimism_portal_proxy_test.go` | 2 | Portal 交互 | `system/bridge/` | 已有，弃用 |
| `rpc/l1rpc/l1rpc_test.go` | 2 | L1 RPC 交互 | `actions/mantletests/` | 已有，弃用 |
| `rpc/optimismsafeheadatl1block/optimism_safeHeadAtL1Block_test.go` | 1 | safe head 查询 | `actions/mantletests/safedb/` | 已有，弃用 |
| `transaction/preconf/preconf_test.go` | 2 | preconfirmation | — | 🔲 需在 op-e2e 新增 |
| `transaction/preconf/simple_preconf_test.go` | 1 | 简单 preconf | — | 🔲 需在 op-e2e 新增 |
| `transaction/preconf/priority_verification_test.go` | 7 | 优先级验证 | — | 🔲 需在 op-e2e 新增 |
| `transaction/preconf/batch_whitelist_transfer_test.go` | 2 | 批量白名单 | — | 🔲 需在 op-e2e 新增 |
| `transaction/preconf/txpool_scenarios_test.go` | 6 | txpool 场景 | — | 🔲 需在 op-e2e 新增 |
| `transaction/preconf/mpt_depth_constructor_test.go` | 6 | MPT 深度 | — | 🔲 需在 op-e2e 新增 |
| `transaction/preconf/address_generation_test.go` | 2 | 地址生成 | — | 🔲 需在 op-e2e 新增 |
| `transaction/preconf/long_running_priority_verification_test.go` | 1 | 长时间优先级 | — | 🔲 需在 op-e2e 新增 |
| `transaction/preconf/preconf_stress_test.go` | 1 | preconf 压测 | — | → 独立压测项目 |

---

## 4. → op-acceptance Mantle gate（Mantle 特有验收）

Mantle L2 特有的业务逻辑验收测试。

| 原文件 | 函数数 | 测试内容 | acceptance gate | 迁移方式 |
|--------|-------|---------|----------------|---------|
| `evm/operatorfee/gas_price_oracle_test.go` | 3 | GasPriceOracle 合约：getL1Fee、getOperatorFee、gasPrice | `mantle-arsia` gate | 🔲 需新建测试 |
| `rpc/estimategas/estimategas_vs_actualgas_test.go` | 1 | estimateGas 预估 vs receipt.gasUsed 偏差 | `mantle-arsia` gate | 🔲 需新建测试 |
| `rpc/estimategas/estimate_gas_compare_test.go` | 1 | 多节点 estimateGas 一致性 | `mantle-arsia` gate | 🔲 需新建测试 |
| `rpc/estimategas/estimate_gas_for_diffrent_transfer_test.go` | 16 | 不同交易类型 gas 估算 | `mantle-arsia` gate | 🔲 需新建测试 |
| `rpc/estimategas/estimate_gas_using_different_account_test.go` | 1 | 不同账户 gas 估算 | `mantle-arsia` gate | 🔲 需新建测试 |
| `rpc/estimategas/estimategas_comprehensive_test.go` | 4 | 综合 gas 估算 | `mantle-arsia` gate | 🔲 需新建测试 |
| `rpc/estimategas/integrated/estimate_gas_integrated_test.go` | 2 | 集成 gas 估算 | `mantle-arsia` gate | 🔲 需新建测试 |
| `rpc/token_ratio_test.go` | 1 | token ratio 验证 | `mantle-arsia` gate | 🔲 需新建测试 |
| `rpc/metaTX/metatx_test.go` | 3 | meta transaction | `mantle-arsia` gate | 🔲 需新建测试 |
| `transaction/transferERC20_test.go` | 1 | ERC20 转账 | `mantle-arsia` gate | 🔲 需新建测试 |
| `transaction/transfer_native_test.go` | 2 | 原生转账 | `mantle-arsia` gate | 🔲 需新建测试 |
| `transaction/unprtectedTX_test.go` | 2 | unprotected tx 拒绝 | `mantle-arsia` gate | 🔲 需新建测试 |
| `EIP/eip7951/eip7951_test.go` | 2 | Mantle 自定义 EIP | `mantle-arsia` gate | 🔲 需新建测试 |
| `token/erc20/` | — | ERC20 操作 | `mantle-arsia` gate | 🔲 需新建测试 |
| `token/erc20nodeps/` | — | 无依赖 ERC20 | `mantle-arsia` gate | 🔲 需新建测试 |

---

## 5. 不需要迁移

| 原文件 | 原因 |
|--------|------|
| `helper/compile_contract_test.go` | 工具类测试，不是链测试 |
| `holder/holder_test.go` | 临时数据管理 |
| `op-bindings/ast/canonicalize_test.go` | binding 工具测试，mantle-v2 已有 |
| `op-bindings/hardhat/hardhat_test.go` | binding 工具测试 |
| `op-bindings/predeploys/addresses_test.go` | binding 工具测试 |
| `benchmark/` | → 独立压测项目 |
| `config/` | 各模块有自己的配置 |
| `lib/` | 各模块有自己的工具库 |
| `transaction/preconf/preconf_stress_test.go` | → 独立压测项目 |

---

## 6. 各模块迁移后的目标结构

### 6.1 EEST (mantle-execution-specs)

不需要迁移代码。EEST 已有覆盖，只需确认后弃用 chainregression 对应文件。

如果发现 EEST 缺少某些测试（如 eip7212），在 fork 仓库新增：

```
mantle-execution-specs/tests/
├── frontier/                    # 以太坊官方测试（不改）
│   ├── opcodes/                 # ← chainregression opcode/ 已被覆盖
│   ├── precompiles/             # ← chainregression precompiles/ 已被覆盖
│   └── ...
├── cancun/                      # 以太坊官方测试（不改）
│   ├── eip1153_tstore/          # ← chainregression EIP/eip1153/ 已被覆盖
│   └── ...
└── mantle/                      # ← Mantle 补充测试（仅 EEST 不足时新增）
    └── eip7212_secp256r1/       # 如果 EEST 没有，从 chainregression 迁移
        ├── __init__.py
        └── test_secp256r1.py
```

### 6.2 execution-apis (mantle-execution-apis)

不需要迁移代码。rpctestgen 221 个语义测试已覆盖 chainregression `rpc/method/` 的内容。

如果需要新增 Mantle 自定义 RPC method：

```
mantle-execution-apis/
├── src/
│   ├── eth/                     # 标准 eth_* method 定义（不改）
│   ├── debug/                   # 标准 debug_* method 定义（不改）
│   └── mantle/                  # ← Mantle 自定义 method（新增）
│       └── mantle_getL1DataFee.yaml
├── tests/
│   ├── eth_getBalance/          # 标准测试（不改）
│   └── mantle_getL1DataFee/     # ← Mantle 自定义 method 测试（新增）
│       └── get-l1-data-fee.io
└── tools/
    └── testgen/generators.go    # ← 如需新增 Mantle method 的语义测试
```

### 6.3 op-e2e (mantle-v2/op-e2e)

已有测试大部分已覆盖 deposit/withdraw。preconf 相关需新增：

```
mantle-v2/op-e2e/
├── actions/
│   └── mantletests/
│       ├── derivation/          # 已有
│       ├── proofs/              # 已有
│       ├── batcher/             # 已有
│       ├── sequencer/           # 已有
│       ├── sync/                # 已有
│       └── preconf/             # ← 新增：从 chainregression 迁移
│           ├── preconf_test.go                     # ← transaction/preconf/preconf_test.go
│           ├── simple_preconf_test.go              # ← transaction/preconf/simple_preconf_test.go
│           ├── priority_verification_test.go       # ← transaction/preconf/priority_verification_test.go
│           ├── batch_whitelist_transfer_test.go    # ← transaction/preconf/batch_whitelist_transfer_test.go
│           ├── txpool_scenarios_test.go            # ← transaction/preconf/txpool_scenarios_test.go
│           ├── mpt_depth_constructor_test.go       # ← transaction/preconf/mpt_depth_constructor_test.go
│           ├── address_generation_test.go          # ← transaction/preconf/address_generation_test.go
│           └── long_running_priority_test.go       # ← transaction/preconf/long_running_priority_verification_test.go
├── system/
│   └── mantleda/                # 已有
└── mantleopgeth/                # 已有
```

### 6.4 op-acceptance (mantle-v2/op-acceptance-tests)

需要新建 `tests/mantle/` 目录，按功能分子目录：

```
mantle-v2/op-acceptance-tests/
├── acceptance-tests.yaml         # ← 新增 mantle-arsia gate
├── tests/
│   ├── base/                     # 已有：deposit, chain, withdrawal
│   ├── ecotone/                  # 已有
│   ├── isthmus/                  # 已有：operator_fee（可复用）
│   └── mantle/                   # ← 新建
│       ├── gas_estimation/       # ← 从 chainregression rpc/estimategas/ 迁移
│       │   ├── main_test.go
│       │   ├── estimategas_vs_actual_test.go       # ← estimategas_vs_actualgas_test.go (1 函数)
│       │   ├── estimategas_cross_node_test.go      # ← estimate_gas_compare_test.go (1 函数)
│       │   ├── estimategas_tx_types_test.go        # ← estimate_gas_for_diffrent_transfer_test.go (16 函数)
│       │   ├── estimategas_accounts_test.go        # ← estimate_gas_using_different_account_test.go (1 函数)
│       │   ├── estimategas_comprehensive_test.go   # ← estimategas_comprehensive_test.go (4 函数)
│       │   └── estimategas_integrated_test.go      # ← integrated/estimate_gas_integrated_test.go (2 函数)
│       │
│       ├── operator_fee/         # ← 从 chainregression evm/operatorfee/ 迁移
│       │   ├── main_test.go
│       │   └── gas_price_oracle_test.go            # ← gas_price_oracle_test.go (3 函数)
│       │                                           #    getL1Fee / getOperatorFee / gasPrice 验证
│       │
│       ├── token_ratio/          # ← 从 chainregression rpc/token_ratio_test.go 迁移
│       │   └── token_ratio_test.go                 # ← token_ratio_test.go (1 函数)
│       │
│       ├── meta_tx/              # ← 从 chainregression rpc/metaTX/ 迁移
│       │   └── metatx_test.go                      # ← metatx_test.go (3 函数)
│       │
│       ├── transfers/            # ← 从 chainregression transaction/ 部分迁移
│       │   ├── transfer_native_test.go             # ← transfer_native_test.go (2 函数)
│       │   ├── transfer_erc20_test.go              # ← transferERC20_test.go (1 函数)
│       │   └── unprotected_tx_test.go              # ← unprtectedTX_test.go (2 函数)
│       │
│       ├── token/                # ← 从 chainregression token/ 迁移
│       │   ├── erc20_test.go                       # ← token/erc20/
│       │   └── erc20_nodeps_test.go                # ← token/erc20nodeps/
│       │
│       └── eip7951/              # ← 从 chainregression EIP/eip7951/ 迁移
│           └── eip7951_test.go                     # ← eip7951_test.go (2 函数)
```

**acceptance-tests.yaml 新增 Mantle gate：**

```yaml
# mantle-v2/op-acceptance-tests/acceptance-tests.yaml

- id: mantle-arsia
  inherits:
    - base
  description: "Mantle Arsia fork acceptance tests"
  tests:
    - package: .../tests/mantle/gas_estimation
      timeout: 30m
    - package: .../tests/mantle/operator_fee
      timeout: 10m
    - package: .../tests/mantle/token_ratio
      timeout: 5m
    - package: .../tests/mantle/meta_tx
      timeout: 10m
    - package: .../tests/mantle/transfers
      timeout: 10m
    - package: .../tests/mantle/token
      timeout: 10m
    - package: .../tests/mantle/eip7951
      timeout: 10m
```

---

## 7. 执行计划

### Phase 1: 确认 EEST 覆盖（1 周）

| 任务 | 说明 |
|------|------|
| 跑 EEST 全量（frontier → cancun）对 Mantle QA 链 | 确认 5000+ 用例全部 PASS |
| 逐个 EIP 对比 chainregression 和 EEST 的覆盖范围 | 确认 EEST 覆盖 ≥ chainregression |
| 确认 eip7212 / eip7823 / eip7934 在 EEST 中的覆盖 | 如不足需在 fork 新增 |

### Phase 2: 弃用 EEST 已覆盖的部分（同步进行）

| 任务 | 说明 |
|------|------|
| 标记 `opcode/`、`precompiles/`、标准 `EIP/` 为 deprecated | 在 chainregression 仓库标记 |
| 确认 EEST CI 跑通后删除对应 chainregression 测试 | 逐步删除 |

### Phase 3: 迁移 op-e2e 部分（1 周）

| 任务 | 说明 |
|------|------|
| 确认 deposit/withdraw 在 op-e2e 已有 | 对比测试场景 |
| 新增 preconf 相关测试到 op-e2e | 28 个测试函数 |

### Phase 4: 新建 op-acceptance Mantle gate（2 周）

| 任务 | 说明 |
|------|------|
| 在 mantle-v2 op-acceptance 中新建 `tests/mantle/` | 目录结构 |
| 迁移 estimateGas 测试（25 个函数）| gas 估算验证 |
| 迁移 operatorfee 测试（3 个函数）| GasPriceOracle 合约 |
| 迁移 token_ratio、metaTX、ERC20 等 | Mantle 特有业务 |
| 注册 `mantle-arsia` gate 到 acceptance-tests.yaml | |

### Phase 5: 清理（1 周）

| 任务 | 说明 |
|------|------|
| 全量回归：通过新框架跑一遍所有迁移后的测试 | 确认无遗漏 |
| 归档 chainregression 仓库 | 标记为 archived |
| 更新文档 | 迁移完成状态 |

---

## 8. 迁移状态跟踪

| 原目录 | 函数数 | 目标 | 状态 |
|--------|-------|------|------|
| `opcode/computational/` | 28 | EEST | 🔲 待确认覆盖后弃用 |
| `opcode/evmcontrol/` | 30 | EEST | 🔲 待确认覆盖后弃用 |
| `precompiles/` (全部) | 24 | EEST | 🔲 待确认覆盖后弃用 |
| `evm/storage/` | 1 | EEST | 🔲 待确认覆盖后弃用 |
| `EIP/eip1153` ~ `eip7939` (标准) | 32 | EEST | 🔲 待确认覆盖后弃用 |
| `EIP/eip7212` | 1 | EEST (待确认) | 🔲 |
| `EIP/eip7823eip7883` | 2 | EEST (待确认) | 🔲 |
| `EIP/eip7934` | 3 | EEST (待确认) | 🔲 |
| `EIP/eip7951` | 2 | op-acceptance | 🔲 需新建 |
| `transaction/transactiontype/` | 1 | EEST | 🔲 待确认覆盖后弃用 |
| `rpc/method/` | 9 | execution-apis | 🔲 rpctestgen 适配后弃用 |
| `rpc/gasrefund/` | 0 | EEST | 🔲 弃用 |
| `rpc/estimategas/` (全部) | 26 | op-acceptance | 🔲 需新建 |
| `rpc/token_ratio_test.go` | 1 | op-acceptance | 🔲 需新建 |
| `rpc/metaTX/` | 3 | op-acceptance | 🔲 需新建 |
| `rpc/l1rpc/` | 2 | op-e2e | 🔲 已有对应 |
| `rpc/optimismsafeheadatl1block/` | 1 | op-e2e | 🔲 已有对应 |
| `evm/operatorfee/` | 4 | op-acceptance | 🔲 需新建 |
| `transaction/deposit/` | 14 | op-e2e | 🔲 已有对应 |
| `transaction/withdraw/` | 3 | op-e2e | 🔲 已有对应 |
| `transaction/preconf/` | 28 | op-e2e | 🔲 需新建 |
| `transaction/transferERC20` | 1 | op-acceptance | 🔲 需新建 |
| `transaction/transfer_native` | 2 | op-acceptance | 🔲 需新建 |
| `transaction/unprtectedTX` | 2 | op-acceptance | 🔲 需新建 |
| `transaction/transfer_to_optimism_portal` | 2 | op-e2e | 🔲 已有对应 |
| `token/` | — | op-acceptance | 🔲 需新建 |
| `benchmark/` | — | 独立项目 | 不迁移 |
| `helper/` / `lib/` / `op-bindings/` | — | 不需要 | 不迁移 |
