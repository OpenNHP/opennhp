#!/usr/bin/env bash
# check-plugin-deps.sh
#
# Verifies that every plugin module under examples/server_plugin/*/ pins the
# same version of any shared third-party dependency that endpoints/go.mod
# uses. Go's plugin loader fingerprints every shared package, so a single
# mismatched indirect dep (e.g. github.com/pelletier/go-toml/v2) causes
# `plugin.Open` to fail at runtime with:
#
#   plugin was built with a different version of package <pkg>
#
# Run this in CI (and locally before `make plugins`) to catch drift early.
#
# Exit codes:
#   0 - all plugin modules aligned with endpoints
#   1 - drift detected (details printed)
#   2 - usage / environment error
#
# Portable across bash 3.2 (macOS default) and bash 4+ — uses tempfile
# lookup instead of associative arrays.

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

ENDPOINTS_GOMOD="$REPO_ROOT/endpoints/go.mod"
PLUGINS_DIR="$REPO_ROOT/examples/server_plugin"

if [ ! -f "$ENDPOINTS_GOMOD" ]; then
    echo "error: $ENDPOINTS_GOMOD not found" >&2
    exit 2
fi

TMPDIR_WORK="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_WORK"' EXIT

ENDPOINTS_PAIRS="$TMPDIR_WORK/endpoints.txt"

# Extract "module version" pairs from a go.mod file. Handles both forms:
#   require foo/bar v1.2.3
#   require ( foo/bar v1.2.3 ... )
# Skips comments, blank lines, and the module's own self-reference.
extract_pairs() {
    awk '
        /^require[[:space:]]*\(/ { in_block = 1; next }
        in_block && /^\)/        { in_block = 0; next }

        {
            line = $0
            sub(/\/\/.*$/, "", line)
            sub(/^[[:space:]]+/, "", line)
            sub(/[[:space:]]+$/, "", line)
        }

        in_block && line != "" {
            n = split(line, parts, /[[:space:]]+/)
            if (n >= 2) print parts[1] " " parts[2]
            next
        }

        /^require[[:space:]]/ {
            n = split(line, parts, /[[:space:]]+/)
            if (n >= 3) print parts[2] " " parts[3]
        }
    ' "$1"
}

extract_pairs "$ENDPOINTS_GOMOD" | sort -u > "$ENDPOINTS_PAIRS"

if [ ! -s "$ENDPOINTS_PAIRS" ]; then
    echo "error: no require entries parsed from $ENDPOINTS_GOMOD" >&2
    exit 2
fi

# Look up the endpoints version for a given package name.
# Prints version (or empty) on stdout.
lookup_endpoint_version() {
    awk -v pkg="$1" '$1 == pkg { print $2; exit }' "$ENDPOINTS_PAIRS"
}

drift_found=0
plugin_count=0

for plugin_gomod in "$PLUGINS_DIR"/*/go.mod; do
    [ -f "$plugin_gomod" ] || continue
    plugin_count=$((plugin_count + 1))
    plugin_name="$(basename "$(dirname "$plugin_gomod")")"
    plugin_drift_file="$TMPDIR_WORK/$plugin_name.drift"
    : > "$plugin_drift_file"

    extract_pairs "$plugin_gomod" | while read -r pkg ver; do
        [ -z "$pkg" ] && continue
        case "$pkg" in
            github.com/OpenNHP/opennhp/*) continue ;;
        esac

        endpoint_ver="$(lookup_endpoint_version "$pkg")"
        [ -z "$endpoint_ver" ] && continue

        if [ "$ver" != "$endpoint_ver" ]; then
            echo "$pkg: plugin=$ver  endpoints=$endpoint_ver" >> "$plugin_drift_file"
        fi
    done

    drift_count="$(wc -l < "$plugin_drift_file" | tr -d ' ')"
    if [ "$drift_count" -gt 0 ]; then
        drift_found=1
        echo "[FAIL] $plugin_name has $drift_count dependency mismatch(es):"
        sed 's/^/       /' "$plugin_drift_file"
    else
        echo "[OK]   $plugin_name aligned with endpoints"
    fi
done

if [ "$plugin_count" -eq 0 ]; then
    echo "warning: no plugin modules found under $PLUGINS_DIR" >&2
fi

if [ "$drift_found" -ne 0 ]; then
    echo
    echo "Plugin dependency drift detected. Go plugin fingerprinting requires"
    echo "every shared package version to match between endpoints/go.mod and"
    echo "each plugin's go.mod. Update the plugin go.mod entries above to"
    echo "match endpoints, then re-run this script."
    exit 1
fi

echo
echo "All $plugin_count plugin module(s) aligned with endpoints/go.mod."
