#!/usr/bin/env bash
# Upload a test report to mantle-test-v1 reports directory.
#
# Structure:
#   reports/<module>/<env>/<plan>-<timestamp>.html         (with --env)
#   reports/<module>/<plan>-<timestamp>.html               (without --env, legacy)
#
# When --plan <name> is used:
#   - The report filename is prefixed with <name>-
#   - If reports/plans/<name>.json does NOT exist, it is auto-created (stub plan).
#     Pass --plan-desc / --plan-env / --plan-trigger to populate plan metadata.
#     Note: --plan-env sets plan's default env metadata; --env controls storage path.
#   - With --push, both the report and any newly-created plan JSON are committed
#     together so the Test Plans sidebar on GitHub Pages reflects the new plan.
#
# Usage:
#   With env + plan (recommended):
#     ./orchestrator/scripts/upload-report.sh <module> <report-file> --env <env> --plan <plan-name>
#
#   Without env (legacy, stored at reports/<module>/ root):
#     ./orchestrator/scripts/upload-report.sh <module> <report-file> --plan <plan-name>
#
#   With auto git push (stage + commit + push to remote):
#     ./orchestrator/scripts/upload-report.sh <module> <report-file> [--env <env>] [--plan <name>] --push
#
# Examples:
#   ./orchestrator/scripts/upload-report.sh proxyd report.html --env qa3 --plan proxyd \
#       --plan-desc "Proxyd chain regression" --push
#   ./orchestrator/scripts/upload-report.sh eest report.html --env sepolia --plan arsia-upgrade

set -euo pipefail

if [ $# -lt 2 ]; then
    echo "Usage: $0 <module> <report-file> [--env <env>] [--plan <plan-name>] [--push]"
    echo "              [--plan-desc <text>] [--plan-env <env>] [--plan-trigger <text>]"
    echo ""
    echo "Examples:"
    echo "  $0 proxyd ./report.html --env qa3 --plan proxyd --push"
    echo "  $0 eest ./report.html --env sepolia --plan arsia-upgrade --push"
    echo "  $0 op-acceptance ./results.html --env qa --plan daily-qa --push"
    echo ""
    echo "Flags:"
    echo "  --env <env>           Environment name (qa/qa3/sepolia/mainnet/...);"
    echo "                        stored at reports/<module>/<env>/<file>. Omit for legacy flat layout."
    echo "  --plan <name>         Prefix filename and auto-register plan if not yet registered"
    echo "                        (creates reports/plans/<name>.json with stub metadata)"
    echo "  --plan-desc <text>    Description for auto-created plan (only used on first registration)"
    echo "  --plan-env <env>      Default env metadata stored in plan JSON (independent from --env)"
    echo "  --plan-trigger <txt>  Trigger info for auto-created plan (e.g. 'PR #123')"
    echo "  --push                After copy, auto git add/commit/push to remote"
    echo "                        (publishes via GitHub Pages; includes new plan JSON if created)"
    exit 1
fi

MODULE="$1"
REPORT_FILE="$2"
ENV_NAME=""
PLAN_NAME=""
PLAN_DESC=""
PLAN_ENV=""
PLAN_TRIGGER=""
DO_PUSH=false

shift 2
while [ $# -gt 0 ]; do
    case "$1" in
        --env) ENV_NAME="$2"; shift 2 ;;
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

# If --env not explicitly passed but --plan-env is, use it as storage env too.
if [ -z "$ENV_NAME" ] && [ -n "$PLAN_ENV" ]; then
    ENV_NAME="$PLAN_ENV"
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/../.."
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

if [ -n "$ENV_NAME" ]; then
    TARGET_DIR="${PROJECT_ROOT}/reports/${MODULE}/${ENV_NAME}"
    REL_DIR="reports/${MODULE}/${ENV_NAME}"
else
    TARGET_DIR="${PROJECT_ROOT}/reports/${MODULE}"
    REL_DIR="reports/${MODULE}"
fi
mkdir -p "${TARGET_DIR}"

# Preserve the original filename; only prepend <plan>- as a prefix.
# If a file of the same name already exists, append -<timestamp> to avoid overwrite.
ORIG_BASE="$(basename "$REPORT_FILE")"
ORIG_STEM="${ORIG_BASE%.*}"
EXT="${ORIG_BASE##*.}"

if [ -n "$PLAN_NAME" ]; then
    FILENAME="${PLAN_NAME}-${ORIG_STEM}.${EXT}"
else
    FILENAME="${ORIG_STEM}.${EXT}"
fi

# Dedupe if file with the same name already exists in target
if [ -e "${TARGET_DIR}/${FILENAME}" ]; then
    FILENAME="${FILENAME%.${EXT}}-${TIMESTAMP}.${EXT}"
    echo "  · name collision — appended timestamp suffix"
fi

TARGET_PATH="${TARGET_DIR}/${FILENAME}"
REL_PATH="${REL_DIR}/${FILENAME}"
cp "$REPORT_FILE" "$TARGET_PATH"
echo "  → ${REL_PATH}"

# Auto-register plan if --plan specified and plan JSON not yet exists
PLAN_CREATED=false
PLAN_JSON_RELPATH=""
if [ -n "$PLAN_NAME" ]; then
    PLANS_DIR="${PROJECT_ROOT}/reports/plans"
    PLAN_JSON="${PLANS_DIR}/${PLAN_NAME}.json"
    PLAN_JSON_RELPATH="reports/plans/${PLAN_NAME}.json"
    if [ ! -f "$PLAN_JSON" ]; then
        mkdir -p "$PLANS_DIR"
        # Prefer explicit --plan-env; fall back to --env for plan metadata.
        META_ENV="$PLAN_ENV"
        if [ -z "$META_ENV" ]; then
            META_ENV="$ENV_NAME"
        fi
        cat > "$PLAN_JSON" << EOF
{
  "name": "${PLAN_NAME}",
  "description": "${PLAN_DESC}",
  "created": "${TIMESTAMP}",
  "environment": "${META_ENV}",
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
    if [ -n "$ENV_NAME" ]; then
        COMMIT_MSG="Add ${MODULE}/${ENV_NAME} report"
    fi
    if [ -n "$PLAN_NAME" ]; then
        if [ "$PLAN_CREATED" = true ]; then
            COMMIT_MSG="${COMMIT_MSG} + register plan (${PLAN_NAME})"
        else
            COMMIT_MSG="${COMMIT_MSG} (${PLAN_NAME})"
        fi
    fi

    git add "$REL_PATH"
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
