#!/bin/bash
# nhp-xdplog-export.sh
#
# Forced-command SSH handler that exposes nhp-server's XDP event log files to
# the relay collector. Installed on the nhp-server host by the deploy-server
# job in .github/workflows/deploy-demo-v2.yml; wired into a stats-collector
# public key entry in /home/ec2-user/.ssh/authorized_keys as:
#
#   from="<relay private IP>",restrict,\
#     command="/home/ec2-user/nhp-server/etc/nhp-xdplog-export.sh" <pubkey> stats-opennhp-collector
#
# `restrict` enables every sshd restriction OpenSSH knows about (no-port-
# forwarding, no-X11-forwarding, no-agent-forwarding, no-pty, no-user-rc,
# ... and any future flags), so a leaked key from anywhere but the relay
# is useless at the line.
#
# Allowed commands:
#   list
#       Emit one TSV line per nhp_server_xdp-YYYY-MM-DD.log file under
#       $LOG_DIR, with date, size, mtime (epoch seconds).
#
#   get <YYYY-MM-DD>
#       Stream the matching nhp_server_xdp-<date>.log to stdout. Only the
#       main log is readable: the evaluate/audit siblings and any path
#       component are ignored by the date regex, so they cannot be fetched.
#
# Anything else returns 64 (EX_USAGE) and emits a single diagnostic line on
# stderr. This includes shell metacharacters, path separators, dates that do
# not match YYYY-MM-DD, days the daemon never wrote, and missing files.

LOG_DIR="/home/ec2-user/nhp-server/logs"
LOG_PREFIX="nhp_server_xdp"
DATE_RE='^[0-9]{4}-[0-9]{2}-[0-9]{2}$'

err() { echo "nhp-xdplog-export: $*" >&2; }

case "${SSH_ORIGINAL_COMMAND:-}" in
  "")
    err "no command given"
    exit 64
    ;;
  list)
    # Print one TSV line per main-log file, sorted by date ascending.
    # Filenames come straight from the directory; nothing else in this
    # directory starts with nhp_server_xdp- and ends in .log, so we do not
    # need a stricter filter. We additionally require the date regex so a
    # future addition (e.g. nhp_server_xdp-rotate-YYYY-MM-DD.log) is silently
    # dropped until somebody reviews this script.
    found=0
    if [ -d "$LOG_DIR" ]; then
      while IFS= read -r -d '' f; do
        base=$(basename "$f")
        date=${base#${LOG_PREFIX}-}
        date=${date%.log}
        if ! [[ "$date" =~ $DATE_RE ]]; then
          continue
        fi
        sz=$(stat -c '%s' "$f" 2>/dev/null) || continue
        mt=$(stat -c '%Y' "$f" 2>/dev/null) || continue
        printf '%s\t%s\t%s\n' "$date" "$sz" "$mt"
        found=1
      done < <(find "$LOG_DIR" -maxdepth 1 -type f -name "${LOG_PREFIX}-*.log" -print0 | sort -z)
    fi
    if [ "$found" -eq 0 ]; then
      err "no log files in $LOG_DIR"
      exit 1
    fi
    ;;
  get\ *)
    arg=${SSH_ORIGINAL_COMMAND#get }
    # No spaces, no path separators, strict date format.
    if [[ "$arg" =~ [[:space:]/] ]]; then
      err "invalid date '$arg' (no spaces, slashes, or path components)"
      exit 65
    fi
    if ! [[ "$arg" =~ $DATE_RE ]]; then
      err "invalid date '$arg' (want YYYY-MM-DD)"
      exit 65
    fi
    f="$LOG_DIR/${LOG_PREFIX}-${arg}.log"
    if [ ! -f "$f" ]; then
      err "no log for $arg"
      exit 66
    fi
    # Check the file is owned by the daemon's user (mode 0600), and not a
    # symlink: a malicious owner of $LOG_DIR could otherwise have pointed
    # us at any path. The log writer at nhp/log/logger.go:131 creates the file
    # with O_CREATE|O_APPEND and 0600 directly, never via symlink.
    if [ -L "$f" ]; then
      err "refusing symlink: $f"
      exit 73
    fi
    cat -- "$f"
    ;;
  *)
    err "unknown command (allowed: list, get YYYY-MM-DD)"
    exit 64
    ;;
esac