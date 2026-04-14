#!/usr/bin/env bash
# Collect test reports from all sources into a unified reports directory
#
# Usage:
#   ./collect-reports.sh                    # collect to default location
#   REPORTS_DIR=/path/to/reports ./collect-reports.sh
#
# Sources:
#   1. mantle-execution-specs: EEST logs + HTML report + JSON report
#   2. mantle-execution-apis: RPC conformance results
#   3. orchestrator: suite results

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/../.."
REPORTS_DIR="${REPORTS_DIR:-${PROJECT_ROOT}/reports}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
RUN_DIR="${REPORTS_DIR}/${TIMESTAMP}"

SPECS_DIR="${SPECS_DIR:-/Users/user/space/mantle-execution-specs}"
APIS_DIR="${APIS_DIR:-/Users/user/space/mantle-execution-apis}"

echo "=== Collect Test Reports ==="
echo "Run ID:     ${TIMESTAMP}"
echo "Output:     ${RUN_DIR}"
echo "Specs repo: ${SPECS_DIR}"
echo "APIs repo:  ${APIS_DIR}"
echo ""

mkdir -p "${RUN_DIR}/eest" "${RUN_DIR}/rpc" "${RUN_DIR}/orchestrator"

# ── 1. EEST reports (mantle-execution-specs) ──
echo "Collecting EEST reports..."

# Logs
if [ -d "${SPECS_DIR}/logs" ]; then
    LATEST_LOGS=$(ls -t "${SPECS_DIR}/logs/"execute-remote-*.log 2>/dev/null | head -10)
    if [ -n "${LATEST_LOGS}" ]; then
        cp ${LATEST_LOGS} "${RUN_DIR}/eest/"
        echo "  Copied $(echo "${LATEST_LOGS}" | wc -l | tr -d ' ') log files"
    fi
fi

# HTML report
if [ -f "${SPECS_DIR}/execution_results/report_execute.html" ]; then
    cp "${SPECS_DIR}/execution_results/report_execute.html" "${RUN_DIR}/eest/"
    echo "  Copied HTML report"
fi

# JSON report (pytest-json-report)
LATEST_JSON=$(find "${SPECS_DIR}" -maxdepth 2 -name "*.json" -newer "${SPECS_DIR}/logs" 2>/dev/null | grep -i "report\|result" | head -1)
if [ -n "${LATEST_JSON}" ]; then
    cp "${LATEST_JSON}" "${RUN_DIR}/eest/"
    echo "  Copied JSON report"
fi

# ── 2. RPC conformance reports ──
echo "Collecting RPC conformance reports..."

# If rpc-conformance.sh was run, check for output
if [ -f "${APIS_DIR}/openrpc.json" ]; then
    cp "${APIS_DIR}/openrpc.json" "${RUN_DIR}/rpc/"
    echo "  Copied OpenRPC spec"
fi

# ── 3. Orchestrator reports ──
echo "Collecting orchestrator reports..."

# JSON reports from orchestrator runs
ORCH_REPORTS=$(find "${PROJECT_ROOT}/orchestrator" -name "*.json" -path "*report*" 2>/dev/null | head -5)
if [ -n "${ORCH_REPORTS}" ]; then
    cp ${ORCH_REPORTS} "${RUN_DIR}/orchestrator/"
    echo "  Copied orchestrator reports"
fi

# ── 4. Generate summary ──
echo ""
echo "Generating summary..."

SUMMARY="${RUN_DIR}/summary.md"
cat > "${SUMMARY}" << EOF
# Test Report — ${TIMESTAMP}

## Environment
- Date: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
- RPC: ${L2_RPC_URL:-not set}
- Chain ID: ${L2_CHAIN_ID:-not set}
- Fork: ${FORK_NAME:-Cancun}

## EEST Results
EOF

# Parse EEST logs for pass/fail counts
for log in "${RUN_DIR}"/eest/execute-remote-*-main.log; do
    [ -f "$log" ] || continue
    basename="$(basename "$log")"

    # Extract summary line from log
    passed=$(grep -c "✅" "$log" 2>/dev/null || echo "0")
    failed=$(grep -c "❌" "$log" 2>/dev/null || echo "0")

    # Try to get pytest summary from the original output
    # The logs don't have the summary, but we can count START/END TEST
    total_tests=0
    for wlog in "${RUN_DIR}"/eest/execute-remote-*-gw*.log "${RUN_DIR}"/eest/execute-remote-*-main.log; do
        [ -f "$wlog" ] || continue
        count=$(grep -c "END TEST" "$wlog" 2>/dev/null || echo "0")
        total_tests=$((total_tests + count))
    done

    cat >> "${SUMMARY}" << EOF

### $(echo "$basename" | sed 's/execute-remote-//' | sed 's/-main.log//')
- Total tests executed: ~${total_tests}
- Passed: ${passed}
- Failed: ${failed}
- Log: \`eest/${basename}\`
EOF
    break  # Only process latest
done

cat >> "${SUMMARY}" << EOF

## Files
EOF

# List all collected files
for f in $(find "${RUN_DIR}" -type f | sort); do
    rel=$(echo "$f" | sed "s|${RUN_DIR}/||")
    size=$(ls -lh "$f" | awk '{print $5}')
    echo "- \`${rel}\` (${size})" >> "${SUMMARY}"
done

echo ""
echo "=== Report collected ==="
echo "Location: ${RUN_DIR}"
echo "Summary:  ${RUN_DIR}/summary.md"
echo ""
echo "Files:"
find "${RUN_DIR}" -type f | sort | while read f; do
    rel=$(echo "$f" | sed "s|${RUN_DIR}/||")
    size=$(ls -lh "$f" | awk '{print $5}')
    echo "  ${rel} (${size})"
done
