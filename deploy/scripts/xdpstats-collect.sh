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
# Three roots, deliberately in three places:
#
#   RAW_DIR  — the unmasked log cache (full source IPs). 0700, under $HOME,
#              and NOT reachable from the web root. This is the only copy of
#              unmasked data on the relay, so it is the thing the web root's
#              permissions must never expose.
#   WEB_DIR  — the published site. Lives under /var/www, not $HOME: on
#              AL2023 /home/ec2-user is mode 0700 and nginx workers run as
#              nginx, so a root under $HOME returns 403 to every request.
#   STATE    — the incremental-build state. Kept out of WEB_DIR on purpose:
#              state.json records the absolute path of every raw cache file,
#              which is host layout an anonymous visitor has no business
#              reading.
XDPSTATS_DIR="/home/ec2-user/xdpstats"
RAW_DIR="$XDPSTATS_DIR/raw"
WEB_DIR="/var/www/stats"
STATE_FILE="$XDPSTATS_DIR/state.json"
SOURCES_CONF="$XDPSTATS_DIR/sources.conf"
LOG_FILE="$XDPSTATS_DIR/collect.log"
LOCK_FILE="$RAW_DIR/.lock"
RETENTION_DAYS=90
SITE_BIN="/usr/local/bin/xdplog-report"
SITE_ARGS=("-top" "20" "-state" "$STATE_FILE")

# Identity used for every pull. It is a dedicated ed25519 key whose matching
# public key sits on the server behind the nhp-xdplog-export.sh forced
# command; it is NOT the deploy key, and it grants nothing else.
SSH_IDENTITY="/home/ec2-user/.ssh/stats_collector_key"
KNOWN_HOSTS="/home/ec2-user/.ssh/known_hosts"
SSH_OPTS=(
  -i "$SSH_IDENTITY"
  # Without IdentitiesOnly, ssh offers every identity in the agent and every
  # default key file first; the server would then see a rejected public-key
  # attempt before the collector's key, and with MaxAuthTries set low a
  # future extra key file in $HOME would lock the collector out.
  -o IdentitiesOnly=yes
  # No prompting: this runs from a timer. Trust comes from the pre-seeded
  # ssh_host_keys blob installed by deploy-xdpstats, never from TOFU, and a
  # changed server host key must be a visible failure, not a silent accept.
  -o StrictHostKeyChecking=yes
  -o UserKnownHostsFile="$KNOWN_HOSTS"
  -o BatchMode=yes
  -o ConnectTimeout=15
)
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
chmod 0700 "$RAW_DIR"
touch "$LOG_FILE"

# flock returns the fd on the lock path; if another run holds it, exit cleanly.
# Check for the binary first: without this, a missing flock is reported by
# the shell as "command not found", `! flock` turns that into "lock is held",
# and the run exits 0 having done nothing at all — a timer that looks green
# and never collects.
if ! command -v flock >/dev/null 2>&1; then
  log "error" "flock(1) not found (util-linux); refusing to run unlocked"
  exit 1
fi
exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  log "info" "another collect run is in progress; exiting"
  exit 0
fi

if [ ! -r "$SOURCES_CONF" ]; then
  log "error" "sources.conf not readable: $SOURCES_CONF"
  exit 1
fi

# Both files are installed by deploy-xdpstats on the same host as this
# script. Fail loudly rather than falling back: a missing identity means
# every pull is refused, and a missing known_hosts under
# StrictHostKeyChecking=yes means the same with a much more confusing error.
if [ ! -r "$SSH_IDENTITY" ]; then
  log "error" "collector SSH key missing or unreadable: $SSH_IDENTITY"
  exit 1
fi
if [ ! -r "$KNOWN_HOSTS" ]; then
  log "error" "known_hosts missing or unreadable: $KNOWN_HOSTS"
  log "error" "refusing to trust the server host key on first use"
  exit 1
fi

