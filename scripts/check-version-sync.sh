#!/usr/bin/env bash
# check-version-sync.sh
#
# Verifies that the version number the public pages show is the version that
# was actually released.
#
# The repository carries two version axes:
#
#   A (release)  .release-please-manifest.json   — bumped automatically by
#                  release-please on every release.
#   B (build)    nhp/version/VERSION             — read by the Makefile (Go
#                  binaries), by endpoints/js-agent/vite.config.ts (the SDK's
#                  __SDK_VERSION__ define) and by docker/Dockerfile.demoapp.
#                  The public pages (agent.opennhp.org, reg.opennhp.org,
#                  demo.opennhp.org) render B.
#
# Historically only A was bumped, so B drifted (it stayed at 0.8.0 across the
# 1.0.0 and 1.1.0 releases) and the pages advertised a stale SDK. release-please
# now updates B through release-please-config.json's "version-file"; this script
# is the safety net that turns a future drift into a red CI run instead of a
# user-visible stale version.
#
# Exit codes:
#   0 - every version file agrees with the released version
#   1 - drift detected (details printed)
#   2 - usage / environment error
#
# Portable across bash 3.2 (macOS default) and bash 4+.

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

MANIFEST="$REPO_ROOT/.release-please-manifest.json"
RP_CONFIG="$REPO_ROOT/release-please-config.json"
VERSION_FILE="$REPO_ROOT/nhp/version/VERSION"
PKG_JSON="$REPO_ROOT/endpoints/js-agent/package.json"
PKG_LOCK="$REPO_ROOT/endpoints/js-agent/package-lock.json"

for required in "$MANIFEST" "$RP_CONFIG" "$VERSION_FILE" "$PKG_JSON" "$PKG_LOCK"; do
    if [ ! -f "$required" ]; then
        echo "error: $required not found" >&2
        exit 2
    fi
done

# A semantic version, with an optional pre-release / build suffix.
SEMVER='[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.-]+)?'

# Read the released version from the release-please manifest's "." entry.
# The manifest is a flat JSON object of "package path" -> "version".
manifest_version() {
    tr -d ' \t\n\r' < "$MANIFEST" \
        | grep -oE '"[^"]*"[[:space:]]*:[[:space:]]*"[^"]*"' \
        | grep -E '^"\."[[:space:]]*:' \
        | head -n 1 \
        | grep -oE "$SEMVER" \
        | head -n 1
}

# Read the first "version" string of a JSON file (the top-level one).
json_version() {
    grep -m1 -E '"version"[[:space:]]*:' "$1" \
        | grep -oE "$SEMVER" \
        | head -n 1
}

# Read packages[""].version out of a package-lock.json.
lock_root_version() {
    awk '
        /"packages"[[:space:]]*:/ { in_packages = 1; next }
        in_packages && /"version"[[:space:]]*:/ { print; exit }
    ' "$1" \
        | grep -oE "$SEMVER" \
        | head -n 1
}

expected="$(manifest_version)"
if [ -z "$expected" ]; then
    echo "error: could not read the released version from $MANIFEST" >&2
    exit 2
fi

# The release-please config is what keeps nhp/version/VERSION in step with the
# release, so a broken config would silently stop the sync (and the release
# pipeline itself would fail). Validate it here, before it reaches main.
#
# release-please v17 reads its config file as {"packages": {"<path>": {...}}}
# and ignores the whole file when that "packages" key is missing.
config_error=0

if command -v python3 >/dev/null 2>&1; then
    if ! python3 -c 'import json,sys; json.load(open(sys.argv[1]))' "$RP_CONFIG" >/dev/null 2>&1; then
        echo "[FAIL] $RP_CONFIG is not valid JSON"
        config_error=1
    fi
    if ! python3 -c '
import json,sys
cfg = json.load(open(sys.argv[1]))
pkgs = cfg.get("packages")
if not isinstance(pkgs, dict) or "." not in pkgs:
    raise SystemExit(1)
pkg = pkgs["."]
if not pkg.get("release-type"):
    raise SystemExit(1)
if pkg.get("version-file") != "nhp/version/VERSION":
    raise SystemExit(1)
' "$RP_CONFIG" >/dev/null 2>&1; then
        echo "[FAIL] $RP_CONFIG must declare packages.\".\" with a release-type and version-file nhp/version/VERSION"
        config_error=1
    fi
else
    # No JSON parser available: fall back to a structural check.
    if ! grep -q '"packages"' "$RP_CONFIG"; then
        echo "[FAIL] $RP_CONFIG has no \"packages\" key (release-please v17 ignores the file without it)"
        config_error=1
    fi
    if ! grep -q '"release-type"' "$RP_CONFIG"; then
        echo "[FAIL] $RP_CONFIG has no release-type"
        config_error=1
    fi
    if ! grep -q 'nhp/version/VERSION' "$RP_CONFIG"; then
        echo "[FAIL] $RP_CONFIG does not reference nhp/version/VERSION"
        config_error=1
    fi
fi

if [ "$config_error" -ne 0 ]; then
    echo
    echo "The release-please config is broken, so the next release would either"
    echo "fail or leave nhp/version/VERSION behind. Fix $RP_CONFIG."
    exit 1
fi
echo "[OK]   release-please-config.json is valid and keeps nhp/version/VERSION in sync"

actual_version="$(tr -d '[:space:]' < "$VERSION_FILE")"
pkg_version="$(json_version "$PKG_JSON")"
lock_version="$(json_version "$PKG_LOCK")"
lock_root="$(lock_root_version "$PKG_LOCK")"

drift_found=0

fail() {
    echo "[FAIL] $1: expected '$expected', found '$2'"
    drift_found=1
}

if [ "$actual_version" != "$expected" ]; then
    fail "nhp/version/VERSION" "$actual_version"
else
    echo "[OK]   nhp/version/VERSION = $actual_version"
fi

if [ "$pkg_version" != "$expected" ]; then
    fail "endpoints/js-agent/package.json" "$pkg_version"
else
    echo "[OK]   endpoints/js-agent/package.json = $pkg_version"
fi

# The lockfile's version fields are npm bookkeeping only — npm resolves the
# root package from the package.json on disk, never from the lockfile — so a
# stale value there breaks nothing. Warn, but do not fail.
if [ "$lock_version" != "$expected" ]; then
    echo "[WARN] endpoints/js-agent/package-lock.json (top level): expected '$expected', found '$lock_version'"
fi
if [ "$lock_root" != "$expected" ]; then
    echo "[WARN] endpoints/js-agent/package-lock.json (packages root): expected '$expected', found '$lock_root'"
fi

if [ "$drift_found" -ne 0 ]; then
    echo
    echo "Version drift detected. The public pages render the version from"
    echo "nhp/version/VERSION, so a drifted file makes them advertise a stale"
    echo "SDK version. Set nhp/version/VERSION and endpoints/js-agent/package.json"
    echo "to '$expected', or run the release-please release PR, then re-run this script."
    exit 1
fi

echo
echo "All version files are in sync with the released version ($expected)."
