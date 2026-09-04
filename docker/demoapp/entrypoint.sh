#!/bin/sh
# Entrypoint for the demoapp container.
#
# Runs as root for one chown pass, then drops to the unprivileged
# `demoapp` user (uid 8081) for the actual process. The chown is the
# point: docker-compose bind-mounts the host's ./demoapp/etc and
# ./demoapp/data onto /app/etc and /app/data, which shadows the
# RUN chown in the Dockerfile — the bind-mounted paths end up owned
# by whatever the host checkout user is (e.g. f1 on macOS, a
# uid-numbered user on Linux). Without this pass, demoapp (uid
# 8081) cannot write the auto-generated sidecar into /app/etc or
# the SQLite WAL/SHM siblings into /app/data and the container
# crash-loops on Linux hosts. Docker Desktop hides the bug because
# virtiofs fakes ownership (review #1 of the demoapp review).
#
# gosu is used in preference to `su -c` because it does not spin up
# a session layer (no PAM, no shell fork), which keeps the binary
# as PID 1 in the container and preserves signal handling for
# `docker stop`.

set -eu

# Hand the bind-mounted write paths to demoapp. -R so WAL/SHM
# siblings created on first boot inherit the owner and the next
# restart does not find them re-owned by root.
chown -R demoapp:demoapp /app/data /app/etc /app/logs

# Drop to the unprivileged user and exec the binary as PID 1. Any
# CMD arguments from `docker run` / compose are appended.
exec gosu demoapp /usr/local/bin/demoapp -c /app/etc/config.toml "$@"
