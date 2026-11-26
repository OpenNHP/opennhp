#!/usr/bin/env bash
# Download firewall tools only on a qualified vendor kernel.
set -euo pipefail
export PATH=/usr/sbin:/usr/bin:/sbin:/bin

if [[ -f /proc/sys/net/mptcp/enabled ]]; then
  sysctl -w net.mptcp.enabled=0
fi
if ! command -v iptables >/dev/null 2>&1; then
  running_kernel="$(uname -r)"
  case "$running_kernel" in
    6.1.189-236.400.amzn2023.x86_64) expected_package="kernel-$running_kernel" ;;
    6.12.113-138.209.amzn2023.x86_64) expected_package="kernel6.12-$running_kernel" ;;
    *) echo "FATAL: unqualified kernel $running_kernel cannot download containment tools" >&2; exit 1 ;;
  esac
  if ! owner="$(rpm -qf "/boot/vmlinuz-$running_kernel" --qf '%{NAME}-%{VERSION}-%{RELEASE}.%{ARCH}')" || [[ "$owner" != "$expected_package" ]]; then
    echo 'FATAL: running kernel does not match its qualified vendor package' >&2
    exit 1
  fi
  dnf install -y iptables-nft
fi