# ---- main loop: one source at a time ---------------------------------------
any_pulled=0
any_failed=0
while read -r name user host log_dir || [ -n "$name" ]; do
  # `|| [ -n "$name" ]` so a final line without a trailing newline still
  # parses: read returns non-zero on EOF-without-newline, and a hand-edited
  # sources.conf can easily be saved that way. Without this guard the
  # loop silently drops its last entry on every deploy.
  case "$name" in ''|\#*) continue ;; esac
  if [ -z "$user" ] || [ -z "$host" ] || [ -z "$log_dir" ]; then
    log "error" "bad sources.conf entry for '$name' (want: name user host log_dir)"
    any_failed=1
    continue
  fi

  src_raw="$RAW_DIR/$name"
  mkdir -p "$src_raw"
  chmod 0700 "$src_raw"

  # list: TSV lines "date<TAB>size<TAB>mtime".
  # The forced command on the server is the only thing the key can run;
  # shell, scp, sftp, agent forwarding are all blocked at authorized_keys.
  #
  # </dev/null is load-bearing, not decoration. This loop reads its source
  # list from a here-string on stdin, and ssh inherits that fd and drains it
  # before running the remote command: without the redirect the FIRST day is
  # pulled and every later line of `list` is silently swallowed, so a
  # 14-day retention window yields exactly one day's page and no error
  # anywhere. The outer `while read ... done <"$SOURCES_CONF"` has the same
  # hazard, which is why the redirect appears on both.
  list_out=$(ssh "${SSH_OPTS[@]}" "$user@$host" list </dev/null 2>>"$LOG_FILE") || {
    log "error" "$name: ssh list failed (see $LOG_FILE)"
    any_failed=1
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
    # Sidecar recording the (size, mtime) the server reported for this day
    # at the moment we pulled it.
    meta="$cache.meta"
    need_pull=0
    if [ ! -f "$cache" ] || [ ! -f "$meta" ]; then
      need_pull=1
      reason="missing"
    elif [ "$date" = "$TODAY" ]; then
      # today's log is still being written: always re-pull so partial rows
      # disappear on the next timer tick.
      need_pull=1
      reason="today"
    else
      # Compare against what the server reported when we last pulled, NOT
      # against the size of the local .gz: the remote figure is the
      # uncompressed log and the local one is gzip, so "remote > local" is
      # true for essentially every real log and would re-pull the whole
      # 14-day retention window every night, forever.
      cached_size=$(cut -f1 <"$meta" 2>/dev/null || echo "")
      cached_mtime=$(cut -f2 <"$meta" 2>/dev/null || echo "")
      if [ "$cached_size" != "$size" ] || [ "$cached_mtime" != "$mtime" ]; then
        need_pull=1
        reason="remote ${size}/${mtime} vs cached ${cached_size:-?}/${cached_mtime:-?}"
      fi
    fi

    if [ "$need_pull" -eq 0 ]; then
      vlog "$name/$date: cached, skip"
      continue
    fi

    vlog "$name/$date: pulling ($reason, ${size} bytes)"
    tmp=$(mktemp -p "$src_raw" "tmp-nhp_server_xdp-${date}-XXXXXX.log")
    if ! ssh "${SSH_OPTS[@]}" "$user@$host" "get $date" </dev/null >"$tmp" 2>>"$LOG_FILE"; then
      log "error" "$name/$date: ssh get failed"
      rm -f "$tmp"
      any_failed=1
      continue
    fi
    # Reject empty files and files that do not end in a newline (which would
    # indicate the writer was mid-line when we cut the connection).
    if [ ! -s "$tmp" ]; then
      log "error" "$name/$date: empty get, dropping"
      rm -f "$tmp"
      any_failed=1
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
        any_failed=1
        continue
      fi
    fi
    # Atomic gzip replace: write to <file>.tmp and rename. Use gzip -n so the
    # output is deterministic across rebuilds (no mtime embedded in the gzip
    # header). gzip itself does NOT preserve the source file's mtime on the
    # .gz file system entry: it sets the .gz's mtime to "now". If we let
    # that stand, find -mtime (the retention check below) measures from the
    # pull time, not the log date — and a -force re-pull pushes every
    # affected day's clock forward by however long the deploy took. Pin
    # the .gz's mtime to the source log's mtime (epoch seconds, reported
    # by `list`) so retention always counts from the day the log actually
    # ends.
    gztmp="${cache}.tmp.$$"
    if ! gzip -n -c "$tmp" >"$gztmp"; then
      log "error" "$name/$date: gzip failed"
      rm -f "$tmp" "$gztmp"
      any_failed=1
      continue
    fi
    chmod 0640 "$gztmp"
    if ! touch -d "@$mtime" "$gztmp"; then
      log "error" "$name/$date: touch -d @${mtime} on gzipped cache failed"
      rm -f "$tmp" "$gztmp"
      any_failed=1
      continue
    fi
    mv "$gztmp" "$cache"
    # Record what the server reported for this day BEFORE considering the
    # pull complete. If the process dies between the rename and here, the
    # next run sees a cache with no sidecar and re-pulls, which is right.
    printf '%s\t%s\n' "$size" "$mtime" >"${meta}.tmp.$$"
    chmod 0640 "${meta}.tmp.$$"
    # Match the .gz's mtime so the sidecar and its cache file age together
    # under find -mtime (otherwise the meta gets pruned first and we end up
    # re-pulling for no reason).
    if ! touch -d "@$mtime" "${meta}.tmp.$$"; then
      log "error" "$name/$date: touch -d @${mtime} on cache meta failed"
      rm -f "$tmp" "${meta}.tmp.$$"
      any_failed=1
      continue
    fi
    mv "${meta}.tmp.$$" "$meta"
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
# the pull path pins each .gz's mtime to the source log's mtime (see the
# note above the gzip step). That makes the retention clock count from the
# log date, not the pull time, and keeps -force rebuilds from pushing
# every affected day forward. -mindepth 1 keeps the .lock and any future
# siblings. Sidecars go with their log and are not counted: a day is one
# file, and the reported number is days, not deletions.
find "$RAW_DIR" -mindepth 1 -maxdepth 2 -type f -name '*.log.gz.meta' \
  -mtime "+${RETENTION_DAYS}" -delete >/dev/null
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

# A source that could not be listed means the cache is now older than the
# server's, and the site above has just been rebuilt from it and would look
# perfectly healthy. Exit non-zero so `systemctl status xdpstats` and the
# deploy summary go red: a collector that has quietly stopped reaching the
# server is exactly the failure nobody notices until the data is gone. The
# server keeps only 14 days of logs, so silence is not recoverable later.
if [ "$any_failed" -eq 1 ]; then
  log "error" "one or more sources failed; the published site may be stale"
  exit 1
fi
exit 0