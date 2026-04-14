#!/usr/bin/env bash
# Create a test plan metadata file.
#
# The plan name is used as prefix for all module report filenames:
#   reports/<module>/<plan-name>-<timestamp>.html
#
# This script only creates the plan metadata JSON. Use upload-report.sh --plan <name>
# to upload reports with the matching plan prefix.
#
# Usage:
#   ./orchestrator/scripts/create-plan.sh <plan-name> [--desc "description"] [--env qa] [--trigger "PR #123"]
#
# Examples:
#   ./orchestrator/scripts/create-plan.sh arsia-upgrade --desc "Arsia 升级验收" --env qa
#   ./orchestrator/scripts/create-plan.sh daily-qa --desc "Daily QA regression"

set -euo pipefail

if [ $# -lt 1 ]; then
    echo "Usage: $0 <plan-name> [--desc 'description'] [--env environment] [--trigger 'trigger info']"
    exit 1
fi

PLAN_NAME="$1"
shift
DESC=""
ENV=""
TRIGGER=""

while [ $# -gt 0 ]; do
    case "$1" in
        --desc) DESC="$2"; shift 2 ;;
        --env) ENV="$2"; shift 2 ;;
        --trigger) TRIGGER="$2"; shift 2 ;;
        *) shift ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/../.."
PLANS_DIR="${PROJECT_ROOT}/reports/plans"
mkdir -p "${PLANS_DIR}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

cat > "${PLANS_DIR}/${PLAN_NAME}.json" << EOF
{
  "name": "${PLAN_NAME}",
  "description": "${DESC}",
  "created": "${TIMESTAMP}",
  "environment": "${ENV}",
  "trigger": "${TRIGGER}"
}
EOF

echo "Created plan: reports/plans/${PLAN_NAME}.json"
echo ""
echo "Now upload reports with this plan name:"
echo "  ./orchestrator/scripts/upload-report.sh eest report.html --plan ${PLAN_NAME}"
echo "  ./orchestrator/scripts/upload-report.sh op-acceptance results.html --plan ${PLAN_NAME}"
echo ""
echo "Then push:"
echo "  git add reports/ && git commit -m 'Add plan: ${PLAN_NAME}' && git push"
