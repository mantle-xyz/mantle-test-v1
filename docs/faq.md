# FAQ 常见问题

## 框架设计

### Q: mantle-test-v1 是做什么的？
A: 通用测试编排平台。它不实现测试逻辑，只负责调度各测试模块、管理环境、收集报告。每个模块在自己的仓库有自己的测试和 CI，编排器通过 YAML manifest 调用它们。

### Q: 为什么需要编排器？各模块已经有自己的 CI 了
A: 各模块 CI 验证"自己的代码没 bug"，编排器验证"Mantle 链的整体行为是否正确"。比如 EEST 验证 EVM 标准一致性、execution-apis 验证 RPC 合规，这些是链级别的行为验证，不是某个模块的单元测试。

### Q: 如何集成新模块？
A: 写一个 YAML manifest 放到 `orchestrator/modules/`，声明 command、env_vars、result_format、environments。零代码改动，orchestrator 自动发现。

---

## EEST (EVM 一致性)

### Q: EEST 测什么？
A: 以太坊 Execution Layer 规范一致性。包括 opcode gas 消耗、precompile 输入输出、交易执行、状态转换。5000+ 用例，覆盖 Frontier → Cancun 所有 EIP。

### Q: EEST 怎么验证 gas？
A: 不是读 receipt.gasUsed。而是在合约里用 `GAS` opcode 记录执行前后的 gas 差值，存到 storage，在 post-state 里断言这个值 == 标准定义的 gas。精确到**单个 opcode 级别**。

### Q: EEST 验证结果还是中间过程？
A: 两者都验证。验证 post-state（storage 每个 slot、balance 每个 wei）是验证结果。用 `GAS` opcode 差值验证 gas 消耗是验证中间过程。

### Q: EEST 的边界条件覆盖够吗？
A: 远超 chainregression。每个 opcode 都有 stack overflow/underflow、OOG（out of gas）、零值、最大值、revert 等边界测试。precompile 有不同长度输入（127 bytes、129 bytes、10000 bytes、空输入）。3073 个 test function 参数化展开后 5000+ 用例。

### Q: 两个客户端（op-geth 和 reth）都 PASS EEST 就够了吗？不需要 diff？
A: 够了。EEST fixture 定义了精确的 expected post-state。如果两个客户端都 PASS = 两个客户端的 post-state 都精确等于 expected = 两个客户端行为一致。不需要额外的 diff 工具。

### Q: EEST 会发真实交易到链上吗？
A: `execute remote` 模式会发真实交易。只能在 localchain/qa 上跑，不能在 mainnet 跑。

### Q: 如何添加 EEST 测试用例？
A: 在 `mantle-execution-specs/tests/` 目录下写 Python 函数，放进去就自动被发现。不需要注册。示例：
```python
def test_my_case(state_test: StateTestFiller, pre: Alloc):
    sender = pre.fund_eoa()
    contract = pre.deploy_contract(code=Op.SSTORE(0, Op.ADD(1, 2)) + Op.STOP)
    tx = Transaction(to=contract, gas_limit=100000, sender=sender)
    post = {contract: Account(storage={0: 3})}
    state_test(env=Environment(), pre=pre, tx=tx, post=post)
```

### Q: fee 相关测试应该放 EEST 里吗？
A: 不应该。EEST 验证的是 EVM 的 gas 消耗（execution gas），不验证 L2 的总费用（L1 data fee + operator fee）。L2 fee 在 EVM 之外扣款，EEST 看不到。fee 测试放 op-acceptance。

### Q: 如何只跑单个测试？
A: `uv run execute remote ... "tests/path/file.py::test_function[fork_Cancun-state_test]"`。用 `--collect-only` 可以列出所有用例不执行。

---

## execution-apis (RPC 合规)

### Q: execution-apis 是做什么的？
A: 定义以太坊 JSON-RPC 接口规范（OpenRPC 格式）。每个 RPC method 的参数类型、返回值类型、错误码都有明确定义。不是测试框架，是**规范文档**。

### Q: 那谁来跑测试？
A: execution-apis 仓库里的两个工具：
- `rpctestgen`：对客户端跑 221 个语义测试（参数对不对 + 返回值对不对 + 值对不对）
- `speccheck`：验证测试文件格式是否符合 OpenRPC schema

