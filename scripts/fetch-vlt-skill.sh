#!/usr/bin/env bash
# fetch-vlt-skill.sh -- Download and install the vlt skill from GitHub.
#
# Fetches the latest vlt skill from the RamXX/vlt repository and installs
# it to the user's global Claude Code skills directory (~/.claude/skills/vlt-skill).
#
# Usage:
#   fetch-vlt-skill.sh              Install if missing, skip if present.
#   fetch-vlt-skill.sh --force      Always re-download and overwrite.

set -euo pipefail

REPO="RamXX/vlt"
BRANCH="main"
SKILL_DIR="$HOME/.claude/skills/vlt-skill"
FORCE=false

if [[ "${1:-}" == "--force" ]]; then
    FORCE=true
fi

# ---------------------------------------------------------------------------
# Skip if already installed (unless --force)
# ---------------------------------------------------------------------------
if [ -f "$SKILL_DIR/SKILL.md" ] && [ "$FORCE" = false ]; then
    echo "vlt skill already installed at $SKILL_DIR (use --force to update)"
    exit 0
fi

# ---------------------------------------------------------------------------
# Download and extract
# ---------------------------------------------------------------------------
TARBALL_URL="https://github.com/${REPO}/archive/refs/heads/${BRANCH}.tar.gz"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Fetching vlt skill from github.com/${REPO}..."
if ! curl -sSfL "$TARBALL_URL" -o "$TMP_DIR/vlt.tar.gz"; then
    echo "ERROR: Failed to download from $TARBALL_URL" >&2
    echo "       Check your internet connection or install the vlt skill manually:" >&2
    echo "       git clone https://github.com/${REPO}.git && cd vlt && make install-skill" >&2
    exit 1
fi

# Extract just the skill directory
mkdir -p "$TMP_DIR/extracted"
tar xzf "$TMP_DIR/vlt.tar.gz" --strip-components=3 -C "$TMP_DIR/extracted" "vlt-${BRANCH}/docs/vlt-skill"

# ---------------------------------------------------------------------------
# Install to global skills directory
# ---------------------------------------------------------------------------
mkdir -p "$(dirname "$SKILL_DIR")"
rm -rf "$SKILL_DIR"
mv "$TMP_DIR/extracted" "$SKILL_DIR"

echo "Installed vlt skill to $SKILL_DIR"
