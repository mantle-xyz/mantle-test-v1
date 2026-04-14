#!/usr/bin/env bash
# Upload test reports to mantle-test-v1 reports directory.
#
# Structure: reports/<module>/<timestamp>.html
#
# Usage:
#   Single file:
#     ./scripts/upload-report.sh <module> <report-file>
#
#   Multiple files (e.g., op-acceptance with many testruns):
#     ./scripts/upload-report.sh <module> <file1> <file2> ...
#     ./scripts/upload-report.sh <module> path/to/testrun-*/results.html
#
#   Latest only:
#     ./scripts/upload-report.sh <module> --latest <glob-pattern>
#
# Examples:
#   ./scripts/upload-report.sh eest execution_results/report_execute.html
#   ./scripts/upload-report.sh op-acceptance logs/testrun-*/results.html
#   ./scripts/upload-report.sh op-acceptance --latest "logs/testrun-*/results.html"

set -euo pipefail

if [ $# -lt 2 ]; then
    echo "Usage:"
    echo "  $0 <module> <report-file> [file2 ...]"
    echo "  $0 <module> --latest <glob-pattern>"
    echo ""
    echo "Examples:"
    echo "  $0 eest execution_results/report_execute.html"
    echo "  $0 op-acceptance logs/testrun-*/results.html"
    echo "  $0 op-acceptance --latest 'logs/testrun-*/results.html'"
    exit 1
fi

MODULE="$1"
shift

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/../.."
TARGET_DIR="${PROJECT_ROOT}/reports/${MODULE}"
mkdir -p "${TARGET_DIR}"

UPLOADED=0

if [ "$1" = "--latest" ]; then
    # Only upload the most recent file matching the pattern
    shift
    PATTERN="$1"
    LATEST=$(ls -t $PATTERN 2>/dev/null | head -1)
    if [ -z "$LATEST" ]; then
        echo "ERROR: No files matching $PATTERN"
        exit 1
    fi
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    EXT="${LATEST##*.}"
    cp "$LATEST" "${TARGET_DIR}/${TIMESTAMP}.${EXT}"
    echo "  → reports/${MODULE}/${TIMESTAMP}.${EXT} (latest of pattern)"
    UPLOADED=1
else
    # Upload all specified files, each with its own timestamp
    for f in "$@"; do
        if [ ! -f "$f" ]; then
            echo "  Skip: $f (not found)"
            continue
        fi
        # Use file modification time as timestamp (preserve original time)
        FILE_TIME=$(date -r "$f" +%Y%m%d-%H%M%S 2>/dev/null || date +%Y%m%d-%H%M%S)
        EXT="${f##*.}"
        DEST="${TARGET_DIR}/${FILE_TIME}.${EXT}"
        # Avoid overwriting if same timestamp exists
        if [ -f "$DEST" ]; then
            FILE_TIME="${FILE_TIME}-$(( RANDOM % 1000 ))"
            DEST="${TARGET_DIR}/${FILE_TIME}.${EXT}"
        fi
        cp "$f" "$DEST"
        echo "  → reports/${MODULE}/${FILE_TIME}.${EXT}"
        UPLOADED=$((UPLOADED + 1))
    done
fi

echo ""
echo "Uploaded ${UPLOADED} report(s) to reports/${MODULE}/"
echo ""
echo "To publish:"
echo "  cd ${PROJECT_ROOT}"
echo "  git add reports/ && git commit -m 'Add ${MODULE} reports' && git push"
