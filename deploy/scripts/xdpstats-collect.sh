#!/bin/bash
# xdpstats-collect.sh
#
# Pulls nhp_server_xdp-YYYY-MM-DD.log files from nhp-server via SSH and
# rebuilds the stats.opennhp.org static site. Driven by a systemd timer on
# the relay host (deploy/scripts/xdpstats.timer, OnCalendar=*-*-* 00:20:00
# UTC, Persistent=true). Run by hand with no arguments to trigger a single
# pass; debug flags `-n` (no download) and `-v` (verbose) are accepted.
#
# Sources are described in /home/ec2-user/xdpstats/sources.conf, one per
# line, in the form:
#
#   <name> <ssh_user> <host> <log_dir>
#
# v1 ships with one line, "server". AC is intentionally not here: AC writes
# per-flow accept/deny files (not nhp_server_xdp-*), and the parsing lives
# in endpoints/xdplog-report/parse.go which still only knows the server's
# nhp_server_xdp log layout.
#
# The collector:
#   1. Reads sources.conf and resolves each (source) -> remote host + user.
#   2. For every source, calls `list` on the remote forced-command script
#      and parses (date|size|mtime) tuples.
#   3. For every (source, date) the local cache is missing, smaller than
#      the remote, OR has today's UTC date, runs `get <date>`, validates
#      the last line, gzips, and writes the cache atomically.
#   4. Runs `xdplog-report site` to rebuild the static site from the cache.
#   5. Prunes raw cache files older than 90 days.
#
# Concurrency: protected by a non-blocking flock at $RAW_DIR/.lock. A second
# invocation (timer + manual run racing) exits 0 without doing work.
#
# Errors: never deletes an existing cache file on failure; the systemd unit
# is marked failed and on the next tick we retry. Re-runs are idempotent:
# identical (size, mtime) days are skipped by xdplog-report site.

set -euo pipefail

# ---- paths -----------------------------------------------------------------
XDPSTATS_DIR="/home/ec2-user/xdpstats"
RAW_DIR="$XDPSTATS_DIR/raw"
WEB_DIR="$XDPSTATS_DIR/web"
SOURCES_CONF="$XDPSTATS_DIR/sources.conf"
LOG_FILE="$XDPSTATS_DIR/collect.log"
LOCK_FILE="$RAW_DIR/.lock"
RETENTION_DAYS=90
SITE_BIN="/usr/local/bin/xdplog-report"
SITE_ARGS=("-top" "20")
TODAY="$(date -u +%Y-%m-%d)"

# ---- flags ------------------------------------------------------------------
DEBUG_LIST_ONLY=0
VERBOSE=0
for arg in "$@"; do
  case "$arg" in
    -n) DEBUG_LIST_ONLY=1 ;;
    -v) VERBOSE=1 ;;
    -h|--help)
      sed -n '2,30p' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *) echo "xdpstats-collect: unknown argument: $arg" >&2; exit 64 ;;
  esac
done

log() {
  printf '%s [%s] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "${1:-info}" "${2:-}" >>"$LOG_FILE"
}
vlog() { [ "$VERBOSE" -eq 1 ] && log "debug" "$*"; log "info" "$*"; }

mkdir -p "$RAW_DIR" "$WEB_DIR"
touch "$LOG_FILE"

# flock returns the fd on the lock path; if another run holds it, exit cleanly.
exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  log "info" "another collect run is in progress; exiting"
  exit 0
fi

if [ ! -r "$SOURCES_CONF" ]; then
  log "error" "sources.conf not readable: $SOURCES_CONF"
  exit 1
fi

