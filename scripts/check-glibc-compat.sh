#!/usr/bin/env bash
# check-glibc-compat.sh <max-glibc-version> <elf-file>...
#
# Refuses any ELF file that needs a glibc symbol version newer than
# <max-glibc-version>, i.e. a binary that will not exec on a host whose glibc
# is <max-glibc-version>.
#
# The demo deploy pipeline builds the daemons with cgo (nhp-serverd has to:
# Go's plugin package is a stub without it) and ships them to Amazon Linux
# 2023, glibc 2.34. A cgo binary built against a newer glibc records versioned
# symbol references that the older runtime loader cannot satisfy, and the only
# sign of it is the daemon dying at exec:
#
#   /lib64/libc.so.6: version `GLIBC_2.38' not found (required by .../nhp-serverd)
#
# by which point the deploy has already stopped the daemon that was working.
# This check is the gate that keeps that failure in CI. It is deliberately a
# check on the artifact rather than on how it was produced, so it still holds
# if the build is moved to a different runner, image or toolchain.
#
# Exit codes:
#   0 - every file is within the limit
#   1 - at least one file needs a newer glibc
#   2 - usage / environment error

set -eu

usage() {
    echo "usage: $0 <max-glibc-version> <elf-file>..." >&2
    echo "  e.g. $0 2.34 release/nhp-server/nhp-serverd" >&2
}

if [ "$#" -lt 2 ]; then
    usage
    exit 2
fi

MAX_VERSION="$1"
shift

case "$MAX_VERSION" in
    [0-9]*.[0-9]*) ;;
    *)
        echo "error: '$MAX_VERSION' is not a glibc version (expected e.g. 2.34)" >&2
        exit 2
        ;;
esac

if ! command -v readelf >/dev/null 2>&1; then
    echo "error: readelf not found (install binutils)" >&2
    exit 2
fi

# Highest of two dotted versions, by version sort.
highest_of() {
    printf '%s\n%s\n' "$1" "$2" | sort -V | tail -1
}

FAILED=0

for file in "$@"; do
    if [ ! -f "$file" ]; then
        echo "error: $file: not found" >&2
        exit 2
    fi

    # Insist the argument really is an ELF object. Without this a path typo,
    # or a build step that wrote a shell wrapper where a binary was expected,
    # reads as "no versioned glibc dependency" and the gate waves it through.
    if ! readelf -h "$file" >/dev/null 2>&1; then
        echo "error: $file: not an ELF object" >&2
        exit 2
    fi

    # The version *needs* section is the right question: it lists the symbol
    # versions this object requires from the libraries it is linked against,
    # which is exactly what the loader resolves at exec. GLIBC_PRIVATE and
    # GLIBCXX_* are excluded by the pattern - neither carries a glibc release
    # number that can be compared.
    needed=$(readelf --version-info "$file" 2>/dev/null \
        | grep -oE 'GLIBC_[0-9]+(\.[0-9]+)+' \
        | sed 's/^GLIBC_//' \
        | sort -u -V || true)

    if [ -z "$needed" ]; then
        # Statically linked, or CGO_ENABLED=0 - nothing for the loader to
        # satisfy, so it runs on any glibc.
        echo "ok      $file (static / no versioned glibc dependency)"
        continue
    fi

    max_needed=$(printf '%s\n' "$needed" | tail -1)

    if [ "$(highest_of "$max_needed" "$MAX_VERSION")" = "$MAX_VERSION" ]; then
        echo "ok      $file (needs glibc <= $max_needed)"
    else
        echo "FAIL    $file (needs glibc $max_needed, target has $MAX_VERSION)" >&2
        echo "        versions required: $(printf '%s' "$needed" | tr '\n' ' ')" >&2
        FAILED=1
    fi
done

if [ "$FAILED" -ne 0 ]; then
    echo "" >&2
    echo "At least one binary needs a newer glibc than the deploy target has." >&2
    echo "A cgo build only runs on a glibc at least as new as the one it was" >&2
    echo "built against, so this has to be fixed by building on (or in a" >&2
    echo "container of) the target distribution - see" >&2
    echo "docker/Dockerfile.al2023-builder - not by relaxing this limit." >&2
    exit 1
fi

echo "All binaries run on glibc $MAX_VERSION."
