# Usage Guide 使用指南

## Prerequisites 前置依赖

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.25+ | Build orchestrator |
| Python | 3.12+ | Run EEST (in mantle-execution-specs) |
| uv | latest | Python package manager (for EEST) |
| Go | 1.22+ | Build execution-apis tools (in mantle-execution-apis) |

## Repository Layout 仓库布局

```
/Users/user/space/
├── mantle-test-v1/               # Orchestrator (this project)
├── mantle-execution-specs/       # EVM conformance (fork of ethereum/execution-specs)
├── mantle-execution-apis/        # RPC conformance (fork of ethereum/execution-apis)
└── mantle-v2/                    # op-e2e + op-acceptance (already adapted)
```

---

## 1. Orchestrator 编排器

### Build 构建

```bash
cd /Users/user/space/mantle-test-v1/orchestrator
go build -o bin/mantle-test ./cmd/mantle-test/
```

### View Plan 查看执行计划

```bash
# localchain: all modules
./bin/mantle-test plan --config=configs/localchain.yaml

# QA: EEST + RPC + acceptance
./bin/mantle-test plan --config=configs/qa.yaml

# Specific module only
./bin/mantle-test plan --config=configs/qa.yaml --modules=eest

# Mainnet: RPC only (read-only)
./bin/mantle-test plan --config=configs/mainnet.yaml
```

### Run Tests 执行测试

```bash
# All modules
./bin/mantle-test run --config=configs/localchain.yaml

# Specific module
./bin/mantle-test run --config=configs/qa.yaml --modules=eest

# Fail fast
./bin/mantle-test run --config=configs/localchain.yaml --fail-fast
```

---

## 2. EEST (EVM Conformance) 直接运行

在 mantle-execution-specs 仓库中直接运行，不通过 orchestrator：

```bash
cd /Users/user/space/mantle-execution-specs

# Install dependencies (first time only)
uv sync

# Run against Mantle RPC
uv run execute remote \
  --fork=Cancun \
  --rpc-endpoint=$L2_RPC_URL \
  --chain-id=$L2_CHAIN_ID \
  --rpc-seed-key=$SEED_KEY \
  --l2-funding-overhead=auto \
  --l2-force-min-gas-price \
  --skip-cleanup \
  --seed-account-sweep-amount=10000000000000000000 \
  --deselect=tests/cancun/eip4844

# Parallel (4 workers)
uv run execute remote ... -n 4

# Specific tests only
uv run execute remote ... tests/frontier/opcodes/test_dup.py

# HTML report auto-generated at execution_results/report_execute.html
```

### Mantle L2 Required Parameters Mantle L2 必需参数