# ---- main loop: one source at a time ---------------------------------------
any_pulled=0
while read -r name user host log_dir; do
  case "$name" in ''|\#*) continue ;; esac
  if [ -z "$user" ] || [ -z "$host" ] || [ -z "$log_dir" ]; then
    log "error" "bad sources.conf entry for '$name' (want: name user host log_dir)"
    continue
  fi

  src_raw="$RAW_DIR/$name"
  mkdir -p "$src_raw"

  # list: TSV lines "date<TAB>size<TAB>mtime".
  # The forced command on the server is the only thing the key can run;
  # shell, scp, sftp, agent forwarding are all blocked at authorized_keys.
  list_out=$(ssh -o BatchMode=yes "$user@$host" list 2>>"$LOG_FILE") || {
    log "error" "$name: ssh list failed (see $LOG_FILE)"
    continue
  }
  vlog "$name: list returned $(echo "$list_out" | wc -l) day(s)"

  if [ "$DEBUG_LIST_ONLY" -eq 1 ]; then
    echo "=== $name ($user@$host) ==="
    echo "$list_out"
    continue
  fi

  while IFS=$'\t' read -r date size mtime; do
    case "$date" in ''|[!0-9]*) continue ;; esac
    cache="$src_raw/nhp_server_xdp-${date}.log.gz"
    need_pull=0
    if [ ! -f "$cache" ]; then
      need_pull=1
      reason="missing"
    elif [ "$date" = "$TODAY" ]; then
      # today's log is still being written: always re-pull so partial rows
      # disappear on the next timer tick.
      need_pull=1
      reason="today"
    else
      # Compare only size; the daemon's rotation-by-date writes to the new
      # file once a day, so a non-today file should not change size unless
      # somebody manually touched it.
      cached_size=$(stat -c %s "$cache" 2>/dev/null || echo 0)
      if [ "$size" -gt "$cached_size" ]; then
        need_pull=1
        reason="size $cached_size->$size"
      fi
    fi

    if [ "$need_pull" -eq 0 ]; then
      vlog "$name/$date: cached, skip"
      continue
    fi

    vlog "$name/$date: pulling ($reason, ${size} bytes)"
    tmp=$(mktemp -p "$src_raw" "tmp-nhp_server_xdp-${date}-XXXXXX.log")
    if ! ssh -o BatchMode=yes "$user@$host" "get $date" >"$tmp" 2>>"$LOG_FILE"; then
      log "error" "$name/$date: ssh get failed"
      rm -f "$tmp"
      continue
    fi
    # Reject empty files and files that do not end in a newline (which would
    # indicate the writer was mid-line when we cut the connection).
    if [ ! -s "$tmp" ]; then
      log "error" "$name/$date: empty get, dropping"
      rm -f "$tmp"
      continue
    fi
    last_byte=$(tail -c1 "$tmp" | od -An -c | tr -d ' ')
    if [ "$last_byte" != "\n" ]; then
      # The writer truncates on the next line; dropping the trailing partial
      # line keeps the analyzer honest. Truncate the file at the last
      # newline so the gzip'd cache is still useful.
      vlog "$name/$date: trailing partial line; trimming"
      head -n -1 "$tmp" >"${tmp}.trim" && mv "${tmp}.trim" "$tmp"
      if [ ! -s "$tmp" ]; then
        log "error" "$name/$date: only partial content, dropping"
        rm -f "$tmp"
        continue
      fi
    fi
    # Atomic gzip replace: write to <file>.tmp and rename. Use gzip -n so the
    # output is deterministic across rebuilds (no mtime embedded).
    gztmp="${cache}.tmp.$$"
    if ! gzip -n -c "$tmp" >"$gztmp"; then
      log "error" "$name/$date: gzip failed"
      rm -f "$tmp" "$gztmp"
      continue
    fi
    chmod 0640 "$gztmp"
    mv "$gztmp" "$cache"
    rm -f "$tmp"
    any_pulled=1
  done <<<"$list_out"
done <"$SOURCES_CONF"

if [ "$DEBUG_LIST_ONLY" -eq 1 ]; then
  exit 0
fi

# ---- rebuild site -----------------------------------------------------------
if [ -x "$SITE_BIN" ]; then
  log "info" "running xdplog-report site"
  if ! "$SITE_BIN" site -logs "$RAW_DIR" -out "$WEB_DIR" "${SITE_ARGS[@]}" >>"$LOG_FILE" 2>&1; then
    log "error" "xdplog-report site failed; old web content kept"
    exit 1
  fi
else
  log "error" "xdplog-report binary missing at $SITE_BIN; skipping site render"
  exit 1
fi

# ---- prune old raw cache ----------------------------------------------------
# Keep at most RETENTION_DAYS days of compressed logs; we use mtime because
# gzip preserves it. -mindepth 1 keeps the .lock and any future siblings.
pruned=$(find "$RAW_DIR" -mindepth 1 -maxdepth 2 -type f -name '*.log.gz' \
  -mtime "+${RETENTION_DAYS}" -delete -print | wc -l)
if [ "$pruned" -gt 0 ]; then
  log "info" "pruned $pruned raw cache file(s) older than $RETENTION_DAYS days"
fi

if [ "$any_pulled" -eq 1 ]; then
  log "info" "complete: pulled at least one day, site rewritten"
else
  log "info" "complete: no new days; idempotent path"
fi
exit 0