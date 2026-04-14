# Module Onboarding Guide 模块接入指南

## 完整接入一个新模块的步骤

以接入 `proxyd` 测试为例，完整演示整个流程。

---

## Step 1: 写 Module Manifest

在 `orchestrator/modules/` 下创建 YAML 文件：

```yaml
# orchestrator/modules/proxyd.yaml
name: proxyd
description: Proxyd RPC proxy integration tests

source:
  # 本地模式：clone 到本地执行
  local:
    repo: https://github.com/mantlenetworkio/mantle-test.git
    branch: main
    path: mantle-test    # clone 到 ../mantle-test/

  # CI 模式：触发目标仓库的 workflow
  ci:
    repo: mantlenetworkio/mantle-test          # GitHub owner/repo
    workflow: proxyd-test.yaml                  # 目标仓库里的 workflow 文件名
    event: mantle-test                          # repository_dispatch event type

suites:
  - name: proxyd-smoke
    phase: e2e
    environments: [localchain, qa]             # 在哪些环境跑
    command: |
      cd ${MODULE_DIR:-../mantle-test}/testkit/chainregression &&
      go test ./proxyd/... -v -json -count=1
    env_vars: [MODULE_DIR, L2_RPC_URL, L2_CHAIN_ID]   # orchestrator 注入的环境变量
    result_format: gotest-json                  # 结果解析格式
    timeout: 30m
```

### 字段说明

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | ✅ | 模块唯一标识，用于 `--modules=proxyd` |
| `description` | ✅ | 模块说明 |
| `source.local.repo` | ✅ | git clone URL |
| `source.local.branch` | ✅ | 分支名 |
| `source.local.path` | ✅ | clone 到本地的目录名（相对于 mantle-test-v1 的上级目录） |
| `source.ci.repo` | ✅ | GitHub `owner/repo`（用于 API 调用） |
| `source.ci.workflow` | ✅ | 目标仓库的 workflow 文件名 |
| `source.ci.event` | ✅ | `repository_dispatch` 的 event type |
| `suites[].name` | ✅ | 套件名 |
| `suites[].phase` | ✅ | `unit` / `integration` / `e2e` / `acceptance` |
| `suites[].environments` | ✅ | 支持的环境列表 |
| `suites[].command` | ✅ | 本地模式执行的 shell 命令 |
| `suites[].env_vars` | | orchestrator 注入的环境变量名 |
| `suites[].result_format` | ✅ | `gotest-json` / `junit-xml` / `eest-json` |
| `suites[].timeout` | ✅ | 超时时间 |

---

## Step 2: 目标仓库配置 Workflow（CI 模式需要）

在目标仓库添加一个 workflow，支持 `repository_dispatch` 触发：

```yaml
# mantlenetworkio/mantle-test/.github/workflows/proxyd-test.yaml
name: Proxyd Tests

on:
  push:
    branches: [main]
  pull_request:
  repository_dispatch:
    types: [mantle-test]      # ← 和 manifest 里的 event 一致
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Run proxyd tests
        env:
          L2_RPC_URL: ${{ github.event.client_payload.L2_RPC_URL || secrets.L2_RPC_URL }}
          L2_CHAIN_ID: ${{ github.event.client_payload.L2_CHAIN_ID || secrets.L2_CHAIN_ID }}
        run: |
          cd testkit/chainregression
          go test ./proxyd/... -v -count=1

      # 推送报告到 mantle-test-v1（可选）
      - name: Push report
        if: always()
        env:
          PLAN: ${{ github.event.client_payload.plan || '' }}
        run: |
          TIMESTAMP=$(date -u +%Y%m%d-%H%M%S)
          FILENAME="${TIMESTAMP}.html"
          if [ -n "$PLAN" ]; then
            FILENAME="${PLAN}-${TIMESTAMP}.html"
          fi
          cd /tmp
          git clone https://x-access-token:${{ secrets.CROSS_REPO_TOKEN }}@github.com/mantle-xyz/mantle-test-v1.git
          cd mantle-test-v1
          mkdir -p reports/proxyd/
          cp ${{ github.workspace }}/testkit/chainregression/proxyd/reports/*.html reports/proxyd/${FILENAME} 2>/dev/null || echo "No HTML report"
          git add reports/
          git diff --cached --quiet || {
            git config user.name "github-actions[bot]"
            git config user.email "github-actions[bot]@users.noreply.github.com"
            git commit -m "Add proxyd report ${FILENAME}"
            git push || { git pull --rebase && git push; }
          }
```

