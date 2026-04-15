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
#   With auto git push (stage + commit + push to remote):
#     ./orchestrator/scripts/upload-report.sh <module> <report-file> [--plan <name>] --push
#
# Examples:
#   ./orchestrator/scripts/upload-report.sh eest report.html --plan arsia-upgrade
#   ./orchestrator/scripts/upload-report.sh op-acceptance results.html --plan daily-qa --push
#   ./orchestrator/scripts/upload-report.sh proxyd report.html

set -euo pipefail

if [ $# -lt 2 ]; then
    echo "Usage: $0 <module> <report-file> [--plan <plan-name>] [--push]"
    echo ""
    echo "Examples:"
    echo "  $0 eest ./report.html --plan arsia-upgrade"
    echo "  $0 op-acceptance ./results.html --plan daily-qa --push"
    echo "  $0 proxyd ./report.html"
    echo ""
    echo "Flags:"
    echo "  --plan <name>   Prefix filename with plan name (reports/<module>/<plan>-<ts>.html)"
    echo "  --push          After copy, auto git add/commit/push to remote (publishes via GitHub Pages)"
    exit 1
fi

MODULE="$1"
REPORT_FILE="$2"
PLAN_NAME=""
DO_PUSH=false

shift 2
while [ $# -gt 0 ]; do
    case "$1" in
        --plan) PLAN_NAME="$2"; shift 2 ;;
        --push) DO_PUSH=true; shift ;;
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

TARGET_PATH="${TARGET_DIR}/${FILENAME}"
cp "$REPORT_FILE" "$TARGET_PATH"
echo "  → reports/${MODULE}/${FILENAME}"

if [ "$DO_PUSH" = true ]; then
    echo ""
    echo "Publishing via git..."
    cd "$PROJECT_ROOT"

    if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        echo "ERROR: ${PROJECT_ROOT} is not a git repository; cannot --push"
        exit 1
    fi

    COMMIT_MSG="Add ${MODULE} report"
    if [ -n "$PLAN_NAME" ]; then
        COMMIT_MSG="Add ${MODULE} report (${PLAN_NAME})"
    fi

    git add "reports/${MODULE}/${FILENAME}"
    git commit -m "$COMMIT_MSG"
    git push
    echo "  ✓ Pushed: ${COMMIT_MSG}"
else
    echo ""
    echo "To publish:"
    echo "  cd ${PROJECT_ROOT}"
    echo "  git add reports/ && git commit -m 'Add ${MODULE} report' && git push"
    echo ""
    echo "Or re-run with --push to auto-publish."
fi
