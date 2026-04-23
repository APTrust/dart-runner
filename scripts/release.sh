#!/usr/bin/env bash
#
# release.sh - Build and release dart-runner binaries to S3.
#
# Usage:
#   release.sh --pre-check <version>   Run pre-release checks only.
#   release.sh <version>               Build and release binaries.
#
# Example:
#   release.sh --pre-check v1.0.4
#   release.sh v1.0.4
#
# ----------------------------------------------------------------------------

set -euo pipefail

# ---------------------------------------------------------------------------
# Determine script and project directories
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
die() {
    echo "ERROR: $*" >&2
    exit 1
}

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
PRE_CHECK=false
VERSION=""

if [[ $# -eq 0 ]]; then
    die "Usage: release.sh [--pre-check] <version>  (e.g. release.sh v1.0.4)"
fi

if [[ "$1" == "--pre-check" ]]; then
    PRE_CHECK=true
    shift
fi

if [[ $# -eq 0 ]]; then
    die "Usage: release.sh [--pre-check] <version>  (e.g. release.sh v1.0.4)"
fi

VERSION="$1"

# ---------------------------------------------------------------------------
# Check required environment variables
# ---------------------------------------------------------------------------
check_env() {
    [[ -n "${AWS_ACCESS_KEY_ID:-}"     ]] || die "AWS_ACCESS_KEY_ID is not set in the environment."
    [[ -n "${AWS_SECRET_ACCESS_KEY:-}" ]] || die "AWS_SECRET_ACCESS_KEY is not set in the environment."
}

# ---------------------------------------------------------------------------
# Check that the current git tag matches the requested version
# ---------------------------------------------------------------------------
check_git_tag() {
    local current_tag
    current_tag="$(git -C "$PROJECT_DIR" describe --tags --exact-match HEAD 2>/dev/null || true)"
    if [[ "$current_tag" != "$VERSION" ]]; then
        die "Current git tag ('${current_tag:-<none>}') does not match the requested version '$VERSION'. " \
            "Please create or check out tag '$VERSION' before releasing."
    fi
}

# ---------------------------------------------------------------------------
# Verify changelog.md contains an h2 entry for this version with a date
# ---------------------------------------------------------------------------
check_changelog() {
    local changelog_file="$PROJECT_DIR/changelog.md"
    [[ -f "$changelog_file" ]] || die "changelog.md not found in $PROJECT_DIR."

    # Match the first ## headline that contains the version string.
    # The headline may look like "## v1.0.4 - 2026-03-24" or "## v1.0.4 - March 24, 2026".
    local match
    local version_no_v="${VERSION#v}"
    match="$(grep -m1 "^## .*${version_no_v}" "$changelog_file" || true)"

    if [[ -z "$match" ]]; then
        die "changelog.md does not contain an h2 (##) section for '$VERSION'. " \
            "Please add a changelog entry and release date for this version before releasing."
    fi

    # Check that the matched line contains a date.
    # Accept ISO dates (2026-03-24) or written dates (March 24, 2026 / Mar 24, 2026).
    if ! echo "$match" | grep -qE '([0-9]{4}-[0-9]{2}-[0-9]{2}|[A-Za-z]+ [0-9]{1,2},? [0-9]{4})'; then
        die "The '## $VERSION' headline in changelog.md does not include a release date. " \
            "Please add a date (e.g. '2026-03-24' or 'March 24, 2026') to the headline before releasing."
    fi

    echo "changelog.md: found entry for '$VERSION' with a release date."
}

# ---------------------------------------------------------------------------
# Pre-check mode
# ---------------------------------------------------------------------------
run_pre_check() {
    echo "Running pre-release checks for version '$VERSION'..."
    check_changelog
    check_git_tag
    check_env
    echo "All pre-release checks passed."
}

# ---------------------------------------------------------------------------
# Full release
# ---------------------------------------------------------------------------
run_release() {
    # 1. Verify git tag and changelog
    check_git_tag
    check_changelog

    # 2. Build binaries for all platforms
    echo "Building dart-runner binaries for all platforms..."
    cd "$PROJECT_DIR"
    bash "$SCRIPT_DIR/build_dart_runner.sh"

    # 3. Create S3 folders
    echo "Creating S3 folders for version '$VERSION'..."
    cd "$SCRIPT_DIR"
    go run s3_helper.go -make-folders -version "$VERSION" || \
        die "Failed to create S3 folders for version '$VERSION'."

    # 4. Upload binaries
    # Parallel arrays: UPLOAD_FILES[i] maps to UPLOAD_ARCHES[i]
    UPLOAD_FILES=(
        "$PROJECT_DIR/dist/linux-x64/dart-runner"
        "$PROJECT_DIR/dist/linux-arm64/dart-runner"
        "$PROJECT_DIR/dist/mac-x64/dart-runner"
        "$PROJECT_DIR/dist/mac-arm64/dart-runner"
        "$PROJECT_DIR/dist/windows-x64/dart-runner.exe"
        "$PROJECT_DIR/dist/windows-arm64/dart-runner.exe"
    )
    UPLOAD_ARCHES=(
        "linux/amd64"
        "linux/arm64"
        "mac/amd64"
        "mac/arm64"
        "windows/amd64"
        "windows/arm64"
    )

    for i in "${!UPLOAD_FILES[@]}"; do
        filepath="${UPLOAD_FILES[$i]}"
        arch="${UPLOAD_ARCHES[$i]}"
        [[ -f "$filepath" ]] || die "Expected binary not found at: $filepath"
        echo "Uploading $filepath (arch: $arch)..."
        cd "$SCRIPT_DIR"
        go run s3_helper.go -upload -version "$VERSION" -arch "$arch" "$filepath" || \
            die "Failed to upload $filepath to S3."
    done

    # 5. Print download links
    echo ""
    echo "Download links for version '$VERSION':"
    cd "$SCRIPT_DIR"
    go run s3_helper.go -get-links "$VERSION" || \
        die "Failed to retrieve download links for version '$VERSION'."

    # 6. Print post-release checklist
    cat <<EOF

Checklist:

1. Ensure builds for all platforms (linux/amd64, linux/arm64, mac/amd64, mac/arm64,
   windows/amd64, windows/arm64) are present in aptrust.public.download/dart-runner/${VERSION}/.
2. Using the download links above, update the dart-runner version number, release date,
   links, and SHA256 checksums in https://aptrust.github.io/dart-docs/runner/#downloads
3. After updating dart-docs, run \`mkdocs gh-deploy\` to publish the updates.
4. Push any README changes to GitHub.
EOF
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
if [[ "$PRE_CHECK" == "true" ]]; then
    run_pre_check
else
    check_env
    run_release
fi
