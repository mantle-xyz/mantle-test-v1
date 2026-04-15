#!/usr/bin/env bash
# Upload a test report to mantle-test-v1 reports directory.
#
# Structure: reports/<module>/<plan-name>-<timestamp>.html
#
# When --plan <name> is used:
#   - The report filename is prefixed with <name>-
#   - If reports/plans/<name>.json does NOT exist, it is auto-created (stub plan).
#     Pass --plan-desc / --plan-env / --plan-trigger to populate plan metadata.
#   - With --push, both the report and any newly-created plan JSON are committed
#     together so the Test Plans sidebar on GitHub Pages reflects the new plan.
#
# Usage:
#   With plan name (auto-registers plan if not yet registered):
#     ./orchestrator/scripts/upload-report.sh <module> <report-file> --plan <plan-name>
#
#   Without plan name (manual upload, no plan registration):
#     ./orchestrator/scripts/upload-report.sh <module> <report-file>
#
#   With auto git push (stage + commit + push to remote):
#     ./orchestrator/scripts/upload-report.sh <module> <report-file> [--plan <name>] --push
#
# Examples:
#   ./orchestrator/scripts/upload-report.sh eest report.html --plan arsia-upgrade
#   ./orchestrator/scripts/upload-report.sh op-acceptance results.html --plan daily-qa --push
#   ./orchestrator/scripts/upload-report.sh proxyd report.html --plan proxyd \
#       --plan-desc "Proxyd chain regression" --plan-env qa --push

set -euo pipefail

if [ $# -lt 2 ]; then
    echo "Usage: $0 <module> <report-file> [--plan <plan-name>] [--push]"
    echo "              [--plan-desc <text>] [--plan-env <env>] [--plan-trigger <text>]"
    echo ""
    echo "Examples:"
    echo "  $0 eest ./report.html --plan arsia-upgrade"
    echo "  $0 op-acceptance ./results.html --plan daily-qa --push"
    echo "  $0 proxyd ./report.html --plan proxyd --plan-desc 'Proxyd regression' --push"
    echo ""
    echo "Flags:"
    echo "  --plan <name>         Prefix filename and auto-register plan if not yet registered"
    echo "                        (creates reports/plans/<name>.json with stub metadata)"
    echo "  --plan-desc <text>    Description for auto-created plan (only used on first registration)"
    echo "  --plan-env <env>      Environment for auto-created plan (e.g. qa, mainnet)"
    echo "  --plan-trigger <txt>  Trigger info for auto-created plan (e.g. 'PR #123')"
    echo "  --push                After copy, auto git add/commit/push to remote"
    echo "                        (publishes via GitHub Pages; includes new plan JSON if created)"
    exit 1
fi

MODULE="$1"
REPORT_FILE="$2"
PLAN_NAME=""
PLAN_DESC=""
PLAN_ENV=""
PLAN_TRIGGER=""
DO_PUSH=false

shift 2
while [ $# -gt 0 ]; do
    case "$1" in
        --plan) PLAN_NAME="$2"; shift 2 ;;
        --plan-desc) PLAN_DESC="$2"; shift 2 ;;
        --plan-env) PLAN_ENV="$2"; shift 2 ;;
        --plan-trigger) PLAN_TRIGGER="$2"; shift 2 ;;
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

# Auto-register plan if --plan specified and plan JSON not yet exists
PLAN_CREATED=false
PLAN_JSON_RELPATH=""
if [ -n "$PLAN_NAME" ]; then
    PLANS_DIR="${PROJECT_ROOT}/reports/plans"
    PLAN_JSON="${PLANS_DIR}/${PLAN_NAME}.json"
    PLAN_JSON_RELPATH="reports/plans/${PLAN_NAME}.json"
    if [ ! -f "$PLAN_JSON" ]; then
        mkdir -p "$PLANS_DIR"
        cat > "$PLAN_JSON" << EOF
{
  "name": "${PLAN_NAME}",
  "description": "${PLAN_DESC}",
  "created": "${TIMESTAMP}",
  "environment": "${PLAN_ENV}",
  "trigger": "${PLAN_TRIGGER}"
}
EOF
        PLAN_CREATED=true
        echo "  → ${PLAN_JSON_RELPATH} (new plan registered)"
    else
        echo "  · plan '${PLAN_NAME}' already registered — skipping"
    fi
fi

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
        if [ "$PLAN_CREATED" = true ]; then
            COMMIT_MSG="Add ${MODULE} report + register plan (${PLAN_NAME})"
        else
            COMMIT_MSG="Add ${MODULE} report (${PLAN_NAME})"
        fi
    fi

    git add "reports/${MODULE}/${FILENAME}"
    if [ "$PLAN_CREATED" = true ]; then
        git add "$PLAN_JSON_RELPATH"
    fi
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
