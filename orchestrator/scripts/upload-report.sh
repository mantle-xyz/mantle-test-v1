#!/usr/bin/env bash
# Upload a test report to mantle-test-v1 reports directory.
#
# Structure: reports/<module>/<plan-name>-<timestamp>.html
#
# Usage:
#   With plan name:
#     ./orchestrator/scripts/upload-report.sh <module> <report-file> --plan <plan-name>
#
#   Without plan name (manual upload):
#     ./orchestrator/scripts/upload-report.sh <module> <report-file>
#
# Examples:
#   ./orchestrator/scripts/upload-report.sh eest report.html --plan arsia-upgrade
#   ./orchestrator/scripts/upload-report.sh op-acceptance results.html --plan daily-qa
#   ./orchestrator/scripts/upload-report.sh proxyd report.html

set -euo pipefail

if [ $# -lt 2 ]; then
    echo "Usage: $0 <module> <report-file> [--plan <plan-name>]"
    echo ""
    echo "Examples:"
    echo "  $0 eest ./report.html --plan arsia-upgrade"
    echo "  $0 op-acceptance ./results.html --plan daily-qa"
    echo "  $0 proxyd ./report.html"
    exit 1
fi

MODULE="$1"
REPORT_FILE="$2"
PLAN_NAME=""

shift 2
while [ $# -gt 0 ]; do
    case "$1" in
        --plan) PLAN_NAME="$2"; shift 2 ;;
        *) shift ;;
    esac
done

if [ ! -f "$REPORT_FILE" ]; then
    echo "ERROR: $REPORT_FILE not found"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/../.."
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
TARGET_DIR="${PROJECT_ROOT}/reports/${MODULE}"
mkdir -p "${TARGET_DIR}"

EXT="${REPORT_FILE##*.}"
if [ -n "$PLAN_NAME" ]; then
    FILENAME="${PLAN_NAME}-${TIMESTAMP}.${EXT}"
else
    FILENAME="${TIMESTAMP}.${EXT}"
fi

cp "$REPORT_FILE" "${TARGET_DIR}/${FILENAME}"
echo "  → reports/${MODULE}/${FILENAME}"

echo ""
echo "To publish:"
echo "  cd ${PROJECT_ROOT}"
echo "  git add reports/ && git commit -m 'Add ${MODULE} report' && git push"