---

## Step 3: 配置 Secrets

### mantle-test-v1 仓库（触发方）

在 https://github.com/mantle-xyz/mantle-test-v1/settings/secrets/actions 添加：

| Secret Name | 值 | 用途 |
|-------------|-----|------|
| `DISPATCH_TOKEN` | Fine-grained PAT | 触发其他仓库的 workflow |
| `L2_RPC_URL` | `https://op-geth-rpc0-...` | Mantle RPC 端点 |
| `L2_CHAIN_ID` | `1115511107` | 链 ID |
| `SEED_KEY` | `0x8ea2...` | EEST 测试充值私钥 |

**创建 DISPATCH_TOKEN：**
1. GitHub → Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Resource owner: `mantle-xyz`（或 `mantlenetworkio`）
3. Repository access: 选择需要触发的仓库
4. Permissions: **Actions → Read and write**
5. Generate → 复制 token → 存到 mantle-test-v1 的 Secret `DISPATCH_TOKEN`

### 目标仓库（被触发方）

在目标仓库的 Settings → Secrets 添加：

| Secret Name | 值 | 用途 |
|-------------|-----|------|
| `L2_RPC_URL` | 同上 | fallback（dispatch 没传时用） |
| `L2_CHAIN_ID` | 同上 | fallback |
| `CROSS_REPO_TOKEN` | PAT | 推送报告到 mantle-test-v1（可选） |

---

## Step 4: 验证

### 本地模式验证

```bash
# 1. 确保目标仓库已 clone 到上级目录
ls ../mantle-test/testkit/chainregression/proxyd/

# 2. 查看执行计划
cd mantle-test-v1/orchestrator
./bin/mantle-test plan --config=configs/localchain.yaml --modules=proxyd

# 3. 执行
L2_RPC_URL=https://... L2_CHAIN_ID=1115511107 \
  ./bin/mantle-test run --config=configs/localchain.yaml --modules=proxyd
```

### CI 模式验证

```bash
# 需要 GITHUB_TOKEN 有 repo scope
GITHUB_TOKEN=ghp_xxx \
  ./bin/mantle-test run --config=configs/qa.yaml --modules=proxyd --mode=ci

# 输出：
#   [CI] Triggering mantlenetworkio/mantle-test/proxyd-test.yaml ...
#   Waiting for workflow to complete...
#   Workflow completed: success (2m30s)
```

### 手动推送报告验证

```bash
# 上传报告
./orchestrator/scripts/upload-report.sh proxyd /path/to/report.html --plan my-test

# 推送
git add reports/ && git commit -m "Add proxyd report" && git push

# 查看：https://mantle-xyz.github.io/mantle-test-v1/
```

---

## Step 5: （可选）配置自动触发

如果希望目标仓库代码变更时**自动触发** mantle-test-v1 的 CI：

在目标仓库加一个 webhook workflow：

```yaml
# mantlenetworkio/mantle-test/.github/workflows/trigger-mantle-test.yaml
name: Trigger Mantle Test

on:
  push:
    branches: [main]
    paths: ['testkit/chainregression/proxyd/**']  # 只有 proxyd 代码变了才触发

jobs:
  trigger:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger mantle-test-v1
        uses: peter-evans/repository-dispatch@v3
        with:
          token: ${{ secrets.CROSS_REPO_TOKEN }}
          repository: mantle-xyz/mantle-test-v1
          event-type: upstream-change
          client-payload: '{"source_repo": "mantle-test", "module": "proxyd"}'
```

---

## 完整 Checklist

| # | 步骤 | 做了？ |
|---|------|--------|
| 1 | 写 `orchestrator/modules/<name>.yaml` | 🔲 |
| 2 | 目标仓库加 workflow（支持 `repository_dispatch`） | 🔲 |
| 3 | mantle-test-v1 配 `DISPATCH_TOKEN` Secret | 🔲 |
| 4 | 目标仓库配 `L2_RPC_URL` 等 Secrets | 🔲 |
| 5 | 目标仓库配 `CROSS_REPO_TOKEN`（推送报告用） | 🔲 |
| 6 | 本地模式 `--mode=local` 验证通过 | 🔲 |
| 7 | CI 模式 `--mode=ci` 验证通过 | 🔲 |
| 8 | 报告能在 GitHub Pages 上看到 | 🔲 |
| 9 | （可选）目标仓库配自动触发 workflow | 🔲 |