### Q: rpctestgen 会校验返回值对不对吗？
A: 会。它启动一个已知状态的链，调 RPC，验证返回的**精确值**（不只是格式）。比如 `eth_getBalance` 返回值必须是 `0x56`（预置状态里的余额），不是"只要是 hex string 就行"。

### Q: EEST 和 execution-apis 的区别？
A: EEST 验证"EVM 计算对不对"（通过 RPC 发交易验证状态转换），execution-apis 验证"RPC 接口对不对"（参数格式、返回类型、错误码、值的正确性）。EEST 用 RPC 作为通信通道，但不测 RPC 本身。

### Q: rpctestgen 支持远程 RPC 吗？
A: 目前不支持。它只支持自己启动本地 geth 节点。需要适配支持远程 RPC 后才能对 Mantle 节点跑。

---

## Mantle L2 适配

### Q: Mantle 直接跑 EEST 需要适配吗？
A: 需要 4 处适配，否则跑不了：
1. **EIP-155**：Mantle 拒绝 unprotected tx，EEST 部署 factory 用的是 pre-EIP-155 tx
2. **L2 funding**：L2 交易成本包含 L1 data fee，EEST 按 L1 模型算的充值金额不够
3. **Min gas price**：EEST 测试用例写死 gas_price=10 wei，Mantle 有 min base fee ~17.5 gwei
4. **EIP-4844**：Mantle 不支持 blob tx，需排除相关测试

### Q: EIP-155 是什么？为什么需要适配？
A: EIP-155 在交易签名中加入 chain ID，防止跨链重放。Mantle 强制要求 EIP-155。EEST 部署 deterministic factory 故意用 pre-EIP-155 tx（为了跨链地址一致），被 Mantle 拒绝。适配方式：检测到拒绝后 fallback 到 EIP-155 protected tx。

### Q: EIP-155 fallback 后合约地址不同了有影响吗？
A: 没影响。factory 地址不同但功能一样。EEST 不硬编码 factory 地址，都是动态计算的。后续测试用例通过 factory 用 CREATE2 部署的合约地址也是动态计算的。

### Q: funding overhead 为什么用 auto？
A: `--l2-funding-overhead=auto` 自动查询链上 `GasPriceOracle.getL1Fee()` + `eth_gasPrice` 动态计算。不同环境（localchain/qa/mainnet）的 L1 base fee 不同，auto 自动适配。

### Q: 为什么需要 `--seed-account-sweep-amount`？
A: EEST 并行模式（`-n 4`）默认把 seed 全部余额均分给 worker。51000 ETH 分 4 份每份 12000+ ETH，浪费且退款困难。设 10 ETH 只用 10 ETH，每个 worker 2.5 ETH 足够。

### Q: worker 地址的残留余额怎么回收？
A: EEST 日志中搜 `Initializing EOA iterator with start index`，这个 start index 就是 worker 的私钥。用这个私钥发退款交易即可。

---

## Hive (多客户端)

### Q: Hive 支持 L2 吗？
A: Hive 本身只支持 L1 客户端测试。EEST 有内置的 `execute hive` 命令，但也是 L1 模式（不包含 L2 predeploy、sequencer、L1 data fee）。L2 使用 Hive 需要适配。

### Q: 那多客户端比对怎么做？
A: 不需要 Hive。直接用 EEST `execute remote` 分别对 op-geth RPC 和 reth RPC 跑同一套测试。两个都 PASS = 行为一致。

### Q: Hive 什么时候需要？
A: 需要验证 Engine API / P2P / 同步 这些 EEST 不覆盖的层面时。比如两个客户端能否正常 P2P 发现和同步区块。

---

## chainregression 迁移

### Q: chainregression 弃用后怎么办？
A: 按功能迁移到对应模块：opcode/precompile/EIP → EEST，RPC method → execution-apis，deposit/withdraw → op-e2e，operator fee/gas oracle → op-acceptance。详见 [migration-plan.md](migration-plan.md)。

### Q: estimateGas 测试迁移到哪？
A: op-acceptance Mantle gate。因为 Mantle 的 `eth_estimateGas` 返回值包含 L1 data fee 折算，`receipt.gasUsed` 只有 execution gas。这是 L2 特有逻辑，不是 EVM 规范，不放 EEST。

### Q: EEST 能替代 chainregression 的 opcode 测试吗？
A: 能，而且更全。chainregression 的 opcode 测试 ~58 个函数，EEST 有 5000+ 用例覆盖所有 opcode 的边界条件。