| Parameter | Description |
|-----------|-------------|
| `--l2-funding-overhead=auto` | Auto-query GasPriceOracle for L1 data fee overhead |
| `--l2-force-min-gas-price` | Force test tx gas_price ≥ network base fee |
| `--skip-cleanup` | Skip refund phase (L2 refund tx may fail) |
| `--seed-account-sweep-amount=10000000000000000000` | Only use 10 ETH from seed, not entire balance |
| `--deselect=tests/cancun/eip4844` | Exclude blob tx tests (Mantle doesn't support EIP-4844) |

---

## 3. execution-apis (RPC Conformance) 直接运行

```bash
cd /Users/user/space/mantle-execution-apis

# Build tools (first time)
make tools

# Build OpenRPC spec
make build

# Run spec check
make test
```

---

## 4. op-e2e 直接运行

```bash
cd /Users/user/space/mantle-v2/op-e2e

# All action tests
make test-actions

# Mantle-specific tests
go test ./actions/mantletests/... -v -count=1

# Mantle upgrade tests
go test ./actions/mantleupgrades/... -v -count=1
```

---

## 5. op-acceptance 直接运行

```bash
cd /Users/user/space/mantle-v2/op-acceptance-tests

# Base gate
go run cmd/main.go --gate=base --testdir=. --validators=./acceptance-tests.yaml

# Mantle gate (if configured)
go run cmd/main.go --gate=mantle-arsia --testdir=. --validators=./acceptance-tests.yaml
```

---

## 6. Adding a New Module 添加新模块

Create a YAML manifest in `orchestrator/modules/`:

```yaml
name: my-module
description: What this module tests
repo: repository-name

suites:
  - name: suite-name
    phase: unit | integration | e2e | acceptance
    environments: [localchain, qa, mainnet]
    command: "shell command to execute"
    env_vars: [L2_RPC_URL, L2_CHAIN_ID, ...]
    result_format: gotest-json | junit-xml | eest-json
    timeout: 30m
```

No code changes needed. The orchestrator auto-discovers it from `modules/` directory.

---

## 7. Syncing Fork Repos 同步 Fork 仓库

```bash
# mantle-execution-specs
cd /Users/user/space/mantle-execution-specs
git fetch upstream
git merge upstream/forks/amsterdam
# Resolve conflicts → push

# mantle-execution-apis
cd /Users/user/space/mantle-execution-apis
git fetch upstream
git merge upstream/main
# Resolve conflicts → push
```

---

## 8. CI

Each module has its own CI in its own repo:

| Module | CI File | Trigger |
|--------|---------|---------|
| EEST | `mantle-execution-specs/.github/workflows/mantle-test.yaml` | Push to mantle/main, manual dispatch |
| execution-apis | `mantle-execution-apis/.github/workflows/mantle-test.yaml` | Push to mantle/main |
| op-e2e | `mantle-v2/.github/workflows/` (existing) | Already configured |
| op-acceptance | `mantle-v2/` (existing) | Already configured |

The orchestrator CI (`mantle-test-v1/.github/workflows/test.yml`) triggers module CIs and collects reports.

---

## 9. Test Reports & GitHub Pages 测试报告

Reports are published to GitHub Pages at: `https://mantle-xyz.github.io/mantle-test-v1/`

测试报告通过 GitHub Pages 自动发布。

### 9.1 目录结构

```
reports/
├── eest/                              # EEST (EVM Conformance)
│   ├── 20260408-080000.html           # 按时间戳命名
│   ├── 20260407-160000.html
│   └── ...
├── execution-apis/                    # execution-apis (RPC)
│   ├── 20260408-080000.html
│   └── ...
├── op-e2e/                            # op-e2e
│   ├── 20260408-080000.html
│   └── ...
├── op-acceptance/                     # op-acceptance
│   ├── 20260408-080000.html
│   └── ...
└── <any-new-module>/                  # 新模块自动出现
    └── ...
```

**命名规则：** `reports/<module>/<timestamp>.html`

### 9.2 各模块报告来源

| 模块 | 报告生成位置 | 报告文件 | 推送命令 |
|------|------------|---------|---------|
| **EEST** | `mantle-execution-specs/execution_results/` | `report_execute.html` + `assets/` | `./orchestrator/scripts/upload-report.sh eest execution_results/report_execute.html` |
| **execution-apis** | `mantle-execution-apis/` | `make test` 的 stdout（需重定向为 HTML） | `./orchestrator/scripts/upload-report.sh execution-apis speccheck-report.html` |
| **op-e2e** | `mantle-v2/op-e2e/` | go test 输出（需 `-json` + 转 HTML） | `./orchestrator/scripts/upload-report.sh op-e2e test-report.html` |
| **op-acceptance** | `mantle-v2/op-acceptance-tests/logs/testrun-*/` | `results.html` + `static/` | `./orchestrator/scripts/upload-report.sh op-acceptance logs/testrun-*/results.html` |

### 9.3 模块 CI 自动推送

各模块 CI 跑完后，自动推送报告到 mantle-test-v1：

```yaml
# 在任意模块的 .github/workflows/xxx.yml 最后加：
- name: Push report to mantle-test-v1
  run: |
    TIMESTAMP=$(date -u +%Y%m%d-%H%M%S)
    MODULE="eest"  # 改成你的模块名
    REPORT_FILE="execution_results/report_execute.html"  # 改成你的报告路径

    cd /tmp
    git clone https://x-access-token:${{ secrets.CROSS_REPO_TOKEN }}@github.com/mantle-xyz/mantle-test-v1.git
    cd mantle-test-v1
    mkdir -p reports/${MODULE}/
    cp ${{ github.workspace }}/${REPORT_FILE} reports/${MODULE}/${TIMESTAMP}.html
    git add reports/
    git commit -m "Add ${MODULE} report ${TIMESTAMP}"
    git push
```

**各模块完整 CI 配置：**

**EEST (mantle-execution-specs):**
```yaml
# .github/workflows/mantle-test.yaml 最后加：
# 报告位置: execution_results/report_execute.html (EEST 自动生成)
- name: Push report
  if: always()
  run: |
    TIMESTAMP=$(date -u +%Y%m%d-%H%M%S)
    cd /tmp && git clone https://x-access-token:${{ secrets.CROSS_REPO_TOKEN }}@github.com/mantle-xyz/mantle-test-v1.git
    cd mantle-test-v1
    mkdir -p reports/eest/
    cp ${{ github.workspace }}/execution_results/report_execute.html reports/eest/${TIMESTAMP}.html
    git add reports/ && git commit -m "Add eest report ${TIMESTAMP}" && git push
```

**execution-apis (mantle-execution-apis):**
```yaml
# .github/workflows/mantle-test.yaml 最后加：
# 报告位置: speccheck 输出 stdout → 需要重定向成 HTML
- name: Generate report
  run: |
    ./tools/speccheck -v 2>&1 | tee speccheck.txt
    echo "<html><body><pre>$(cat speccheck.txt)</pre></body></html>" > speccheck.html
- name: Push report
  if: always()
  run: |
    TIMESTAMP=$(date -u +%Y%m%d-%H%M%S)
    cd /tmp && git clone https://x-access-token:${{ secrets.CROSS_REPO_TOKEN }}@github.com/mantle-xyz/mantle-test-v1.git
    cd mantle-test-v1
    mkdir -p reports/execution-apis/
    cp ${{ github.workspace }}/speccheck.html reports/execution-apis/${TIMESTAMP}.html
    git add reports/ && git commit -m "Add execution-apis report ${TIMESTAMP}" && git push
```

**op-e2e (mantle-v2):**
```yaml
# .github/workflows/e2e.yaml 最后加：
# 报告位置: go test -json 输出 → go-test-report 转 HTML
- name: Generate report
  run: |
    cd op-e2e
    go test ./actions/mantletests/... -v -json -count=1 2>&1 | tee test-output.json
    go install github.com/vakenbolt/go-test-report@latest
    cat test-output.json | go-test-report -o report.html
- name: Push report
  if: always()
  run: |
    TIMESTAMP=$(date -u +%Y%m%d-%H%M%S)
    cd /tmp && git clone https://x-access-token:${{ secrets.CROSS_REPO_TOKEN }}@github.com/mantle-xyz/mantle-test-v1.git
    cd mantle-test-v1
    mkdir -p reports/op-e2e/
    cp ${{ github.workspace }}/op-e2e/report.html reports/op-e2e/${TIMESTAMP}.html
    git add reports/ && git commit -m "Add op-e2e report ${TIMESTAMP}" && git push
```

**op-acceptance (mantle-v2):**
```yaml
# .github/workflows/acceptance.yaml 最后加：
# 报告位置: op-acceptance-tests/logs/testrun-*/results.html (op-acceptor 自动生成)
- name: Push report
  if: always()
  run: |
    TIMESTAMP=$(date -u +%Y%m%d-%H%M%S)
    REPORT=$(ls -t ${{ github.workspace }}/op-acceptance-tests/logs/testrun-*/results.html | head -1)
    cd /tmp && git clone https://x-access-token:${{ secrets.CROSS_REPO_TOKEN }}@github.com/mantle-xyz/mantle-test-v1.git
    cd mantle-test-v1
    mkdir -p reports/op-acceptance/
    cp "${REPORT}" reports/op-acceptance/${TIMESTAMP}.html
    git add reports/ && git commit -m "Add op-acceptance report ${TIMESTAMP}" && git push
```

**新模块接入模板：**
```yaml
# 在你的模块 CI workflow 最后加：
- name: Push report
  if: always()
  run: |
    TIMESTAMP=$(date -u +%Y%m%d-%H%M%S)
    MODULE="your-module-name"        # ← 改成你的模块名
    REPORT="path/to/report.html"     # ← 改成你的报告路径
    cd /tmp && git clone https://x-access-token:${{ secrets.CROSS_REPO_TOKEN }}@github.com/mantle-xyz/mantle-test-v1.git
    cd mantle-test-v1
    mkdir -p reports/${MODULE}/
    cp ${{ github.workspace }}/${REPORT} reports/${MODULE}/${TIMESTAMP}.html
    git add reports/
    git commit -m "Add ${MODULE} report ${TIMESTAMP}"
    # Retry push in case of concurrent commits from other modules
    for i in 1 2 3; do
      git push && break || { git pull --rebase && git push && break; } || sleep 5
    done
```

**注意事项：**
- 报告文件按 `<timestamp>.html` 命名，**不会覆盖**已有报告
- 如果本地跑多次测试，EEST 的 `report_execute.html` 会被最新一次覆盖，建议**跑完立即推送**
- 多个模块 CI 同时推送时可能 git 冲突，模板里的 retry + rebase 会自动处理

### 9.4 通过 workflow_dispatch 触发收集

模块 CI 把报告上传为 artifact，然后触发 mantle-test-v1 收集：

```yaml
# 在任意项目的 CI 最后加：
- name: Upload report artifact
  uses: actions/upload-artifact@v4
  with:
    name: test-report
    path: report.html

- name: Trigger report collection
  uses: peter-evans/repository-dispatch@v3
  with:
    token: ${{ secrets.CROSS_REPO_TOKEN }}
    repository: mantle-xyz/mantle-test-v1
    event-type: module-ci-complete
    client-payload: |
      {
        "module": "my-module",
        "repo": "${{ github.repository }}",
        "workflow": "test.yml",
        "artifact": "test-report"
      }
```

mantle-test-v1 的 `pages.yml` 收到 `module-ci-complete` 事件后，自动下载 artifact 并 commit 到 `reports/<module>/`。

### 9.5 手动上传

```bash
# 上传 EEST 报告
./orchestrator/scripts/upload-report.sh eest /Users/user/space/mantle-execution-specs/execution_results/report_execute.html

# 上传 op-acceptance 报告
./orchestrator/scripts/upload-report.sh op-acceptance /Users/user/space/mantle-v2/op-acceptance-tests/logs/testrun-*/results.html

# 推送触发 Pages 部署
git add reports/ && git commit -m "Add report" && git push
```

### 9.6 GitHub Pages 首页

`pages.yml` 自动生成 `reports/index.html`，按模块 → 时间倒序展示：

```
Mantle Test Reports

── EEST (EVM Conformance) ──────
│ 20260408-080000.html     [最新]
│ 20260407-160000.html
│ 20260406-120000.html

── execution-apis (RPC) ────────
│ 20260408-080000.html
│ 20260407-160000.html

── op-e2e ──────────────────────
│ 20260408-080000.html

── op-acceptance ───────────────
│ 20260408-080000.html
```
