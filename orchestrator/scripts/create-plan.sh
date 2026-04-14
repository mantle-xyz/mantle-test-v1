#!/usr/bin/env bash
# Create a test plan that groups multiple module reports together.
#
# Usage:
#   ./orchestrator/scripts/create-plan.sh <plan-name> <module1:report1> <module2:report2> ...
#
# Examples:
#   # Arsia upgrade: link existing reports
#   ./orchestrator/scripts/create-plan.sh "Arsia Upgrade" eest:eest/20260414-150001.html op-acceptance:op-acceptance/20260414-150005.html
#
#   # Upload + create plan in one go
#   ./orchestrator/scripts/upload-report.sh eest report.html
#   ./orchestrator/scripts/upload-report.sh op-acceptance results.html
#   ./orchestrator/scripts/create-plan.sh "Daily QA" eest:eest/$(ls -t reports/eest/ | head -1) op-acceptance:op-acceptance/$(ls -t reports/op-acceptance/ | head -1)

set -euo pipefail

if [ $# -lt 2 ]; then
    echo "Usage: $0 <plan-name> <module:report-path> [module:report-path ...]"
    echo ""
    echo "Example:"
    echo "  $0 'Arsia Upgrade' eest:eest/20260414.html op-acceptance:op-acceptance/20260414.html"
    exit 1
fi

PLAN_NAME="$1"
shift

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/../.."
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
PLANS_DIR="${PROJECT_ROOT}/reports/plans"
mkdir -p "${PLANS_DIR}"

# Build JSON
REPORTS_JSON="["
FIRST=true
for entry in "$@"; do
    MODULE="${entry%%:*}"
    FILE="${entry#*:}"
    if [ "$FIRST" = true ]; then
        FIRST=false
    else
        REPORTS_JSON="${REPORTS_JSON},"
    fi
    REPORTS_JSON="${REPORTS_JSON}{\"module\":\"${MODULE}\",\"file\":\"${FILE}\"}"
done
REPORTS_JSON="${REPORTS_JSON}]"

PLAN_FILE="${PLANS_DIR}/${TIMESTAMP}.json"
cat > "${PLAN_FILE}" << EOF
{
  "name": "${PLAN_NAME}",
  "timestamp": "${TIMESTAMP}",
  "reports": ${REPORTS_JSON}
}
EOF

echo "Created plan: reports/plans/${TIMESTAMP}.json"
echo "  Name: ${PLAN_NAME}"
echo "  Reports:"
for entry in "$@"; do
    echo "    - ${entry}"
done
echo ""
echo "To publish:"
echo "  git add reports/ && git commit -m 'Add plan: ${PLAN_NAME}' && git push"
